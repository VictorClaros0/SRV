package handlers

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/srvof/votos-backend/models"
	"gorm.io/gorm"
)

const defaultN8NWebhookURL = "http://n8n:5678/webhook/trigger-transcripcion"

// TranscripcionHandler orquesta: frontend -> backend -> n8n -> backend webhook -> PostgreSQL.
type TranscripcionHandler struct {
	DB            *gorm.DB
	N8NWebhookURL string
}

// Simular dispara el workflow de n8n. Con ?fallback=true procesa el CSV desde backend.
func (h *TranscripcionHandler) Simular(c *gin.Context) {
	if strings.EqualFold(c.Query("fallback"), "true") {
		h.runFallback(c)
		return
	}

	payload := map[string]string{
		"modo":         "simulacion",
		"source":       "frontend",
		"requested_by": "admin",
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		h.respondN8NUnavailable(c)
		return
	}

	n8nURL := strings.TrimSpace(h.N8NWebhookURL)
	if n8nURL == "" {
		n8nURL = defaultN8NWebhookURL
	}

	req, err := http.NewRequest(http.MethodPost, n8nURL, bytes.NewReader(bodyBytes))
	if err != nil {
		h.respondN8NUnavailable(c)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		h.respondN8NUnavailable(c)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 || looksLikeHTML(resp.Header.Get("Content-Type"), body) {
		h.respondN8NUnavailable(c)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Flujo de transcripción iniciado correctamente en n8n",
		"source":  "n8n",
		"estado":  "EN_PROCESO",
	})
}

func (h *TranscripcionHandler) respondN8NUnavailable(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": false,
		"message": "n8n no tiene activo el webhook trigger-transcripcion o no está disponible",
		"source":  "n8n",
		"estado":  "ERROR",
		"hint":    "Importa y activa n8n-workflow.json o usa fallback=true",
	})
}

func looksLikeHTML(contentType string, body []byte) bool {
	if strings.Contains(strings.ToLower(contentType), "text/html") {
		return true
	}
	trimmed := strings.TrimSpace(string(body))
	return strings.HasPrefix(trimmed, "<!DOCTYPE html") ||
		strings.HasPrefix(trimmed, "<html") ||
		strings.HasPrefix(trimmed, "<")
}

// runFallback procesa Transcripciones.csv directamente sin pasar por n8n.
func (h *TranscripcionHandler) runFallback(c *gin.Context) {
	f, err := os.Open(filepath.Join("data", "Transcripciones.csv"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "n8n no disponible y Transcripciones.csv tampoco está accesible",
			"source":  "backend_fallback",
			"estado":  "ERROR",
			"errores": 1,
		})
		return
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.FieldsPerRecord = -1
	if _, err := reader.Read(); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Error leyendo Transcripciones.csv en fallback",
			"source":  "backend_fallback",
			"estado":  "ERROR",
			"errores": 1,
		})
		return
	}

	actualizadas, observadas, errores := 0, 0, 0
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil || len(row) < 26 {
			errores++
			continue
		}

		codigoActa, err := strconv.ParseInt(strings.TrimSpace(row[8]), 10, 64)
		if err != nil {
			errores++
			continue
		}

		obs := strings.TrimSpace(row[20])
		estado := "transcrita"
		if obs != "" {
			estado = "observada"
		}

		updates := map[string]interface{}{
			"estado":              estado,
			"p1":                  parseCampo(row[13]),
			"p2":                  parseCampo(row[14]),
			"p3":                  parseCampo(row[15]),
			"p4":                  parseCampo(row[16]),
			"votos_validos":       parseCampo(row[17]),
			"votos_blanco":        parseCampo(row[18]),
			"votos_nulos":         parseCampo(row[19]),
			"papeletas_anfora":    parseCampo(row[11]),
			"papeletas_no_usadas": parseCampo(row[12]),
			"observaciones":       obs,
			"apertura_hora":       parseCampo(row[22]),
			"apertura_minutos":    parseCampo(row[23]),
			"cierre_hora":         parseCampo(row[24]),
			"cierre_minutos":      parseCampo(row[25]),
		}

		if err := h.DB.Model(&models.Acta{}).Where("codigo_acta = ?", codigoActa).Updates(updates).Error; err != nil {
			errores++
			continue
		}

		actualizadas++
		if estado == "observada" {
			observadas++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"message":      "n8n no disponible; se usó fallback backend para demo",
		"source":       "backend_fallback",
		"estado":       "COMPLETADO",
		"actualizadas": actualizadas,
		"observadas":   observadas,
		"errores":      errores,
	})
}
