package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"srrv/internal/models"
	"srrv/internal/repository"
	"srrv/internal/services"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

const (
	hdrContentType = "Content-Type"
	mimeXML        = "text/xml"
	twimlEmpty     = "<Response/>"
)

var validPINs = map[string]bool{
	"1234": true,
	"5678": true,
	"9012": true,
	"2025": true,
	"4321": true,
}

type RRVHandler struct {
	actaRepo   *repository.RRVActaRepository
	eventoRepo *repository.EventoRepository
	twilio     *services.TwilioClient
}

func NewRRVHandler(ar *repository.RRVActaRepository, er *repository.EventoRepository, tw *services.TwilioClient) *RRVHandler {
	if extra := os.Getenv("VALID_PINS"); extra != "" {
		for _, pin := range strings.Split(extra, ",") {
			if p := strings.TrimSpace(pin); p != "" {
				validPINs[p] = true
			}
		}
	}
	return &RRVHandler{actaRepo: ar, eventoRepo: er, twilio: tw}
}

func (h *RRVHandler) registrarEvento(actaID, tipo, fuente string, payload bson.M, errMsg string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ev := &models.Evento{
		ActaID:    actaID,
		Tipo:      tipo,
		Fuente:    fuente,
		Payload:   payload,
		Error:     errMsg,
		Timestamp: time.Now(),
	}
	if err := h.eventoRepo.Create(ctx, ev); err != nil {
		log.Printf("⚠️  Error registrando evento %s para acta %s: %v", tipo, actaID, err)
	}
}

// validarAritmetica checks vote totals. Returns "" if consistent, error message otherwise.
func validarAritmetica(acta *models.RRVActa) string {
	sumaCand := 0
	for _, c := range acta.Candidatos {
		if c.Votos < 0 {
			return fmt.Sprintf("votos negativos en candidato %s", c.CandidatoID)
		}
		sumaCand += c.Votos
	}
	if acta.VotosNulos < 0 {
		return "votos_nulos no puede ser negativo"
	}
	if acta.VotosBlancos < 0 {
		return "votos_blancos no puede ser negativo"
	}
	if acta.VotosValidos > 0 && sumaCand != acta.VotosValidos {
		return fmt.Sprintf("suma candidatos=%d ≠ votos_validos=%d", sumaCand, acta.VotosValidos)
	}
	suma := sumaCand + acta.VotosNulos + acta.VotosBlancos
	if suma != acta.TotalVotos {
		return fmt.Sprintf("inconsistencia aritmética: suma=%d, total_votos=%d", suma, acta.TotalVotos)
	}
	return ""
}

func totalVotos(acta *models.RRVActa) int {
	t := acta.VotosNulos + acta.VotosBlancos
	for _, c := range acta.Candidatos {
		t += c.Votos
	}
	return t
}

func normalizarCandidatoID(id string) string {
	id = strings.ToUpper(strings.TrimSpace(id))
	switch id {
	case "CAND-01", "CAND-1", "C1":
		return "P1"
	case "CAND-02", "CAND-2", "C2":
		return "P2"
	case "CAND-03", "CAND-3", "C3":
		return "P3"
	case "CAND-04", "CAND-4", "C4":
		return "P4"
	default:
		return id
	}
}

