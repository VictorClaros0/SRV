package services

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ledongthuc/pdf"
	"srrv/internal/models"
)

// OCRReport holds metadata about a single OCR operation.
type OCRReport struct {
	Mode           string
	Confidence     float64
	RequiresReview bool
	Warnings       []string
	Anomalies      []string
}

// candidatePattern maps a candidate ID to keyword variants robust to OCR noise.
type candidatePattern struct {
	id       string
	keywords []string
}

// OCRRegion describes a fixed percentage-based crop over the rendered acta PNG.
// Values are fractions of image width/height in the [0,1] range.
type OCRRegion struct {
	Name   string
	X      float64
	Y      float64
	Width  float64
	Height float64
	Digits bool
}

type pdfImageOCRResult struct {
	Text         string
	Events       []string
	RegionActa   *models.RRVActa
	RegionReport OCRReport
}

// ─── package-level regexps & constants ───────────────────────────────────────

var (
	reActaID       = regexp.MustCompile(`\b(\d{13})\b`)
	reActaIDSpaced = regexp.MustCompile(`\b(\d{5})\s+(\d{5})\s+(\d{3})\b`)
	reDigitsOnly   = regexp.MustCompile(`^\d+$`)
	reVoteNum      = regexp.MustCompile(`\b\d{1,4}\b`)
	reMesaOCR      = regexp.MustCompile(`(?i)mesa\s*[n°o]?\s*[:\-\.]?\s*(\d{1,3})\b`)
)

var candidateNames = map[string]string{
	"P1": "Daenerys Targaryen",
	"P2": "Sansa Stark",
	"P3": "Robert Baratheon",
	"P4": "Tyrion Lannister",
}

// ocrCandidatePatterns lists keywords to search for each candidate in OCR text,
// ordered from most distinctive to least to minimise false matches.
var ocrCandidatePatterns = []candidatePattern{
	{"P1", []string{"Daenerys Targaryen", "Daenerys", "daenerys", "DAENERYS", "Daener"}},
	{"P2", []string{"Sansa Stark", "Sansa", "sansa", "SANSA"}},
	{"P3", []string{"Robert Baratheon", "Baratheon", "baratheon", "Robert", "robert", "ROBERT"}},
	{"P4", []string{"Tyrion Lannister", "Tyrion", "tyrion", "TYRION", "Lannister"}},
}

var regionOCRLayout = []OCRRegion{
	// Calibrated against 300 DPI pdftoppm output for the current landscape acta format.
	{Name: "codigo_acta", X: 0.046, Y: 0.165, Width: 0.226, Height: 0.082, Digits: true},
	{Name: "ubicacion", X: 0.128, Y: 0.165, Width: 0.244, Height: 0.141, Digits: false},
	{Name: "nro_mesa", X: 0.064, Y: 0.255, Width: 0.097, Height: 0.125, Digits: true},
	{Name: "P1", X: 0.331, Y: 0.296, Width: 0.056, Height: 0.033, Digits: true},
	{Name: "P2", X: 0.331, Y: 0.339, Width: 0.056, Height: 0.033, Digits: true},
	{Name: "P3", X: 0.331, Y: 0.382, Width: 0.056, Height: 0.033, Digits: true},
	{Name: "P4", X: 0.331, Y: 0.425, Width: 0.056, Height: 0.033, Digits: true},
	{Name: "votos_validos", X: 0.331, Y: 0.680, Width: 0.056, Height: 0.033, Digits: true},
	{Name: "votos_blancos", X: 0.331, Y: 0.741, Width: 0.056, Height: 0.033, Digits: true},
	{Name: "votos_nulos", X: 0.331, Y: 0.784, Width: 0.056, Height: 0.037, Digits: true},
	{Name: "electores_habilitados", X: 0.073, Y: 0.718, Width: 0.090, Height: 0.045, Digits: true},
	{Name: "papeletas_anfora", X: 0.073, Y: 0.827, Width: 0.090, Height: 0.045, Digits: true},
	{Name: "papeletas_no_usadas", X: 0.073, Y: 0.910, Width: 0.090, Height: 0.045, Digits: true},
}

// ─── HashBytes ────────────────────────────────────────────────────────────────

