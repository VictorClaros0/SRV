package handlers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type OCRVisionHandler struct {
	DB *gorm.DB
}

type claudeVisionReq struct {
	Model     string            `json:"model"`
	MaxTokens int               `json:"max_tokens"`
	Messages  []claudeVisionMsg `json:"messages"`
}

type claudeVisionMsg struct {
	Role    string        `json:"role"`
	Content []interface{} `json:"content"`
}

type claudeImgBlock struct {
	Type   string       `json:"type"`
	Source claudeImgSrc `json:"source"`
}

type claudeImgSrc struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

type claudeTxtBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type claudeAPIResp struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

const ocrPrompt = `Analiza esta acta electoral boliviana y extrae exactamente los siguientes campos.
Los votos de cada candidato están escritos en casillas separadas de 3 dígitos (ej: las casillas |0|9|5| representan el número 095 = 95).
Lee cada casilla con mucho cuidado, son números individuales en celdas separadas.
El código de mesa es el número largo en la esquina superior izquierda (ej: 1030400115004).

Devuelve ÚNICAMENTE un objeto JSON válido con estos campos (usa null para los que no puedas leer con certeza):
{
  "codigo_mesa": número del código de mesa (entero largo),
  "numero_mesa": número corto de mesa (entero),
  "p1": votos para Daenerys Targaryen (entero),
  "p2": votos para Sansa Stark (entero),
  "p3": votos para Robert Baratheon (entero),
  "p4": votos para Tyrion Lannister (entero),
  "validos": total votos válidos (entero),
  "blancos": votos blancos (entero),
  "nulos": votos nulos (entero),
  "habilitados": electores habilitados (entero),
  "papeletas": papeletas en ánfora (entero),
  "no_utilizadas": papeletas no utilizadas (entero),
  "apertura_h": hora de apertura (entero),
  "apertura_m": minutos de apertura (entero),
  "cierre_h": hora de cierre (entero),
  "cierre_m": minutos de cierre (entero),
  "anulada": true si el acta está anulada, false si no,
  "observaciones": texto de observaciones o cadena vacía
}`

// Process recibe una imagen del acta y la analiza con Claude Vision API.
func (h *OCRVisionHandler) Process(c *gin.Context) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "ANTHROPIC_API_KEY no configurada en el servidor"})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "campo 'file' requerido"})
		return
	}
	defer file.Close()

	if header.Size > 15*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "archivo demasiado grande (máx 15MB)"})
		return
	}

	imgData, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error leyendo archivo"})
		return
	}

	mediaType := "image/png"
	name := strings.ToLower(header.Filename)
	if strings.HasSuffix(name, ".jpg") || strings.HasSuffix(name, ".jpeg") {
		mediaType = "image/jpeg"
	} else if strings.HasSuffix(name, ".webp") {
		mediaType = "image/webp"
	}

	b64 := base64.StdEncoding.EncodeToString(imgData)

	payload := claudeVisionReq{
		Model:     "claude-haiku-4-5-20251001",
		MaxTokens: 1024,
		Messages: []claudeVisionMsg{
			{
				Role: "user",
				Content: []interface{}{
					claudeImgBlock{
						Type: "image",
						Source: claudeImgSrc{
							Type:      "base64",
							MediaType: mediaType,
							Data:      b64,
						},
					},
					claudeTxtBlock{
						Type: "text",
						Text: ocrPrompt,
					},
				},
			},
		},
	}

	reqBytes, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(reqBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "error contactando Claude API: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	var apiResp claudeAPIResp
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error decodificando respuesta de Claude"})
		return
	}

	if apiResp.Error != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Claude API: " + apiResp.Error.Message})
		return
	}

	if len(apiResp.Content) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "respuesta vacía de Claude"})
		return
	}

	text := apiResp.Content[0].Text
	// Strip markdown code fences if Claude wrapped the JSON
	if idx := strings.Index(text, "{"); idx >= 0 {
		if end := strings.LastIndex(text, "}"); end > idx {
			text = text[idx : end+1]
		}
	}

	// Return raw JSON as-is so the frontend receives it directly
	c.Header("Content-Type", "application/json")
	c.String(http.StatusOK, text)
}
