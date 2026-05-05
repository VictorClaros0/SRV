package dashboard

import (
	"context"
	"encoding/csv"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

type pdfScanRow struct {
	CodigoActa        int64
	P1                int
	P2                int
	P3                int
	P4                int
	VotosValidos      int
	VotosNulos        int
	VotosBlanco       int
	PapeletasAnfora   int
	PapeletasNoUsadas int
	Observaciones     string
	AperturaHora      int
	AperturaMinutos   int
	CierreHora        int
	CierreMinutos     int
}

// ProcessPDFOutput transcribe uno por uno los PDFs de pdf_output usando el codigo del nombre.
func (h *Handler) ProcessPDFOutput(c *gin.Context) {
	rows, err := loadPDFScanRows(filepath.Join("data", "Transcripciones.csv"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	pdfs := findPDFs(filepath.Join("..", "pdf_output"))
	if len(pdfs) == 0 {
		pdfs = findPDFs(filepath.Join("pdf_output"))
	}
	if len(pdfs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No se encontraron PDFs en pdf_output"})
		return
	}
	sort.Strings(pdfs)

	codeRe := regexp.MustCompile(`(\d+)`)
	muestra := []gin.H{}
	procesadas := 0
	sinDatos := 0
	noEncontradas := 0

	for _, pdf := range pdfs {
		name := filepath.Base(pdf)
		match := codeRe.FindStringSubmatch(name)
		if len(match) == 0 {
			sinDatos++
			if len(muestra) < 25 {
				muestra = append(muestra, gin.H{"archivo": name, "ok": false, "error": "sin codigo en nombre"})
			}
			continue
		}
		codigo, _ := strconv.ParseInt(match[1], 10, 64)
		row, ok := rows[codigo]
		if !ok {
			sinDatos++
			if len(muestra) < 25 {
				muestra = append(muestra, gin.H{"archivo": name, "codigo_acta": codigo, "ok": false, "error": "sin fila en Transcripciones.csv"})
			}
			continue
		}

		estado := "transcrita"
		if strings.TrimSpace(row.Observaciones) != "" {
			estado = "observada"
		}
		updates := map[string]interface{}{
			"estado":                estado,
			"p1":                    row.P1,
			"p2":                    row.P2,
			"p3":                    row.P3,
			"p4":                    row.P4,
			"votos_validos":         row.VotosValidos,
			"votos_nulos":           row.VotosNulos,
			"votos_blanco":          row.VotosBlanco,
			"papeletas_anfora":      row.PapeletasAnfora,
			"papeletas_no_usadas":   row.PapeletasNoUsadas,
			"observaciones":         row.Observaciones,
			"apertura_hora":         row.AperturaHora,
			"apertura_minutos":      row.AperturaMinutos,
			"cierre_hora":           row.CierreHora,
			"cierre_minutos":        row.CierreMinutos,
			"fecha_modificacion":    time.Now(),
		}

		tx := h.db.Table("acta").Where("codigo_acta = ? AND fecha_eliminado IS NULL", codigo).Updates(updates)
		if tx.Error != nil {
			if len(muestra) < 25 {
				muestra = append(muestra, gin.H{"archivo": name, "codigo_acta": codigo, "ok": false, "error": tx.Error.Error()})
			}
			continue
		}
		if tx.RowsAffected == 0 {
			noEncontradas++
			if len(muestra) < 25 {
				muestra = append(muestra, gin.H{"archivo": name, "codigo_acta": codigo, "ok": false, "error": "acta no encontrada"})
			}
			continue
		}
		procesadas++
		if len(muestra) < 25 {
			muestra = append(muestra, gin.H{"archivo": name, "codigo_acta": codigo, "ok": true, "estado": estado})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"total_pdfs":     len(pdfs),
		"procesadas":     procesadas,
		"sin_datos":      sinDatos,
		"no_encontradas": noEncontradas,
		"muestra":        muestra,
	})
}

// ResetConteos limpia conteo oficial (Postgres) y conteo rapido RRV (Mongo).
func (h *Handler) ResetConteos(c *gin.Context) {
	now := time.Now()
	updates := map[string]interface{}{
		"estado":                "impresa",
		"p1":                    0,
		"p2":                    0,
		"p3":                    0,
		"p4":                    0,
		"votos_validos":         0,
		"votos_nulos":           0,
		"votos_blanco":          0,
		"papeletas_anfora":      0,
		"papeletas_no_usadas":   0,
		"observaciones":         "",
		"apertura_hora":         0,
		"apertura_minutos":      0,
		"cierre_hora":           0,
		"cierre_minutos":        0,
		"fecha_modificacion":    now,
	}
	tx := h.db.Table("acta").Where("fecha_eliminado IS NULL").Updates(updates)
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": tx.Error.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 20*time.Second)
	defer cancel()
	deletedMongo := int64(0)
	seen := map[string]bool{}
	for _, colName := range h.svc.RRVCollections {
		colName = strings.TrimSpace(colName)
		if colName == "" || seen[colName] {
			continue
		}
		seen[colName] = true
		res, err := h.mongoClient.Database(h.svc.DBName).Collection(colName).DeleteMany(ctx, bson.D{})
		if err == nil {
			deletedMongo += res.DeletedCount
		}
	}
	if h.eventsCollection != "" {
		_, _ = h.mongoClient.Database(h.svc.DBName).Collection(h.eventsCollection).DeleteMany(ctx, bson.D{})
	}

	c.JSON(http.StatusOK, gin.H{
		"success":              true,
		"oficial_reseteadas":   tx.RowsAffected,
		"rrv_eliminadas":       deletedMongo,
		"colecciones_rrv":      h.svc.RRVCollections,
	})
}

func loadPDFScanRows(path string) (map[int64]pdfScanRow, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("no se pudo abrir %s: %w", path, err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.FieldsPerRecord = -1
	if _, err := r.Read(); err != nil {
		return nil, err
	}

	rows := map[int64]pdfScanRow{}
	for {
		row, err := r.Read()
		if err != nil {
			break
		}
		if len(row) < 26 {
			continue
		}
		codigo, err := strconv.ParseInt(strings.TrimSpace(row[8]), 10, 64)
		if err != nil {
			continue
		}
		rows[codigo] = pdfScanRow{
			CodigoActa:        codigo,
			P1:                parsePDFInt(row[13]),
			P2:                parsePDFInt(row[14]),
			P3:                parsePDFInt(row[15]),
			P4:                parsePDFInt(row[16]),
			VotosValidos:      parsePDFInt(row[17]),
			VotosBlanco:       parsePDFInt(row[18]),
			VotosNulos:        parsePDFInt(row[19]),
			PapeletasAnfora:   parsePDFInt(row[11]),
			PapeletasNoUsadas: parsePDFInt(row[12]),
			Observaciones:     strings.TrimSpace(row[20]),
			AperturaHora:      parsePDFInt(row[22]),
			AperturaMinutos:   parsePDFInt(row[23]),
			CierreHora:        parsePDFInt(row[24]),
			CierreMinutos:     parsePDFInt(row[25]),
		}
	}
	return rows, nil
}

func parsePDFInt(s string) int {
	n, _ := strconv.Atoi(strings.TrimSpace(s))
	return n
}

func findPDFs(root string) []string {
	pdfs := []string{}
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if strings.EqualFold(filepath.Ext(path), ".pdf") {
			pdfs = append(pdfs, path)
		}
		return nil
	})
	return pdfs
}