func normalizarActaRRV(acta *models.RRVActa) {
	if acta == nil {
		return
	}
	if strings.TrimSpace(acta.Provincia) == "" {
		acta.Provincia = "(Sin provincia)"
	}
	if strings.TrimSpace(acta.Fuente) == "" {
		acta.Fuente = "RRV"
	}
	if acta.Warnings == nil {
		acta.Warnings = []string{}
	}
	if acta.Anomalies == nil {
		acta.Anomalies = []string{}
	}

	porID := make(map[string]models.Candidato, len(acta.Candidatos)+4)
	for _, c := range acta.Candidatos {
		id := normalizarCandidatoID(c.CandidatoID)
		if id == "" {
			continue
		}
		c.CandidatoID = id
		if strings.TrimSpace(c.Nombre) == "" {
			c.Nombre = id
		}
		if prev, ok := porID[id]; ok {
			prev.Votos += c.Votos
			if strings.TrimSpace(prev.Nombre) == "" || prev.Nombre == id {
				prev.Nombre = c.Nombre
			}
			porID[id] = prev
			continue
		}
		porID[id] = c
	}

	defaultNames := map[string]string{
		"P1": "Daenerys Targaryen",
		"P2": "Sansa Stark",
		"P3": "Robert Baratheon",
		"P4": "Tyrion Lannister",
	}
	normalizados := make([]models.Candidato, 0, 4)
	for _, id := range []string{"P1", "P2", "P3", "P4"} {
		c, ok := porID[id]
		if !ok {
			c = models.Candidato{CandidatoID: id, Nombre: defaultNames[id], Votos: 0}
		}
		normalizados = append(normalizados, c)
		delete(porID, id)
	}
	for _, c := range porID {
		normalizados = append(normalizados, c)
	}
	acta.Candidatos = normalizados

	// Compute VotosValidos from candidates if not already set
	if acta.VotosValidos == 0 {
		for _, c := range acta.Candidatos {
			acta.VotosValidos += c.Votos
		}
	}
	// Compute TotalVotos if missing
	if acta.TotalVotos == 0 {
		acta.TotalVotos = totalVotos(acta)
	}
	// Sync NroMesa ↔ Mesa
	if acta.NroMesa == "" && acta.Mesa != "" {
		acta.NroMesa = acta.Mesa
	}
	if acta.Mesa == "" && acta.NroMesa != "" {
		acta.Mesa = acta.NroMesa
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// POST /api/rrv/actas/upload
// ──────────────────────────────────────────────────────────────────────────────
func (h *RRVHandler) Upload(c *gin.Context) {
	// 1. Leer archivo
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "campo 'file' requerido"})
		return
	}
	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo abrir el archivo"})
		return
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error leyendo el archivo"})
		return
	}

	// 2. Hash para idempotencia
	hash := services.HashBytes(fileBytes)

	// 3. Verificar duplicado por hash
	//    Si OCR_ALLOW_REPROCESS=true y ?reprocess=true, se borra el registro anterior
	//    y se reprocesa. Idempotencia normal permanece activa por defecto.
	requestedReprocess := strings.EqualFold(c.Query("reprocess"), "true")
	allowReprocess := strings.EqualFold(os.Getenv("OCR_ALLOW_REPROCESS"), "true")
	reprocess := requestedReprocess && allowReprocess

	dupHash, err := h.actaRepo.ExistsByHash(c.Request.Context(), hash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if dupHash {
		if reprocess {
			if delErr := h.actaRepo.DeleteByHash(c.Request.Context(), hash); delErr != nil {
				log.Printf("⚠️  reprocess: error borrando acta por hash: %v", delErr)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "no se pudo preparar el reprocesamiento",
					"hash":  hash,
				})
				return
			}
			log.Printf("🔄 Reprocesando acta hash=%s…", hash[:16])
		} else if requestedReprocess {
			h.registrarEvento("desconocido", "REPROCESS_DESHABILITADO", "IMAGEN",
				bson.M{"hash": hash, "filename": fileHeader.Filename},
				"OCR_ALLOW_REPROCESS=false")
			c.JSON(http.StatusConflict, gin.H{
				"error": "reprocess=true solicitado, pero OCR_ALLOW_REPROCESS=false; no se omite idempotencia",
				"hash":  hash,
			})
			return
		} else {
			h.registrarEvento("desconocido", "DUPLICADO_HASH", "IMAGEN",
				bson.M{"hash": hash, "filename": fileHeader.Filename},
				"archivo ya procesado anteriormente")
			c.JSON(http.StatusConflict, gin.H{"error": "este archivo ya fue procesado", "hash": hash})
			return
		}
	}

	// 4. Obtener datos del acta: MANUAL o pipeline OCR
	var acta *models.RRVActa
	var ocrReport services.OCRReport

	if raw := c.PostForm("acta_data"); raw != "" {
		// Modo MANUAL — datos provistos por el usuario
		acta = &models.RRVActa{}
		if err := json.Unmarshal([]byte(raw), acta); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "acta_data JSON inválido: " + err.Error()})
			return
		}
		acta.TipoEntrada = "IMAGEN"
		acta.Fuente = "RRV"
		acta.OCRMode = "MANUAL"
		acta.Confidence = 1.0
		ocrReport = services.OCRReport{
			Mode:       "MANUAL",
			Confidence: 1.0,
			Warnings:   []string{},
			Anomalies:  []string{},
		}
	} else {
		// Pipeline OCR — intenta leer el documento real
		acta, ocrReport, _ = services.ProcessActa(fileBytes, fileHeader.Filename)
		log.Printf("🔍 OCR acta=%s mode=%s confidence=%.2f requires_review=%v",
			acta.ActaID, ocrReport.Mode, ocrReport.Confidence, ocrReport.RequiresReview)
	}

	normalizarActaRRV(acta)
	acta.HashOrigen = hash
	acta.FechaRecepcion = time.Now()

	// 5. Verificar duplicado por acta_id
	dupID, err := h.actaRepo.ExistsByActaID(c.Request.Context(), acta.ActaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if dupID {
		h.registrarEvento(acta.ActaID, "DUPLICADO_ACTA_ID", "IMAGEN",
			bson.M{"acta_id": acta.ActaID, "hash": hash},
			"acta_id ya existe en el sistema")
		c.JSON(http.StatusConflict, gin.H{"error": "acta ya registrada", "acta_id": acta.ActaID})
		return
	}

	// 6. Para MANUAL: validar aritmética (el OCR pipeline ya lo hizo internamente)
	if acta.OCRMode == "MANUAL" {
		if msg := validarAritmetica(acta); msg != "" {
			acta.Estado = "INCONSISTENTE"
			ocrReport.Warnings = append(ocrReport.Warnings, msg)
			ocrReport.Anomalies = append(ocrReport.Anomalies, "INCONSISTENCIA_ARITMETICA")
		} else {
			acta.Estado = "PROCESADA"
		}
	}

	// 7. Guardar en MongoDB
	saved, err := h.actaRepo.Create(c.Request.Context(), acta)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 8. Registrar eventos OCR
	h.registrarEventosOCR(saved, ocrReport, hash)
	if reprocess {
		h.registrarEvento(saved.ActaID, "ACTA_REPROCESADA", "IMAGEN",
			bson.M{"acta_id": saved.ActaID, "hash": hash, "filename": fileHeader.Filename},
			"reprocesamiento solicitado")
	}

	// 9. Responder con formato enriquecido
	msg := "Acta procesada"
	if saved.RequiresReview {
		msg = "Acta procesada con baja confianza; requiere revisión"
	}
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": msg,
		"acta":    saved,
		"ocr": gin.H{
			"mode":            ocrReport.Mode,
			"confidence":      ocrReport.Confidence,
			"requires_review": ocrReport.RequiresReview,
			"warnings":        ocrReport.Warnings,
			"anomalies":       ocrReport.Anomalies,
		},
	})
}