// HashBytes returns the SHA-256 hex digest of data.
func HashBytes(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// ─── ProcessActa — main entry point ──────────────────────────────────────────

// ProcessActa runs the three-layer OCR pipeline:
//  1. PDF_TEXT   — extract embedded text with ledongthuc/pdf
//  2. PDF_IMAGE_OCR — render page to PNG with pdftoppm, then Tesseract
//  3. SIMULATED_FALLBACK — honest stub, never presented as real OCR
func ProcessActa(fileBytes []byte, filename string) (*models.RRVActa, OCRReport, error) {
	if isPDFFile(fileBytes, filename) {
		// ── Layer 1: PDF text extraction ──────────────────────────────────────
		text, err := ExtractTextFromPDF(fileBytes)
		if err == nil && strings.TrimSpace(text) != "" {
			acta, report := ParseActaText(text)
			if hasMeaningfulOCRData(acta) {
				report.Anomalies = appendUnique(report.Anomalies, "PDF_TEXT_EXTRAIDO")
				report = ValidateOCRResult(acta, report)
				syncActaFromReport(acta, report)
				log.Printf("✅ OCR layer=PDF_TEXT acta=%s conf=%.2f", acta.ActaID, report.Confidence)
				return acta, report, nil
			}
			log.Printf("ℹ️  PDF_TEXT extraído pero sin datos suficientes; probando OCR visual")
		}

		// ── Layer 2: PDF image OCR via pdftoppm + Tesseract ───────────────────
		imageOCR, ocrErr := ExtractPDFImageOCR(fileBytes)
		ocrText := imageOCR.Text
		ocrEvents := imageOCR.Events
		if ocrErr == nil && strings.TrimSpace(ocrText) != "" {
			log.Printf("📄 Tesseract extrajo %d chars de texto", len(ocrText))
			acta, report := ParseVisualOCRText(ocrText)
			if regionOCRIsUsable(imageOCR.RegionActa) {
				regionActa := mergeRegionActa(acta, imageOCR.RegionActa)
				regionReport := mergeOCRReports(imageOCR.RegionReport, report)
				for _, ev := range ocrEvents {
					regionReport.Anomalies = appendUnique(regionReport.Anomalies, ev)
				}
				regionReport.Anomalies = appendUnique(regionReport.Anomalies, "PDF_TEXT_FALLIDO")
				regionReport = ValidateOCRResult(regionActa, regionReport)
				syncActaFromReport(regionActa, regionReport)
				log.Printf("✅ OCR layer=PDF_REGION_OCR acta=%s conf=%.2f", regionActa.ActaID, regionReport.Confidence)
				return regionActa, regionReport, nil
			}
			if acta != nil && acta.ActaID != "" {
				for _, ev := range ocrEvents {
					report.Anomalies = appendUnique(report.Anomalies, ev)
				}
				report.Anomalies = appendUnique(report.Anomalies, "PDF_TEXT_FALLIDO")
				report = ValidateOCRResult(acta, report)
				syncActaFromReport(acta, report)
				log.Printf("✅ OCR layer=PDF_IMAGE_OCR acta=%s conf=%.2f", acta.ActaID, report.Confidence)
				return acta, report, nil
			}
			ocrEvents = append(ocrEvents, "OCR_VISUAL_PARCIAL")
		} else if ocrErr == nil && regionOCRIsUsable(imageOCR.RegionActa) {
			regionReport := imageOCR.RegionReport
			for _, ev := range ocrEvents {
				regionReport.Anomalies = appendUnique(regionReport.Anomalies, ev)
			}
			regionReport.Anomalies = appendUnique(regionReport.Anomalies, "PDF_TEXT_FALLIDO")
			regionReport = ValidateOCRResult(imageOCR.RegionActa, regionReport)
			syncActaFromReport(imageOCR.RegionActa, regionReport)
			log.Printf("✅ OCR layer=PDF_REGION_OCR acta=%s conf=%.2f", imageOCR.RegionActa.ActaID, regionReport.Confidence)
			return imageOCR.RegionActa, regionReport, nil
		} else if ocrErr != nil {
			log.Printf("⚠️  PDF_IMAGE_OCR error: %v", ocrErr)
		}

		// ── Layer 3: Honest fallback ───────────────────────────────────────────
		acta, report := simulateOCRFallback(fileBytes)
		for _, ev := range ocrEvents {
			report.Anomalies = appendUnique(report.Anomalies, ev)
		}
		report.Anomalies = appendUnique(report.Anomalies, "PDF_TEXT_FALLIDO")
		if ocrErr != nil {
			report.Anomalies = appendUnique(report.Anomalies, "OCR_VISUAL_ERROR")
		} else {
			report.Anomalies = appendUnique(report.Anomalies, "PDF_SIN_TEXTO")
		}
		report.Warnings = append([]string{"PDF_TEXT y PDF_IMAGE_OCR fallaron; datos simulados"}, report.Warnings...)
		acta.Warnings = report.Warnings
		acta.Anomalies = report.Anomalies
		return acta, report, nil
	}

	// ── Image file: try OCR directly ──────────────────────────────────────────
	ocrText, ocrEvents, ocrErr := ExtractTextFromImageBytes(fileBytes, filename)
	if ocrErr == nil && strings.TrimSpace(ocrText) != "" {
		acta, report := ParseVisualOCRText(ocrText)
		if acta != nil && acta.ActaID != "" {
			for _, ev := range ocrEvents {
				report.Anomalies = appendUnique(report.Anomalies, ev)
			}
			report = ValidateOCRResult(acta, report)
			syncActaFromReport(acta, report)
			return acta, report, nil
		}
	}

	acta, report := simulateOCRFallback(fileBytes)
	return acta, report, nil
}

// ─── Layer 1: PDF text extraction ────────────────────────────────────────────

// ExtractTextFromPDF extracts plain text from a PDF byte slice using ledongthuc/pdf.
func ExtractTextFromPDF(fileBytes []byte) (string, error) {
	r, err := pdf.NewReader(bytes.NewReader(fileBytes), int64(len(fileBytes)))
	if err != nil {
		return "", fmt.Errorf("pdf.NewReader: %w", err)
	}
	plain, err := r.GetPlainText()
	if err != nil {
		return "", fmt.Errorf("GetPlainText: %w", err)
	}
	var buf bytes.Buffer
	buf.ReadFrom(plain)
	return buf.String(), nil
}

// ParseActaText parses clean PDF-embedded text into an RRVActa using sequential
// number group mapping. Suitable for well-structured PDF text.
func ParseActaText(text string) (*models.RRVActa, OCRReport) {
	report := OCRReport{
		Mode:       "PDF_TEXT",
		Confidence: 1.0,
		Warnings:   []string{},
		Anomalies:  []string{},
	}

	lines := splitLines(text)

	actaIdx := -1
	actaID := ""
	for i, line := range lines {
		if m := extractActaID(strings.TrimSpace(line)); m != "" {
			actaID = m
			actaIdx = i
			break
		}
	}

	if actaID == "" {
		report.Confidence = 0.0
		report.Warnings = append(report.Warnings, "No se encontró código de acta de 13 dígitos")
		report.Anomalies = append(report.Anomalies, "CAMPOS_CRITICOS_FALTANTES")
		return nil, report
	}

	acta := &models.RRVActa{
		ActaID:      actaID,
		Fuente:      "RRV",
		TipoEntrada: "IMAGEN",
		Candidatos:  []models.Candidato{},
		Warnings:    []string{},
		Anomalies:   []string{},
	}

	if len(actaID) == 13 {
		if n, err := strconv.Atoi(actaID[10:]); err == nil && n > 0 {
			acta.NroMesa = strconv.Itoa(n)
			acta.Mesa = acta.NroMesa
		}
	}

	extractLocation(lines, actaIdx, acta)
	numGroups := extractNumberGroups(lines[actaIdx+1:])
	mapNumbersToActa(acta, numGroups, &report)
	return acta, report
}

// ─── Layer 2: PDF image OCR ───────────────────────────────────────────────────

// ExtractTextFromPDFImageOCR converts the first PDF page to PNG with pdftoppm,
// then runs Tesseract OCR on it. Returns (text, event-anomaly-list, error).
func ExtractTextFromPDFImageOCR(fileBytes []byte) (string, []string, error) {
	result, err := ExtractPDFImageOCR(fileBytes)
	return result.Text, result.Events, err
}

// ExtractPDFImageOCR renders the first PDF page once, then runs both page-level
// OCR and fixed-region OCR against that image.
func ExtractPDFImageOCR(fileBytes []byte) (pdfImageOCRResult, error) {
	result := pdfImageOCRResult{Events: []string{}}
	events := []string{}

	tempDir, err := os.MkdirTemp("", "srrv2-ocr-*")
	if err != nil {
		return result, fmt.Errorf("MkdirTemp: %w", err)
	}
	defer os.RemoveAll(tempDir)

	pdfPath := filepath.Join(tempDir, "acta.pdf")
	if err := os.WriteFile(pdfPath, fileBytes, 0644); err != nil {
		return result, fmt.Errorf("write PDF: %w", err)
	}

	imgPath, err := RenderPDFToImage(pdfPath, filepath.Join(tempDir, "page"))
	if err != nil {
		result.Events = append(events, "OCR_RENDER_ERROR")
		return result, fmt.Errorf("pdftoppm: %w", err)
	}
	events = append(events, "OCR_IMAGE_RENDER_OK")

	text, err := RunTesseract(imgPath)
	if err != nil {
		events = append(events, "OCR_TESSERACT_ERROR")
	} else {
		events = append(events, "OCR_TESSERACT_OK")
		result.Text = text
	}

	regionActa, regionReport := RunRegionOCR(imgPath)
	if regionActa != nil && regionActa.ActaID != "" {
		result.RegionActa = regionActa
		result.RegionReport = regionReport
		events = append(events, "OCR_REGION_OK")
	} else {
		events = append(events, "OCR_REGION_PARCIAL")
		if len(regionReport.Warnings) > 0 {
			result.RegionReport = regionReport
		}
	}

	result.Events = events
	if result.Text == "" && result.RegionActa == nil {
		return result, fmt.Errorf("tesseract no produjo texto OCR aprovechable")
	}
	return result, nil
}

// RunRegionOCR crops fixed regions from the rendered acta image and runs
// specialised Tesseract passes for numeric and text areas.
func RunRegionOCR(imagePath string) (*models.RRVActa, OCRReport) {
	report := OCRReport{
		Mode:           "PDF_REGION_OCR",
		Confidence:     0.85,
		RequiresReview: true,
		Warnings:       []string{},
		Anomalies:      []string{"PDF_REGION_OCR_USADO"},
	}
	acta := &models.RRVActa{
		Fuente:      "RRV",
		TipoEntrada: "IMAGEN",
		Candidatos: []models.Candidato{
			{CandidatoID: "P1", Nombre: candidateNames["P1"]},
			{CandidatoID: "P2", Nombre: candidateNames["P2"]},
			{CandidatoID: "P3", Nombre: candidateNames["P3"]},
			{CandidatoID: "P4", Nombre: candidateNames["P4"]},
		},
		Warnings:  []string{},
		Anomalies: []string{},
	}

	for _, region := range regionOCRLayout {
		cropPath, err := CropImageRegion(imagePath, region)
		if err != nil {
			report.Warnings = append(report.Warnings, fmt.Sprintf("region %s no pudo recortarse: %v", region.Name, err))
			report.Confidence -= 0.03
			continue
		}

		if !region.Digits {
			text, err := runTesseractText(cropPath)
			if err != nil {
				report.Warnings = append(report.Warnings, "ubicacion no detectada por OCR regional")
				report.Confidence -= 0.03
				continue
			}
			applyRegionLocation(acta, text)
			continue
		}

		text, err := RunTesseractDigits(cropPath)
		if err != nil {
			report.Warnings = append(report.Warnings, fmt.Sprintf("%s no detectado por OCR regional", region.Name))
			report.Confidence -= 0.04
			continue
		}

		if region.Name == "codigo_acta" {
			actaID := extractActaID(text)
			if actaID == "" {
				digits := onlyDigits(text)
				if len(digits) == 13 {
					actaID = digits
				}
			}
			if actaID == "" {
				report.Warnings = append(report.Warnings, "codigo_acta no detectado por OCR regional")
				report.Anomalies = appendUnique(report.Anomalies, "CAMPOS_CRITICOS_FALTANTES")
				report.Confidence -= 0.20
				continue
			}
			acta.ActaID = actaID
			if n, err := strconv.Atoi(actaID[10:]); err == nil && n > 0 {
				acta.NroMesa = strconv.Itoa(n)
				acta.Mesa = acta.NroMesa
			}
			continue
		}

		value, ok := ParseRegionNumber(text)
		if !ok {
			report.Warnings = append(report.Warnings, fmt.Sprintf("%s no detectado por OCR regional", region.Name))
			report.Confidence -= 0.04
			continue
		}
		applyRegionNumber(acta, region.Name, value, &report)
	}

	if acta.ActaID == "" {
		report.Anomalies = appendUnique(report.Anomalies, "CAMPOS_CRITICOS_FALTANTES")
		report.Anomalies = appendUnique(report.Anomalies, "OCR_REGION_PARCIAL")
		report.RequiresReview = true
		return nil, report
	}

	if acta.VotosValidos == 0 {
		candidateSum := 0
		for _, c := range acta.Candidatos {
			candidateSum += c.Votos
		}
		if candidateSum > 0 {
			acta.VotosValidos = candidateSum
			report.Warnings = append(report.Warnings, "votos_validos derivado de votos regionales por candidatura")
			report.Anomalies = appendUnique(report.Anomalies, "OCR_REGION_DERIVADO")
			report.RequiresReview = true
		}
	}
	if acta.TotalVotos == 0 {
		acta.TotalVotos = acta.VotosValidos + acta.VotosBlancos + acta.VotosNulos
	}
	if regionNumericFieldCount(acta) < 3 {
		report.Anomalies = appendUnique(report.Anomalies, "OCR_REGION_PARCIAL")
		report.RequiresReview = true
		report.Confidence -= 0.15
	}
	return acta, report
}

// CropImageRegion writes a cropped PNG for a configured percentage region.
// Numeric regions are upscaled and binarised to help Tesseract focus on digits.
func CropImageRegion(imagePath string, region OCRRegion) (string, error) {
	f, err := os.Open(imagePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return "", err
	}
	b := img.Bounds()
	rect := image.Rect(
		b.Min.X+int(region.X*float64(b.Dx())),
		b.Min.Y+int(region.Y*float64(b.Dy())),
		b.Min.X+int((region.X+region.Width)*float64(b.Dx())),
		b.Min.Y+int((region.Y+region.Height)*float64(b.Dy())),
	).Intersect(b)
	if rect.Empty() {
		return "", fmt.Errorf("region vacia")
	}

	rgba := image.NewRGBA(image.Rect(0, 0, rect.Dx(), rect.Dy()))
	draw.Draw(rgba, rgba.Bounds(), img, rect.Min, draw.Src)
	out := image.Image(rgba)

	outPath := filepath.Join(filepath.Dir(imagePath), "region_"+region.Name+".png")
	outFile, err := os.Create(outPath)
	if err != nil {
		return "", err
	}
	defer outFile.Close()
	if err := png.Encode(outFile, out); err != nil {
		return "", err
	}
	return outPath, nil
}

// RunTesseractDigits runs Tesseract with numeric whitelist for one cropped region.
func RunTesseractDigits(regionPath string) (string, error) {
	if _, err := exec.LookPath("tesseract"); err != nil {
		return "", fmt.Errorf("tesseract no encontrado en PATH: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "tesseract",
		regionPath, "stdout",
		"--psm", "7",
		"-c", "tessedit_char_whitelist=0123456789")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	text := strings.TrimSpace(string(out))
	if text == "" {
		return "", fmt.Errorf("region sin digitos")
	}
	return text, nil
}

// ParseRegionNumber extracts a non-negative integer from Tesseract output.
func ParseRegionNumber(text string) (int, bool) {
	digits := onlyDigits(text)
	if digits == "" || len(digits) > 4 {
		return 0, false
	}
	n, err := strconv.Atoi(digits)
	return n, err == nil
}

// ExtractTextFromImageBytes writes an image file to a temp dir and runs Tesseract on it.
func ExtractTextFromImageBytes(fileBytes []byte, filename string) (string, []string, error) {
	events := []string{}

	tempDir, err := os.MkdirTemp("", "srrv2-ocr-*")
	if err != nil {
		return "", events, err
	}
	defer os.RemoveAll(tempDir)

	ext := ".png"
	low := strings.ToLower(filename)
	if strings.HasSuffix(low, ".jpg") || strings.HasSuffix(low, ".jpeg") {
		ext = ".jpg"
	} else if strings.HasSuffix(low, ".tiff") || strings.HasSuffix(low, ".tif") {
		ext = ".tiff"
	}
	imgPath := filepath.Join(tempDir, "image"+ext)
	if err := os.WriteFile(imgPath, fileBytes, 0644); err != nil {
		return "", events, err
	}

	text, err := RunTesseract(imgPath)
	if err != nil {
		return "", append(events, "OCR_TESSERACT_ERROR"), err
	}
	return text, append(events, "OCR_TESSERACT_OK"), nil
}

// RenderPDFToImage converts the first page of a PDF to a PNG using pdftoppm.
// outputPrefix should be a path like "/tmp/srrv2-abc/page".
// Returns the path of the generated image file.
func RenderPDFToImage(pdfPath, outputPrefix string) (string, error) {
	if _, err := exec.LookPath("pdftoppm"); err != nil {
		return "", fmt.Errorf("pdftoppm no encontrado en PATH: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "pdftoppm",
		"-png", "-singlefile", "-r", "300",
		"-f", "1", "-l", "1",
		pdfPath, outputPrefix)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("pdftoppm: %w — %s", err, strings.TrimSpace(string(out)))
	}

	// pdftoppm -singlefile writes outputPrefix.png
	for _, candidate := range []string{
		outputPrefix + ".png",
		outputPrefix + "-1.png",
		outputPrefix + "-01.png",
	} {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("imagen generada no encontrada en %s", outputPrefix)
}

// RunTesseract runs Tesseract on an image file, trying spa+eng → eng → no-lang fallbacks.
func RunTesseract(imagePath string) (string, error) {
	if _, err := exec.LookPath("tesseract"); err != nil {
		return "", fmt.Errorf("tesseract no encontrado en PATH: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	langCombos := [][]string{
		{"-l", "spa+eng"},
		{"-l", "eng"},
		{}, // no language flag
	}
	for _, langArgs := range langCombos {
		args := []string{imagePath, "stdout"}
		args = append(args, langArgs...)
		args = append(args, "--psm", "6")
		cmd := exec.CommandContext(ctx, "tesseract", args...)
		out, err := cmd.Output()
		if err == nil && strings.TrimSpace(string(out)) != "" {
			return string(out), nil
		}
	}
	return "", fmt.Errorf("tesseract no produjo texto")
}

// ─── Layer 2: tolerant OCR text parser ───────────────────────────────────────

// ParseVisualOCRText parses noisy Tesseract output using label-based extraction
// followed by sequential number mapping as fallback.
func ParseVisualOCRText(rawText string) (*models.RRVActa, OCRReport) {
	report := OCRReport{
		Mode:       "PDF_IMAGE_OCR",
		Confidence: 1.0,
		Warnings:   []string{},
		Anomalies:  []string{"PDF_IMAGE_OCR_USADO"},
	}

	text := PreprocessOCRText(rawText)

	// Find acta_id (13-digit sequence)
	actaID := ""
	if m := extractActaID(text); m != "" {
		actaID = m
	}
	if actaID == "" {
		report.Confidence = 0.2
		report.Warnings = append(report.Warnings, "acta_id no detectado por OCR visual")
		report.Anomalies = appendUnique(report.Anomalies, "CAMPOS_CRITICOS_FALTANTES")
		report.Anomalies = appendUnique(report.Anomalies, "OCR_VISUAL_PARCIAL")
		return nil, report
	}

	acta := &models.RRVActa{
		ActaID:      actaID,
		Fuente:      "RRV",
		TipoEntrada: "IMAGEN",
		Candidatos:  []models.Candidato{},
		Warnings:    []string{},
		Anomalies:   []string{},
	}

	// Derive nro_mesa from acta_id suffix (last 3 digits, strip leading zeros)
	if len(actaID) == 13 {
		if n, err := strconv.Atoi(actaID[10:]); err == nil && n > 0 {
			acta.NroMesa = strconv.Itoa(n)
			acta.Mesa = acta.NroMesa
		}
	}

	// Try to refine nro_mesa from OCR label
	if m := reMesaOCR.FindStringSubmatch(text); len(m) > 1 {
		if n, err := strconv.Atoi(m[1]); err == nil && n > 0 {
			acta.NroMesa = strconv.Itoa(n)
			acta.Mesa = acta.NroMesa
		}
	}

	// Extract location from lines before the acta_id line
	lines := splitLines(text)
	actaLineIdx := -1
	for i, l := range lines {
		if lineContainsActaID(l, actaID) {
			actaLineIdx = i
			break
		}
	}
	if actaLineIdx > 0 {
		extractLocation(lines, actaLineIdx, acta)
	}

	// ── Label-based extraction (most reliable for OCR text) ───────────────────
	labelFieldsFound := 0

	// Candidate votes by name
	for _, cp := range ocrCandidatePatterns {
		votos, ok := extractFirstNumberAfterKeyword(text, cp.keywords)
		acta.Candidatos = append(acta.Candidatos, models.Candidato{
			CandidatoID: cp.id,
			Nombre:      candidateNames[cp.id],
			Votos:       votos,
		})
		if ok && votos > 0 {
			labelFieldsFound++
		} else {
			report.Confidence -= 0.05
			report.Warnings = append(report.Warnings,
				fmt.Sprintf("Votos %s no detectados por OCR", cp.id))
		}
	}

	// Votos válidos
	if v, ok := extractFirstNumberAfterKeyword(text,
		[]string{"álidos", "alidos", "ALIDOS", "Válidos", "Validos"}); ok && v > 0 {
		acta.VotosValidos = v
		labelFieldsFound++
	} else {
		report.Confidence -= 0.2
		report.Warnings = append(report.Warnings, "votos_validos no detectado por OCR")
		report.Anomalies = appendUnique(report.Anomalies, "CAMPOS_CRITICOS_FALTANTES")
	}

	// Votos blancos
	if v, ok := extractFirstNumberAfterKeyword(text,
		[]string{"lancos", "Blancos", "blancos", "BLANCOS"}); ok && v >= 0 {
		acta.VotosBlancos = v
		labelFieldsFound++
	} else {
		report.Confidence -= 0.2
		report.Warnings = append(report.Warnings, "votos_blancos no detectado por OCR")
		report.Anomalies = appendUnique(report.Anomalies, "CAMPOS_CRITICOS_FALTANTES")
	}

	// Votos nulos
	if v, ok := extractFirstNumberAfterKeyword(text,
		[]string{"Nulos", "nulos", "NULOS", "ulos"}); ok && v >= 0 {
		acta.VotosNulos = v
		labelFieldsFound++
	} else {
		report.Confidence -= 0.2
		report.Warnings = append(report.Warnings, "votos_nulos no detectado por OCR")
		report.Anomalies = appendUnique(report.Anomalies, "CAMPOS_CRITICOS_FALTANTES")
	}

	// Electores habilitados
	if v, ok := extractFirstNumberAfterKeyword(text,
		[]string{"habilitados", "Habilitados", "HABILITADOS", "inscritos", "Inscritos", "electores"}); ok && v > 0 {
		acta.ElectoresHabilitados = v
	} else {
		report.Confidence -= 0.1
	}

	// Papeletas en ánfora
	if v, ok := extractFirstNumberAfterKeyword(text,
		[]string{"nfora", "ánfora", "Anfora", "anfora"}); ok && v > 0 {
		acta.PapeletasAnfora = v
	}

	// Papeletas no usadas
	if v, ok := extractFirstNumberAfterKeyword(text,
		[]string{"no utilizadas", "No utilizadas", "no usadas", "utilizadas"}); ok && v > 0 {
		acta.PapeletasNoUsadas = v
	}

	// ── Sequential fallback if label-based got very little ────────────────────
	if labelFieldsFound < 3 {
		report.Anomalies = appendUnique(report.Anomalies, "OCR_VISUAL_PARCIAL")
		// Try the sequential parser on the preprocessed text as a secondary attempt
		seqActa, _ := ParseActaText(text)
		if seqActa != nil && hasMeaningfulOCRData(seqActa) {
			if acta.VotosValidos == 0 && seqActa.VotosValidos > 0 {
				acta.VotosValidos = seqActa.VotosValidos
			}
			if acta.VotosNulos == 0 && seqActa.VotosNulos > 0 {
				acta.VotosNulos = seqActa.VotosNulos
			}
			if acta.VotosBlancos == 0 && seqActa.VotosBlancos > 0 {
				acta.VotosBlancos = seqActa.VotosBlancos
			}
			if acta.ElectoresHabilitados == 0 && seqActa.ElectoresHabilitados > 0 {
				acta.ElectoresHabilitados = seqActa.ElectoresHabilitados
			}
			// Merge candidate votes if ours are all zero
			allZero := true
			for _, c := range acta.Candidatos {
				if c.Votos > 0 {
					allZero = false
					break
				}
			}
			if allZero && len(seqActa.Candidatos) > 0 {
				acta.Candidatos = seqActa.Candidatos
			}
		}
	}

	// Derive TotalVotos
	if acta.PapeletasAnfora > 0 {
		acta.TotalVotos = acta.PapeletasAnfora
	} else {
		acta.TotalVotos = acta.VotosValidos + acta.VotosBlancos + acta.VotosNulos
	}

	return acta, report
}

// PreprocessOCRText normalises OCR noise: fixes common digit confusions (O→0,
// l→1 etc.) in tokens that are >60% digit-like, and normalises whitespace.
func PreprocessOCRText(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	lines := strings.Split(text, "\n")
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		tokens := strings.Fields(line)
		for i, tok := range tokens {
			tokens[i] = fixOCRDigits(tok)
		}
		result = append(result, strings.Join(tokens, " "))
	}
	return strings.Join(result, "\n")
}

// fixOCRDigits corrects common OCR digit confusions (O→0, I/l→1, Z→2) in a
// single token, but only when the token looks like it should be a number.
func fixOCRDigits(s string) string {
	if len(s) == 0 || len([]rune(s)) > 13 {
		return s // skip long tokens (words, full acta_id)
	}
	digitLike := 0
	for _, ch := range s {
		if (ch >= '0' && ch <= '9') || ch == 'O' || ch == 'o' ||
			ch == 'I' || ch == 'l' || ch == '|' || ch == 'Z' {
			digitLike++
		}
	}
	runes := []rune(s)
	if float64(digitLike)/float64(len(runes)) < 0.6 {
		return s
	}
	var b strings.Builder
	for _, ch := range runes {
		switch ch {
		case 'O', 'o':
			b.WriteRune('0')
		case 'I', 'l', '|':
			b.WriteRune('1')
		case 'Z':
			b.WriteRune('2')
		default:
			b.WriteRune(ch)
		}
	}
	cleaned := b.String()
	if reDigitsOnly.MatchString(cleaned) {
		return cleaned
	}
	return s
}

// extractFirstNumberAfterKeyword searches for any keyword in text and returns
// the first 1-4 digit number found on the same line after the keyword, or on
// the next non-empty line.
func extractFirstNumberAfterKeyword(text string, keywords []string) (int, bool) {
	lines := strings.Split(text, "\n")
	for _, kw := range keywords {
		kwLower := strings.ToLower(kw)
		for lineIdx, line := range lines {
			if !strings.Contains(strings.ToLower(line), kwLower) {
				continue
			}
			// Search same line after keyword
			kwPos := strings.Index(strings.ToLower(line), kwLower)
			afterKW := line[kwPos+len(kw):]
			if m := reVoteNum.FindString(afterKW); m != "" {
				if n, err := strconv.Atoi(m); err == nil {
					return n, true
				}
			}
			// Search next non-empty line (up to 2 lookahead)
			for j := lineIdx + 1; j <= lineIdx+2 && j < len(lines); j++ {
				nextLine := strings.TrimSpace(lines[j])
				if nextLine == "" {
					continue
				}
				if m := reVoteNum.FindString(nextLine); m != "" {
					if n, err := strconv.Atoi(m); err == nil {
						return n, true
					}
				}
				break
			}
		}
	}
	return 0, false
}

// hasMeaningfulOCRData returns true if the acta contains at least one numeric
// field beyond the acta_id itself, indicating the sequential parser found data.
func hasMeaningfulOCRData(acta *models.RRVActa) bool {
	if acta == nil || acta.ActaID == "" {
		return false
	}
	if acta.VotosValidos > 0 || acta.TotalVotos > 0 || acta.ElectoresHabilitados > 0 {
		return true
	}
	nonZero := 0
	for _, c := range acta.Candidatos {
		if c.Votos > 0 {
			nonZero++
		}
	}
	return nonZero >= 2
}

func regionOCRIsUsable(acta *models.RRVActa) bool {
	return acta != nil && acta.ActaID != "" && regionNumericFieldCount(acta) >= 3
}

func regionNumericFieldCount(acta *models.RRVActa) int {
	if acta == nil {
		return 0
	}
	count := 0
	for _, v := range []int{
		acta.ElectoresHabilitados,
		acta.PapeletasAnfora,
		acta.PapeletasNoUsadas,
		acta.VotosValidos,
		acta.VotosBlancos,
		acta.VotosNulos,
	} {
		if v > 0 {
			count++
		}
	}
	for _, c := range acta.Candidatos {
		if c.Votos > 0 {
			count++
		}
	}
	return count
}

func mergeRegionActa(pageActa, regionActa *models.RRVActa) *models.RRVActa {
	if pageActa == nil {
		return regionActa
	}
	if regionActa == nil {
		return pageActa
	}
	merged := *pageActa
	merged.OCRMode = "PDF_REGION_OCR"

	if regionActa.ActaID != "" {
		merged.ActaID = regionActa.ActaID
	}
	if regionActa.NroMesa != "" {
		merged.NroMesa = regionActa.NroMesa
		merged.Mesa = regionActa.NroMesa
	}
	if regionActa.Departamento != "" {
		merged.Departamento = regionActa.Departamento
	}
	if regionActa.Provincia != "" {
		merged.Provincia = regionActa.Provincia
	}
	if regionActa.Municipio != "" {
		merged.Municipio = regionActa.Municipio
	}
	if regionActa.Recinto != "" {
		merged.Recinto = regionActa.Recinto
	}
	if regionActa.ElectoresHabilitados > 0 {
		merged.ElectoresHabilitados = regionActa.ElectoresHabilitados
	}
	if regionActa.PapeletasAnfora > 0 {
		merged.PapeletasAnfora = regionActa.PapeletasAnfora
	}
	if regionActa.PapeletasNoUsadas > 0 {
		merged.PapeletasNoUsadas = regionActa.PapeletasNoUsadas
	}
	if regionActa.VotosValidos > 0 {
		merged.VotosValidos = regionActa.VotosValidos
	}
	if regionActa.VotosBlancos > 0 {
		merged.VotosBlancos = regionActa.VotosBlancos
	}
	if regionActa.VotosNulos > 0 {
		merged.VotosNulos = regionActa.VotosNulos
	}
	if regionActa.TotalVotos > 0 {
		merged.TotalVotos = regionActa.TotalVotos
	}

	byID := make(map[string]models.Candidato, len(merged.Candidatos))
	for _, c := range merged.Candidatos {
		byID[c.CandidatoID] = c
	}
	for _, rc := range regionActa.Candidatos {
		if rc.Votos <= 0 {
			continue
		}
		if rc.Nombre == "" {
			rc.Nombre = candidateNames[rc.CandidatoID]
		}
		byID[rc.CandidatoID] = rc
	}
	merged.Candidatos = make([]models.Candidato, 0, len(byID))
	for _, id := range []string{"P1", "P2", "P3", "P4"} {
		if c, ok := byID[id]; ok {
			merged.Candidatos = append(merged.Candidatos, c)
		}
	}
	if merged.PapeletasAnfora > 0 {
		merged.TotalVotos = merged.PapeletasAnfora
	} else {
		merged.TotalVotos = merged.VotosValidos + merged.VotosBlancos + merged.VotosNulos
	}
	return &merged
}

func mergeOCRReports(regionReport, pageReport OCRReport) OCRReport {
	report := regionReport
	report.Mode = "PDF_REGION_OCR"
	report.RequiresReview = true
	report.Warnings = append(report.Warnings, pageReport.Warnings...)
	for _, a := range pageReport.Anomalies {
		report.Anomalies = appendUnique(report.Anomalies, a)
	}
	return report
}

// ─── Validation ───────────────────────────────────────────────────────────────

// ValidateOCRResult runs all arithmetic and completeness checks,
// updating the report and setting acta.Estado.
func ValidateOCRResult(acta *models.RRVActa, report OCRReport) OCRReport {
	if acta.ActaID == "" {
		report.Confidence -= 0.2
		report.Anomalies = appendUnique(report.Anomalies, "CAMPOS_CRITICOS_FALTANTES")
		report.Warnings = append(report.Warnings, "acta_id faltante")
	}
	if acta.NroMesa == "" {
		report.Confidence -= 0.2
		report.Anomalies = appendUnique(report.Anomalies, "CAMPOS_CRITICOS_FALTANTES")
		report.Warnings = append(report.Warnings, "nro_mesa faltante")
	}

	sumaCand := 0
	for _, c := range acta.Candidatos {
		sumaCand += c.Votos
	}
	if acta.VotosValidos > 0 && sumaCand != acta.VotosValidos {
		report.Confidence -= 0.3
		report.Anomalies = appendUnique(report.Anomalies, "INCONSISTENCIA_ARITMETICA")
		report.Warnings = append(report.Warnings,
			fmt.Sprintf("suma candidatos=%d ≠ votos_validos=%d", sumaCand, acta.VotosValidos))
	}

	sumaTotal := acta.VotosValidos + acta.VotosBlancos + acta.VotosNulos
	if acta.TotalVotos > 0 && sumaTotal != acta.TotalVotos {
		report.Confidence -= 0.1
		report.Anomalies = appendUnique(report.Anomalies, "INCONSISTENCIA_ARITMETICA")
		report.Warnings = append(report.Warnings,
			fmt.Sprintf("válidos+blancos+nulos=%d ≠ total_votos=%d", sumaTotal, acta.TotalVotos))
	} else if acta.TotalVotos == 0 {
		acta.TotalVotos = sumaTotal
	}

	if acta.ElectoresHabilitados > 0 && acta.TotalVotos > acta.ElectoresHabilitados {
		report.Confidence -= 0.2
		report.Anomalies = appendUnique(report.Anomalies, "ELECTORES_MENOR_TOTAL_VOTOS")
		report.Warnings = append(report.Warnings,
			fmt.Sprintf("total_votos=%d > electores_habilitados=%d",
				acta.TotalVotos, acta.ElectoresHabilitados))
	}

	if acta.PapeletasAnfora > 0 && acta.PapeletasNoUsadas > 0 && acta.ElectoresHabilitados > 0 {
		diff := (acta.PapeletasAnfora + acta.PapeletasNoUsadas) - acta.ElectoresHabilitados
		if diff < 0 {
			diff = -diff
		}
		if diff > 5 {
			report.Anomalies = appendUnique(report.Anomalies, "PAPELETAS_NO_CUADRAN")
			report.Warnings = append(report.Warnings,
				fmt.Sprintf("papeletas anfora+no_usadas=%d ≠ electores=%d",
					acta.PapeletasAnfora+acta.PapeletasNoUsadas, acta.ElectoresHabilitados))
		}
	}

	if report.Confidence < 0 {
		report.Confidence = 0
	}
	if report.Confidence > 1 {
		report.Confidence = 1
	}

	hasMissing := containsStr(report.Anomalies, "CAMPOS_CRITICOS_FALTANTES")
	hasArith := containsStr(report.Anomalies, "INCONSISTENCIA_ARITMETICA")

	switch {
	case hasArith:
		acta.Estado = "INCONSISTENTE"
		report.RequiresReview = true
	case hasMissing || report.Confidence < 0.75:
		acta.Estado = "REQUIERE_REVISION"
		report.RequiresReview = true
		report.Anomalies = appendUnique(report.Anomalies, "REQUIERE_REVISION_MANUAL")
	default:
		acta.Estado = "PROCESADA"
		report.RequiresReview = false
	}
	return report
}

// ─── Honest fallback simulation ───────────────────────────────────────────────

// simulateOCRFallback generates deterministic fake data explicitly labeled as
// simulation. It must never be presented as real OCR output.
func simulateOCRFallback(fileBytes []byte) (*models.RRVActa, OCRReport) {
	hash := HashBytes(fileBytes)
	seed, _ := hex.DecodeString(hash[:2])
	base := int(seed[0])

	c1 := 100 + base*2
	c2 := 80 + base
	c3 := 40 + base/2
	c4 := base / 3
	nulos := 10 + base/10
	blancos := 5 + base/20
	validos := c1 + c2 + c3 + c4
	total := validos + nulos + blancos

	warnings := []string{
		"Se usó fallback simulado; no es OCR real",
		"Los datos de esta acta son ficticios y no provienen del documento",
	}
	anomalies := []string{"OCR_SIMULADO_USADO", "REQUIERE_REVISION_MANUAL"}

	acta := &models.RRVActa{
		ActaID:       fmt.Sprintf("SIM-%s", hash[:8]),
		Departamento: "(Simulado)",
		Provincia:    "(Simulado)",
		Municipio:    "(Simulado)",
		Recinto:      "(Simulado)",
		Mesa:         "SIM",
		NroMesa:      "SIM",
		Candidatos: []models.Candidato{
			{CandidatoID: "P1", Nombre: candidateNames["P1"], Votos: c1},
			{CandidatoID: "P2", Nombre: candidateNames["P2"], Votos: c2},
			{CandidatoID: "P3", Nombre: candidateNames["P3"], Votos: c3},
			{CandidatoID: "P4", Nombre: candidateNames["P4"], Votos: c4},
		},
		VotosValidos:   validos,
		VotosNulos:     nulos,
		VotosBlancos:   blancos,
		TotalVotos:     total,
		Estado:         "REQUIERE_REVISION",
		Fuente:         "RRV",
		TipoEntrada:    "IMAGEN",
		OCRMode:        "SIMULATED_FALLBACK",
		Confidence:     0.30,
		RequiresReview: true,
		Warnings:       warnings,
		Anomalies:      anomalies,
	}
	report := OCRReport{
		Mode:           "SIMULATED_FALLBACK",
		Confidence:     0.30,
		RequiresReview: true,
		Warnings:       warnings,
		Anomalies:      anomalies,
	}
	return acta, report
}

// ─── Shared parsing helpers ───────────────────────────────────────────────────

func extractLocation(lines []string, actaIdx int, acta *models.RRVActa) {
	var loc []string
	for i := 0; i < actaIdx; i++ {
		l := strings.TrimSpace(lines[i])
		if l == "" || reDigitsOnly.MatchString(l) {
			continue
		}
		loc = append(loc, l)
	}
	n := len(loc)
	switch {
	case n >= 4:
		acta.Departamento = loc[n-4]
		acta.Provincia = loc[n-3]
		acta.Municipio = loc[n-2]
		acta.Recinto = loc[n-1]
	case n == 3:
		acta.Departamento = loc[0]
		acta.Provincia = loc[1]
		acta.Municipio = loc[2]
	case n == 2:
		acta.Departamento = loc[0]
		acta.Municipio = loc[1]
	case n == 1:
		acta.Departamento = loc[0]
	}
}

func extractNumberGroups(lines []string) []string {
	var groups []string
	for _, line := range lines {
		if n := normalizeNumber(strings.TrimSpace(line)); n != "" {
			groups = append(groups, n)
		}
	}
	return groups
}

func normalizeNumber(s string) string {
	s = strings.TrimSpace(s)
	if reDigitsOnly.MatchString(s) {
		return s
	}
	parts := strings.Fields(s)
	if len(parts) < 2 {
		return ""
	}
	for _, p := range parts {
		if len(p) != 1 || p[0] < '0' || p[0] > '9' {
			return ""
		}
	}
	return strings.Join(parts, "")
}

func looksLikeTime(s string) bool {
	if len(s) != 4 {
		return false
	}
	h, e1 := strconv.Atoi(s[:2])
	m, e2 := strconv.Atoi(s[2:])
	return e1 == nil && e2 == nil && h >= 0 && h <= 23 && m >= 0 && m <= 59
}

// mapNumbersToActa maps sequential numeric groups to acta fields.
// Expected order after the first time-like (HHMM) group:
//
//	[0] apertura, [1] cierre, [2] electores, [3] papAnfora,
//	[4] papNoUsadas, [5-8] P1-P4, [9] votos_validos,
//	[10] votos_blancos, [11] votos_nulos
func mapNumbersToActa(acta *models.RRVActa, groups []string, report *OCRReport) {
	startIdx := 0
	for i, g := range groups {
		if looksLikeTime(g) {
			startIdx = i
			break
		}
	}

	for i := 0; i < startIdx; i++ {
		if n, err := strconv.Atoi(groups[i]); err == nil && n > 0 && n < 1000 {
			acta.NroMesa = strconv.Itoa(n)
			acta.Mesa = acta.NroMesa
		}
	}

	get := func(offset int) (int, bool) {
		i := startIdx + offset
		if i >= len(groups) {
			return 0, false
		}
		v, err := strconv.Atoi(groups[i])
		return v, err == nil
	}
	getRaw := func(offset int) string {
		i := startIdx + offset
		if i >= len(groups) {
			return ""
		}
		return groups[i]
	}

	if raw := getRaw(0); len(raw) == 4 {
		acta.AperturaHora = raw[:2] + ":" + raw[2:]
	}
	if raw := getRaw(1); len(raw) == 4 {
		acta.CierreHora = raw[:2] + ":" + raw[2:]
	}
	if v, ok := get(2); ok {
		acta.ElectoresHabilitados = v
	} else {
		report.Warnings = append(report.Warnings, "electores_habilitados no detectado")
		report.Confidence -= 0.1
	}
	if v, ok := get(3); ok {
		acta.PapeletasAnfora = v
	} else {
		report.Confidence -= 0.05
	}
	if v, ok := get(4); ok {
		acta.PapeletasNoUsadas = v
	} else {
		report.Confidence -= 0.05
	}

	for i, id := range []string{"P1", "P2", "P3", "P4"} {
		votos := 0
		if v, ok := get(5 + i); ok {
			votos = v
		} else {
			report.Warnings = append(report.Warnings, fmt.Sprintf("Votos %s no detectados", id))
			report.Confidence -= 0.05
		}
		acta.Candidatos = append(acta.Candidatos, models.Candidato{
			CandidatoID: id,
			Nombre:      candidateNames[id],
			Votos:       votos,
		})
	}

	if v, ok := get(9); ok {
		acta.VotosValidos = v
	} else {
		report.Warnings = append(report.Warnings, "votos_validos no detectado")
		report.Confidence -= 0.2
		report.Anomalies = appendUnique(report.Anomalies, "CAMPOS_CRITICOS_FALTANTES")
	}
	if v, ok := get(10); ok {
		acta.VotosBlancos = v
	} else {
		report.Warnings = append(report.Warnings, "votos_blancos no detectado con confianza suficiente")
		report.Confidence -= 0.2
		report.Anomalies = appendUnique(report.Anomalies, "CAMPOS_CRITICOS_FALTANTES")
	}
	if v, ok := get(11); ok {
		acta.VotosNulos = v
	} else {
		report.Warnings = append(report.Warnings, "votos_nulos no detectado")
		report.Confidence -= 0.2
		report.Anomalies = appendUnique(report.Anomalies, "CAMPOS_CRITICOS_FALTANTES")
	}

	if acta.PapeletasAnfora > 0 {
		acta.TotalVotos = acta.PapeletasAnfora
	} else {
		acta.TotalVotos = acta.VotosValidos + acta.VotosBlancos + acta.VotosNulos
	}
}

// ─── misc helpers ─────────────────────────────────────────────────────────────

func isPDFFile(fileBytes []byte, filename string) bool {
	if strings.HasSuffix(strings.ToLower(filename), ".pdf") {
		return true
	}
	return len(fileBytes) >= 4 && string(fileBytes[:4]) == "%PDF"
}

func extractActaID(text string) string {
	if m := reActaID.FindString(text); m != "" {
		return m
	}
	if m := reActaIDSpaced.FindStringSubmatch(text); len(m) == 4 {
		return m[1] + m[2] + m[3]
	}
	for _, line := range splitLines(text) {
		acc := ""
		for _, field := range strings.Fields(line) {
			digits := digitsOnlyToken(field)
			if digits == "" {
				acc = ""
				continue
			}
			if len(acc)+len(digits) > 13 {
				acc = digits
			} else {
				acc += digits
			}
			if len(acc) == 13 {
				return acc
			}
			if len(acc) > 13 {
				acc = ""
			}
		}
	}
	return ""
}

func lineContainsActaID(line, actaID string) bool {
	if strings.Contains(line, actaID) {
		return true
	}
	return extractActaID(line) == actaID
}

func digitsOnlyToken(s string) string {
	s = strings.Trim(s, " \t|[](){}:;,.<>\"'")
	if s == "" {
		return ""
	}
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return ""
		}
	}
	return s
}

func onlyDigits(s string) string {
	var b strings.Builder
	for _, ch := range s {
		if ch >= '0' && ch <= '9' {
			b.WriteRune(ch)
		}
	}
	return b.String()
}

func binarizeAndScale(img image.Image, rect image.Rectangle, scale int) image.Image {
	if scale < 1 {
		scale = 1
	}
	out := image.NewGray(image.Rect(0, 0, rect.Dx()*scale, rect.Dy()*scale))
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			r8, g8, b8 := uint8(r>>8), uint8(g>>8), uint8(b>>8)
			luma := (299*int(r8) + 587*int(g8) + 114*int(b8)) / 1000
			c := color.Gray{Y: 255}
			if luma < 150 {
				c = color.Gray{Y: 0}
			}
			for yy := 0; yy < scale; yy++ {
				for xx := 0; xx < scale; xx++ {
					out.SetGray((x-rect.Min.X)*scale+xx, (y-rect.Min.Y)*scale+yy, c)
				}
			}
		}
	}
	return out
}

