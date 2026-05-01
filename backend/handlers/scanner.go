package handlers

import (
	"encoding/csv"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/srvof/votos-backend/models"
	"gorm.io/gorm"
)

// ScannerHandler maneja la transcripción de actas a partir de un archivo cargado.
type ScannerHandler struct {
	DB *gorm.DB
}

type scannerRow struct {
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

// Transcribir recibe multipart/form-data con campo 'archivo' y simula transcripción de acta.
// Si llega codigo_acta o acta_id busca esa acta; si no, usa una fila aleatoria de Transcripciones.csv.
// Requiere JWT (protegido en routes.go).
func (h *ScannerHandler) Transcribir(c *gin.Context) {
	file, header, err := c.Request.FormFile("archivo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Se requiere el campo 'archivo'"})
		return
	}
	defer file.Close()

	if header.Size > 10*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El archivo supera el tamaño máximo de 10 MB"})
		return
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".pdf": true}
	if !allowed[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Extensión no permitida. Use: jpg, jpeg, png o pdf"})
		return
	}

	codigoActaStr := c.PostForm("codigo_acta")
	actaIDStr := c.PostForm("acta_id")

	payload, err := loadScannerRow(codigoActaStr, actaIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al leer datos de transcripción: " + err.Error()})
		return
	}

	estado := "transcrita"
	if strings.TrimSpace(payload.Observaciones) != "" {
		estado = "observada"
	}

	actaActualizada := false
	var acta models.Acta
	if payload.CodigoActa > 0 {
		if dbErr := h.DB.Where("codigo_acta = ?", payload.CodigoActa).First(&acta).Error; dbErr == nil {
			updates := map[string]interface{}{
				"estado":               estado,
				"p1":                   payload.P1,
				"p2":                   payload.P2,
				"p3":                   payload.P3,
				"p4":                   payload.P4,
				"votos_validos":        payload.VotosValidos,
				"votos_nulos":          payload.VotosNulos,
				"votos_blanco":         payload.VotosBlanco,
				"papeletas_anfora":     payload.PapeletasAnfora,
				"papeletas_no_usadas":  payload.PapeletasNoUsadas,
				"observaciones":        payload.Observaciones,
				"apertura_hora":        payload.AperturaHora,
				"apertura_minutos":     payload.AperturaMinutos,
				"cierre_hora":          payload.CierreHora,
				"cierre_minutos":       payload.CierreMinutos,
			}
			if dbErr2 := h.DB.Model(&acta).Updates(updates).Error; dbErr2 == nil {
				actaActualizada = true
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Acta transcrita correctamente",
		"estado":  estado,
		"acta_id": payload.CodigoActa,
		"data": gin.H{
			"codigo_acta":      payload.CodigoActa,
			"p1":               payload.P1,
			"p2":               payload.P2,
			"p3":               payload.P3,
			"p4":               payload.P4,
			"votos_validos":    payload.VotosValidos,
			"votos_nulos":      payload.VotosNulos,
			"votos_blancos":    payload.VotosBlanco,
			"papeletas_anfora": payload.PapeletasAnfora,
			"observaciones":    payload.Observaciones,
			"acta_actualizada": actaActualizada,
		},
	})
}

func loadScannerRow(codigoActaStr, actaIDStr string) (*scannerRow, error) {
	f, err := os.Open(filepath.Join("data", "Transcripciones.csv"))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.FieldsPerRecord = -1
	if _, err := r.Read(); err != nil {
		return nil, err
	}

	target := int64(0)
	if codigoActaStr != "" {
		target, _ = strconv.ParseInt(strings.TrimSpace(codigoActaStr), 10, 64)
	} else if actaIDStr != "" {
		target, _ = strconv.ParseInt(strings.TrimSpace(actaIDStr), 10, 64)
	}

	var rows []scannerRow
	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil || len(row) < 26 {
			continue
		}
		codigo, parseErr := strconv.ParseInt(strings.TrimSpace(row[8]), 10, 64)
		if parseErr != nil {
			continue
		}
		sr := scannerRow{
			CodigoActa:        codigo,
			P1:                parseCampo(row[13]),
			P2:                parseCampo(row[14]),
			P3:                parseCampo(row[15]),
			P4:                parseCampo(row[16]),
			VotosValidos:      parseCampo(row[17]),
			VotosBlanco:       parseCampo(row[18]),
			VotosNulos:        parseCampo(row[19]),
			PapeletasAnfora:   parseCampo(row[11]),
			PapeletasNoUsadas: parseCampo(row[12]),
			Observaciones:     strings.TrimSpace(row[20]),
			AperturaHora:      parseCampo(row[22]),
			AperturaMinutos:   parseCampo(row[23]),
			CierreHora:        parseCampo(row[24]),
			CierreMinutos:     parseCampo(row[25]),
		}
		if target > 0 && codigo == target {
			return &sr, nil
		}
		rows = append(rows, sr)
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("Transcripciones.csv no tiene filas válidas")
	}

	return &rows[rand.Intn(len(rows))], nil
}