// registrarEventosOCR emite los eventos de auditoría correspondientes al resultado del OCR.
func (h *RRVHandler) registrarEventosOCR(acta *models.RRVActa, report services.OCRReport, hash string) {
	base := bson.M{
		"acta_id":    acta.ActaID,
		"hash":       hash,
		"ocr_mode":   report.Mode,
		"confidence": report.Confidence,
	}

	switch report.Mode {
	case "PDF_TEXT":
		h.registrarEvento(acta.ActaID, "OCR_PDF_TEXT_EXTRAIDO", "IMAGEN", base, "")
		if report.RequiresReview {
			h.registrarEvento(acta.ActaID, "OCR_PARSE_PARCIAL", "IMAGEN",
				mergeBSON(base, bson.M{"warnings": report.Warnings}), "")
			h.registrarEvento(acta.ActaID, "OCR_BAJA_CONFIANZA", "IMAGEN",
				mergeBSON(base, bson.M{"confidence": report.Confidence}), "")
		} else {
			h.registrarEvento(acta.ActaID, "OCR_PARSE_OK", "IMAGEN", base, "")
		}

	case "PDF_IMAGE_OCR", "IMAGE_OCR", "PDF_REGION_OCR":
		h.registrarEvento(acta.ActaID, "OCR_PDF_TEXT_FALLIDO", "IMAGEN", base, "")
		if services.ContainsAnomaly(report.Anomalies, "OCR_IMAGE_RENDER_OK") {
			h.registrarEvento(acta.ActaID, "OCR_IMAGE_RENDER_OK", "IMAGEN", base, "")
		}
		if services.ContainsAnomaly(report.Anomalies, "OCR_TESSERACT_OK") {
			h.registrarEvento(acta.ActaID, "OCR_TESSERACT_OK", "IMAGEN", base, "")
		} else {
			h.registrarEvento(acta.ActaID, "OCR_TESSERACT_ERROR", "IMAGEN", base,
				"tesseract no produjo texto reconocible")
		}
		if services.ContainsAnomaly(report.Anomalies, "OCR_REGION_OK") {
			h.registrarEvento(acta.ActaID, "OCR_REGION_OK", "IMAGEN", base, "")
		}
		if report.RequiresReview {
			h.registrarEvento(acta.ActaID, "OCR_PARSE_PARCIAL", "IMAGEN",
				mergeBSON(base, bson.M{"warnings": report.Warnings}), "")
			h.registrarEvento(acta.ActaID, "OCR_BAJA_CONFIANZA", "IMAGEN",
				mergeBSON(base, bson.M{"confidence": report.Confidence}), "")
		} else {
			h.registrarEvento(acta.ActaID, "OCR_PARSE_OK", "IMAGEN", base, "")
		}

	case "SIMULATED_FALLBACK":
		h.registrarEvento(acta.ActaID, "OCR_PDF_TEXT_FALLIDO", "IMAGEN", base, "")
		if services.ContainsAnomaly(report.Anomalies, "OCR_IMAGE_RENDER_OK") {
			h.registrarEvento(acta.ActaID, "OCR_IMAGE_RENDER_OK", "IMAGEN", base, "")
		}
		if services.ContainsAnomaly(report.Anomalies, "OCR_TESSERACT_OK") {
			h.registrarEvento(acta.ActaID, "OCR_TESSERACT_OK", "IMAGEN", base, "")
		}
		if services.ContainsAnomaly(report.Anomalies, "OCR_TESSERACT_ERROR") ||
			services.ContainsAnomaly(report.Anomalies, "OCR_VISUAL_ERROR") {
			h.registrarEvento(acta.ActaID, "OCR_TESSERACT_ERROR", "IMAGEN", base,
				"tesseract/pdftoppm no produjo texto parseable")
		}
		if services.ContainsAnomaly(report.Anomalies, "OCR_VISUAL_PARCIAL") {
			h.registrarEvento(acta.ActaID, "OCR_PARSE_PARCIAL", "IMAGEN",
				mergeBSON(base, bson.M{"warnings": report.Warnings}), "")
		}
		h.registrarEvento(acta.ActaID, "OCR_SIMULADO_USADO", "IMAGEN",
			mergeBSON(base, bson.M{"advertencia": "datos simulados, no OCR real"}), "")
	case "MANUAL":
		h.registrarEvento(acta.ActaID, "ACTA_INGRESO_MANUAL", "IMAGEN", base, "")
	}

	switch acta.Estado {
	case "PROCESADA":
		h.registrarEvento(acta.ActaID, "ACTA_PROCESADA", "IMAGEN",
			mergeBSON(base, bson.M{"estado": acta.Estado}), "")
	case "INCONSISTENTE":
		h.registrarEvento(acta.ActaID, "ACTA_INCONSISTENTE", "IMAGEN",
			mergeBSON(base, bson.M{"warnings": report.Warnings}),
			"inconsistencia aritmética detectada")
	case "REQUIERE_REVISION":
		h.registrarEvento(acta.ActaID, "ACTA_REQUIERE_REVISION", "IMAGEN",
			mergeBSON(base, bson.M{"confidence": report.Confidence, "warnings": report.Warnings}), "")
	}
}