func runTesseractText(regionPath string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "tesseract", regionPath, "stdout", "-l", "spa+eng", "--psm", "6")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	text := strings.TrimSpace(string(out))
	if text == "" {
		return "", fmt.Errorf("region sin texto")
	}
	return text, nil
}

func applyRegionLocation(acta *models.RRVActa, text string) {
	for _, line := range splitLines(text) {
		line = strings.ReplaceAll(line, "¿", ":")
		line = strings.ReplaceAll(line, ";", ":")
		line = strings.ReplaceAll(line, "+", ":")
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		label := strings.ToLower(strings.TrimSpace(parts[0]))
		value := strings.TrimSpace(parts[1])
		if value == "" {
			continue
		}
		switch {
		case strings.Contains(label, "departamento"):
			acta.Departamento = value
		case strings.Contains(label, "provincia"):
			acta.Provincia = value
		case strings.Contains(label, "municipio"):
			acta.Municipio = value
		case strings.Contains(label, "recinto"):
			acta.Recinto = value
		}
	}
}

func applyRegionNumber(acta *models.RRVActa, name string, value int, report *OCRReport) {
	maxValue := 1200
	switch name {
	case "P1", "P2", "P3", "P4", "votos_validos":
		maxValue = 999
	case "votos_blancos", "votos_nulos":
		maxValue = 300
	case "nro_mesa":
		maxValue = 999
	}
	if value > maxValue {
		report.Warnings = append(report.Warnings, fmt.Sprintf("%s=%d fuera de rango razonable; ignorado", name, value))
		report.Anomalies = appendUnique(report.Anomalies, "OCR_REGION_VALOR_DESCARTADO")
		report.Confidence -= 0.04
		return
	}
	switch name {
	case "nro_mesa":
		if value > 0 && value < 1000 {
			acta.NroMesa = strconv.Itoa(value)
			acta.Mesa = acta.NroMesa
		}
	case "electores_habilitados":
		acta.ElectoresHabilitados = value
	case "papeletas_anfora":
		acta.PapeletasAnfora = value
		acta.TotalVotos = value
	case "papeletas_no_usadas":
		acta.PapeletasNoUsadas = value
	case "P1", "P2", "P3", "P4":
		for i := range acta.Candidatos {
			if acta.Candidatos[i].CandidatoID == name {
				acta.Candidatos[i].Votos = value
				return
			}
		}
		acta.Candidatos = append(acta.Candidatos, models.Candidato{
			CandidatoID: name,
			Nombre:      candidateNames[name],
			Votos:       value,
		})
	case "votos_validos":
		acta.VotosValidos = value
	case "votos_blancos":
		acta.VotosBlancos = value
	case "votos_nulos":
		acta.VotosNulos = value
	}
}

func syncActaFromReport(acta *models.RRVActa, report OCRReport) {
	acta.OCRMode = report.Mode
	acta.Confidence = report.Confidence
	acta.RequiresReview = report.RequiresReview
	acta.Warnings = report.Warnings
	acta.Anomalies = report.Anomalies
}

func splitLines(text string) []string {
	return strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
}

func appendUnique(sl []string, s string) []string {
	for _, v := range sl {
		if v == s {
			return sl
		}
	}
	return append(sl, s)
}

func containsStr(sl []string, s string) bool {
	for _, v := range sl {
		if v == s {
			return true
		}
	}
	return false
}

// ContainsAnomaly is the exported version of containsStr for use in handlers.
func ContainsAnomaly(sl []string, s string) bool { return containsStr(sl, s) }
