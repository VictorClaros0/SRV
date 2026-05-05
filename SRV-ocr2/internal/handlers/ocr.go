package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"regexp"
	"srrv/internal/models"
	"srrv/internal/repository"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type OCRHandler struct {
	actaRepo *repository.ActaRepository
}

func NewOCRHandler(actaRepo *repository.ActaRepository) *OCRHandler {
	return &OCRHandler{actaRepo: actaRepo}
}

type OCRProcessRequest struct {
	PDFPath string `json:"pdf_path"`
}

type OCRProcessResponse struct {
	Status  string      `json:"status"`
	Data    models.Acta `json:"data"`
	Message string      `json:"message"`
}

// ScanAll dispara el escaneo masivo usando Goroutines y un Semáforo de Control
func (h *OCRHandler) ScanAll(c *gin.Context) {
	pdfDir := "/app/data/PdfOutput"
	files, err := ioutil.ReadDir(pdfDir)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("No se pudo leer el directorio: %v", err)})
		return
	}

	var pdfFiles []string
	for _, f := range files {
		if !f.IsDir() && filepath.Ext(f.Name()) == ".pdf" {
			pdfFiles = append(pdfFiles, filepath.Join(pdfDir, f.Name()))
		}
	}

	if len(pdfFiles) == 0 {
		c.JSON(200, gin.H{"message": "No se encontraron PDFs", "count": 0})
		return
	}

	// Iniciamos el proceso en segundo plano para no bloquear la respuesta HTTP
	go h.processBatch(pdfFiles)

	c.JSON(200, gin.H{
		"message": "Escaneo masivo iniciado 'Optimizado'",
		"count":   len(pdfFiles),
		"note":    "Usando Goroutines con Semáforo (Pool de Trabajadores)",
	})
}

func (h *OCRHandler) processBatch(files []string) {
	var wg sync.WaitGroup
	// SEMAFORO: Limitamos a 15 OCRs simultáneos (Modo MEGA-Nitro).
	semaphore := make(chan struct{}, 15)

	startTime := time.Now()
	log.Printf("🚀 Iniciando procesamiento batch de %d archivos...", len(files))

	for _, file := range files {
		wg.Add(1)
		go func(path string) {
			defer wg.Done()

			// Esperar turno en el semáforo
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			h.processSingleFile(path)
		}(file)
	}

	wg.Wait()
	log.Printf("✅ Procesamiento batch completado en %v", time.Since(startTime))
}

func (h *OCRHandler) processSingleFile(path string) {
	filename := filepath.Base(path)

	// Extraer código de mesa del nombre del archivo (acta_XXXXXXXXXXXXX.pdf)
	re := regexp.MustCompile(`acta_(\d+)\.pdf`)
	match := re.FindStringSubmatch(filename)
	if len(match) < 2 {
		log.Printf("⚠️ Nombre de archivo no válido: %s", filename)
		return
	}
	codigoMesa := match[1]

	// Llamar al microservicio de Python para el OCR pesado
	// El servicio de Python ahora está optimizado con resizing y sin Celery
	pythonURL := "http://ocr-api:8000/api/ocr/process"
	reqBody, _ := json.Marshal(OCRProcessRequest{PDFPath: path})

	client := &http.Client{
		Timeout: 60 * time.Second,
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
	}
	resp, err := client.Post(pythonURL, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		log.Printf("❌ Error llamando a Python OCR para %s: %v", filename, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("❌ Error en microservicio OCR (%d) para %s", resp.StatusCode, filename)
		return
	}

	var ocrResp OCRProcessResponse
	if err := json.NewDecoder(resp.Body).Decode(&ocrResp); err != nil {
		log.Printf("❌ Error decodificando respuesta para %s: %v", filename, err)
		return
	}

	if ocrResp.Status != "success" {
		log.Printf("❌ Fallo en OCR para %s: %s", filename, ocrResp.Message)
		return
	}

	// Completar datos faltantes
	acta := ocrResp.Data
	acta.CodigoMesa = codigoMesa

	// Intentar guardar en la base de datos
	if _, err := h.actaRepo.Create(context.Background(), &acta); err != nil {
		log.Printf("❌ Error guardando acta %s en DB: %v", codigoMesa, err)
		return
	}

	log.Printf("✨ Acta %s procesada y guardada.", codigoMesa)
}