func mergeBSON(base, extra bson.M) bson.M {
	result := bson.M{}
	for k, v := range base {
		result[k] = v
	}
	for k, v := range extra {
		result[k] = v
	}
	return result
}

// ──────────────────────────────────────────────────────────────────────────────
// POST /api/rrv/sms
// ──────────────────────────────────────────────────────────────────────────────
func (h *RRVHandler) SMS(c *gin.Context) {
	var body struct {
		Mensaje string `json:"mensaje" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	msgHash := services.HashBytes([]byte(body.Mensaje))
	dupMsg, err := h.actaRepo.ExistsByHash(c.Request.Context(), msgHash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if dupMsg {
		h.registrarEvento("desconocido", "DUPLICADO_SMS", "SMS",
			bson.M{"hash": msgHash}, "mensaje SMS ya recibido anteriormente")
		c.JSON(http.StatusConflict, gin.H{"error": "mensaje SMS duplicado"})
		return
	}

	acta, parseErr := parseSMS(body.Mensaje)
	if parseErr != "" {
		h.registrarEvento("desconocido", "SMS_INVALIDO", "SMS",
			bson.M{"mensaje": body.Mensaje, "detalle": parseErr}, parseErr)
		c.JSON(http.StatusBadRequest, gin.H{"error": parseErr})
		return
	}
	acta.HashOrigen = msgHash
	acta.FechaRecepcion = time.Now()
	normalizarActaRRV(acta)

	pin := extractField(body.Mensaje, "PIN")
	if !validPINs[pin] {
		h.registrarEvento(acta.ActaID, "PIN_INVALIDO", "SMS",
			bson.M{"acta_id": acta.ActaID}, "PIN de seguridad inválido")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "PIN inválido o no autorizado"})
		return
	}

	dupID, err := h.actaRepo.ExistsByActaID(c.Request.Context(), acta.ActaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if dupID {
		h.registrarEvento(acta.ActaID, "DUPLICADO_ACTA_ID", "SMS",
			bson.M{"acta_id": acta.ActaID}, "acta_id ya existe en el sistema")
		c.JSON(http.StatusConflict, gin.H{"error": "acta ya registrada", "acta_id": acta.ActaID})
		return
	}

	if msg := validarAritmetica(acta); msg != "" {
		acta.Estado = "INCONSISTENTE"
		h.registrarEvento(acta.ActaID, "INCONSISTENCIA_ARITMETICA", "SMS",
			bson.M{"acta_id": acta.ActaID, "detalle": msg}, msg)
		log.Printf("⚠️  INCONSISTENCIA_ARITMETICA acta=%s (SMS): %s", acta.ActaID, msg)
	} else {
		acta.Estado = "PROCESADA"
	}

	saved, err := h.actaRepo.Create(c.Request.Context(), acta)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.registrarEvento(saved.ActaID, "ACTA_RECIBIDA", "SMS",
		bson.M{"acta_id": saved.ActaID, "estado": saved.Estado}, "")

	if h.twilio != nil {
		go h.twilio.SendConfirmation("", fmt.Sprintf(
			"✅ SRRV: Acta %s recibida por SMS. Estado: %s", saved.ActaID, saved.Estado,
		))
	}
	c.JSON(http.StatusCreated, saved)
}

// ──────────────────────────────────────────────────────────────────────────────
// POST /api/rrv/webhook/sms
// ──────────────────────────────────────────────────────────────────────────────
func (h *RRVHandler) WebhookSMS(c *gin.Context) {
	mensaje := strings.TrimSpace(c.PostForm("Body"))
	from := c.PostForm("From")

	confirm, err := h.procesarWebhookSMS(c.Request.Context(), mensaje, from)
	if err != nil {
		log.Printf("⚠️  Webhook SMS: %v", err)
	}
	if h.twilio != nil && confirm != "" {
		go h.twilio.SendConfirmation(from, confirm)
	}
	twimlOK(c)
}

func (h *RRVHandler) procesarWebhookSMS(ctx context.Context, mensaje, from string) (string, error) {
	if mensaje == "" {
		return "", nil
	}

	msgHash := services.HashBytes([]byte(mensaje))
	dupMsg, err := h.actaRepo.ExistsByHash(ctx, msgHash)
	if err != nil {
		return "", fmt.Errorf("verificando hash: %w", err)
	}
	if dupMsg {
		h.registrarEvento("desconocido", "DUPLICADO_SMS", "TWILIO_WEBHOOK",
			bson.M{"hash": msgHash, "from": from}, "mensaje SMS duplicado")
		return "⚠️ SRRV: Este mensaje de acta ya fue procesado anteriormente.", nil
	}

	acta, parseErr := parseSMS(mensaje)
	if parseErr != "" {
		h.registrarEvento("desconocido", "SMS_INVALIDO", "TWILIO_WEBHOOK",
			bson.M{"mensaje": mensaje, "from": from, "detalle": parseErr}, parseErr)
		return fmt.Sprintf("❌ SRRV: Formato inválido — %s", parseErr), nil
	}
	acta.HashOrigen = msgHash
	acta.FechaRecepcion = time.Now()
	normalizarActaRRV(acta)

	if !validPINs[extractField(mensaje, "PIN")] {
		h.registrarEvento(acta.ActaID, "PIN_INVALIDO", "TWILIO_WEBHOOK",
			bson.M{"acta_id": acta.ActaID, "from": from}, "PIN inválido")
		return "❌ SRRV: PIN de seguridad inválido. Acta rechazada.", nil
	}

	dupID, err := h.actaRepo.ExistsByActaID(ctx, acta.ActaID)
	if err != nil {
		return "", fmt.Errorf("verificando acta_id: %w", err)
	}
	if dupID {
		h.registrarEvento(acta.ActaID, "DUPLICADO_ACTA_ID", "TWILIO_WEBHOOK",
			bson.M{"acta_id": acta.ActaID, "from": from}, "acta_id ya existe")
		return fmt.Sprintf("⚠️ SRRV: Acta %s ya fue registrada.", acta.ActaID), nil
	}

	if msg := validarAritmetica(acta); msg != "" {
		acta.Estado = "INCONSISTENTE"
		h.registrarEvento(acta.ActaID, "INCONSISTENCIA_ARITMETICA", "TWILIO_WEBHOOK",
			bson.M{"acta_id": acta.ActaID, "detalle": msg, "from": from}, msg)
	} else {
		acta.Estado = "PROCESADA"
	}

	saved, err := h.actaRepo.Create(ctx, acta)
	if err != nil {
		return "", fmt.Errorf("guardando acta: %w", err)
	}

	h.registrarEvento(saved.ActaID, "ACTA_RECIBIDA", "TWILIO_WEBHOOK",
		bson.M{"acta_id": saved.ActaID, "estado": saved.Estado, "from": from}, "")

	return fmt.Sprintf("✅ SRRV: Acta %s registrada. Estado: %s. Total votos: %d",
		saved.ActaID, saved.Estado, saved.TotalVotos), nil
}

func twimlOK(c *gin.Context) {
	c.Header(hdrContentType, mimeXML)
	c.String(http.StatusOK, twimlEmpty)
}

// ──────────────────────────────────────────────────────────────────────────────
// GET /api/rrv/actas
// ──────────────────────────────────────────────────────────────────────────────
func (h *RRVHandler) GetAll(c *gin.Context) {
	actas, err := h.actaRepo.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, actas)
}

// ──────────────────────────────────────────────────────────────────────────────
// GET /api/rrv/actas/:acta_id
// ──────────────────────────────────────────────────────────────────────────────
func (h *RRVHandler) GetByID(c *gin.Context) {
	actaID := c.Param("acta_id")
	acta, err := h.actaRepo.GetByActaID(c.Request.Context(), actaID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "acta no encontrada"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, acta)
}

// ──────────────────────────────────────────────────────────────────────────────
// GET /api/rrv/eventos
// ──────────────────────────────────────────────────────────────────────────────
func (h *RRVHandler) GetEventos(c *gin.Context) {
	eventos, err := h.eventoRepo.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, eventos)
}

// ─── SMS helpers ──────────────────────────────────────────────────────────────

func extractField(msg, key string) string {
	for _, part := range strings.Split(msg, "|") {
		kv := strings.SplitN(part, ":", 2)
		if len(kv) == 2 && strings.EqualFold(strings.TrimSpace(kv[0]), key) {
			return strings.TrimSpace(kv[1])
		}
	}
	return ""
}

// parseSMS parses the pipe-separated SMS format into an RRVActa.
// Format: ACTA:id|DEP:dpto|PROV:prov|MUN:mun|REC:recinto|MESA:mesa|C1:v|C2:v|...|NULOS:n|BLANCOS:b|TOTAL:t|PIN:pin
func parseSMS(msg string) (*models.RRVActa, string) {
	get := func(key string) string { return extractField(msg, key) }

	actaID := get("ACTA")
	dep := get("DEP")
	prov := get("PROV")
	mun := get("MUN")
	rec := get("REC")
	mesa := get("MESA")
	nulosStr := get("NULOS")
	blancosStr := get("BLANCOS")
	totalStr := get("TOTAL")

	switch "" {
	case actaID:
		return nil, "campo ACTA requerido"
	case dep:
		return nil, "campo DEP requerido"
	case mun:
		return nil, "campo MUN requerido"
	case rec:
		return nil, "campo REC requerido"
	case mesa:
		return nil, "campo MESA requerido"
	case nulosStr:
		return nil, "campo NULOS requerido"
	case blancosStr:
		return nil, "campo BLANCOS requerido"
	case totalStr:
		return nil, "campo TOTAL requerido"
	}

	nulos, err := strconv.Atoi(nulosStr)
	if err != nil || nulos < 0 {
		return nil, "NULOS debe ser un entero no negativo"
	}
	blancos, err := strconv.Atoi(blancosStr)
	if err != nil || blancos < 0 {
		return nil, "BLANCOS debe ser un entero no negativo"
	}
	total, err := strconv.Atoi(totalStr)
	if err != nil || total < 0 {
		return nil, "TOTAL debe ser un entero no negativo"
	}

	var candidatos []models.Candidato
	for i := 1; ; i++ {
		vStr := get(fmt.Sprintf("C%d", i))
		if vStr == "" {
			break
		}
		v, err := strconv.Atoi(vStr)
		if err != nil || v < 0 {
			return nil, fmt.Sprintf("C%d debe ser un entero no negativo", i)
		}
		candidatos = append(candidatos, models.Candidato{
			CandidatoID: fmt.Sprintf("CAND-%02d", i),
			Nombre:      fmt.Sprintf("Candidato %d", i),
			Votos:       v,
		})
	}
	if len(candidatos) == 0 {
		return nil, "se requiere al menos un candidato (C1)"
	}

	return &models.RRVActa{
		ActaID:       actaID,
		Departamento: dep,
		Provincia:    prov,
		Municipio:    mun,
		Recinto:      rec,
		Mesa:         mesa,
		NroMesa:      mesa,
		Candidatos:   candidatos,
		VotosNulos:   nulos,
		VotosBlancos: blancos,
		TotalVotos:   total,
		Fuente:       "RRV",
		TipoEntrada:  "SMS",
		OCRMode:      "MANUAL",
		Confidence:   1.0,
		Warnings:     []string{},
		Anomalies:    []string{},
	}, ""
}
