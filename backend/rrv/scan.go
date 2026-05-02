package rrv

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// ScanHandler guarda actas OCR/scanner directamente en el flujo RRV (MongoDB).
type ScanHandler struct {
	MongoClient *mongo.Client
	DBName      string
	Collection  string // "actas_rrv"
	LogPath     string // ruta para observadas.log
}

type candidatoScan struct {
	CandidatoID string `bson:"candidato_id" json:"candidato_id"`
	Nombre      string `bson:"nombre"       json:"nombre"`
	Votos       int    `bson:"votos"        json:"votos"`
}

type actaScanDoc struct {
	ActaID         string          `bson:"acta_id"`
	Mesa           string          `bson:"mesa"`
	Candidatos     []candidatoScan `bson:"candidatos"`
	VotosNulos     int             `bson:"votos_nulos"`
	VotosBlancos   int             `bson:"votos_blancos"`
	TotalVotos     int             `bson:"total_votos"`
	Estado         string          `bson:"estado"`
	Fuente         string          `bson:"fuente"`
	CodigoMesa     int64           `bson:"codigo_mesa"`
	Habilitados    int             `bson:"habilitados"`
	Papeletas      int             `bson:"papeletas"`
	ImagenURL      string          `bson:"imagen_url"`
	FuenteScan     string          `bson:"fuente_scan"`
	FechaRecepcion time.Time       `bson:"fecha_recepcion"`
	FechaProcesado time.Time       `bson:"fecha_procesado"`
}

// SaveScan guarda un acta escaneada (OCR) en MongoDB actas_rrv (flujo RRV / conteo rápido).
// POST /api/v1/rrv/scan
func (h *ScanHandler) SaveScan(c *gin.Context) {
	var req struct {
		CodigoMesa  int64  `json:"codigo_mesa"`
		NroMesa     int    `json:"numero_mesa"`
		P1          int    `json:"p1"`
		P2          int    `json:"p2"`
		P3          int    `json:"p3"`
		P4          int    `json:"p4"`
		Partido1    string `json:"partido_1"`
		Partido2    string `json:"partido_2"`
		Partido3    string `json:"partido_3"`
		Partido4    string `json:"partido_4"`
		Validos     int    `json:"validos"`
		Blancos     int    `json:"blancos"`
		Nulos       int    `json:"nulos"`
		Habilitados int    `json:"habilitados"`
		Papeletas   int    `json:"papeletas"`
		ImagenURL   string `json:"imagen_url"`
		FuenteScan  string `json:"fuente_scan"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payload inválido: " + err.Error()})
		return
	}

	now := time.Now()
	actaID := fmt.Sprintf("SCAN-%d-%d", req.CodigoMesa, now.UnixMilli())
	calculado := req.P1 + req.P2 + req.P3 + req.P4 + req.Blancos + req.Nulos
	estado := "validada"
	if req.Validos > 0 && calculado != req.Validos {
		estado = "pendiente_revision"
	}

	fuente := req.FuenteScan
	if fuente == "" {
		fuente = "web_scanner"
	}

	partido1 := strings.TrimSpace(req.Partido1)
	if partido1 == "" {
		partido1 = "Partido 1"
	}
	partido2 := strings.TrimSpace(req.Partido2)
	if partido2 == "" {
		partido2 = "Partido 2"
	}
	partido3 := strings.TrimSpace(req.Partido3)
	if partido3 == "" {
		partido3 = "Partido 3"
	}
	partido4 := strings.TrimSpace(req.Partido4)
	if partido4 == "" {
		partido4 = "Partido 4"
	}

	acta := actaScanDoc{
		ActaID: actaID,
		Mesa:   fmt.Sprintf("%d", req.CodigoMesa),
		Candidatos: []candidatoScan{
			{CandidatoID: "P1", Nombre: partido1, Votos: req.P1},
			{CandidatoID: "P2", Nombre: partido2, Votos: req.P2},
			{CandidatoID: "P3", Nombre: partido3, Votos: req.P3},
			{CandidatoID: "P4", Nombre: partido4, Votos: req.P4},
		},
		VotosNulos:     req.Nulos,
		VotosBlancos:   req.Blancos,
		TotalVotos:     req.Validos + req.Blancos + req.Nulos,
		Estado:         estado,
		Fuente:         "SCANNER",
		CodigoMesa:     req.CodigoMesa,
		Habilitados:    req.Habilitados,
		Papeletas:      req.Papeletas,
		ImagenURL:      req.ImagenURL,
		FuenteScan:     fuente,
		FechaRecepcion: now,
		FechaProcesado: now,
	}

	col := h.MongoClient.Database(h.DBName).Collection(h.Collection)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Validación de duplicados: no se permite más de un acta por mesa en RRV.
	var existing actaScanDoc
	if err := col.FindOne(ctx, bson.M{"mesa": acta.Mesa}).Decode(&existing); err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"error":   "Acta duplicada: ya existe un acta registrada para esta mesa en RRV",
			"acta_id": existing.ActaID,
			"estado":  existing.Estado,
			"mesa":    existing.Mesa,
		})
		return
	}

	if _, err := col.InsertOne(ctx, acta); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error guardando en MongoDB: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "acta_id": actaID, "estado": estado})
}

// SaveRechazada registra un acta rechazada (con razón) en el log del proyecto.
// POST /api/v1/rrv/scan/rechazada
func (h *ScanHandler) SaveRechazada(c *gin.Context) {
	var req struct {
		CodigoMesa  int64  `json:"codigo_mesa"`
		NroMesa     int    `json:"numero_mesa"`
		P1          int    `json:"p1"`
		P2          int    `json:"p2"`
		P3          int    `json:"p3"`
		P4          int    `json:"p4"`
		Partido1    string `json:"partido_1"`
		Partido2    string `json:"partido_2"`
		Partido3    string `json:"partido_3"`
		Partido4    string `json:"partido_4"`
		Validos     int    `json:"validos"`
		Blancos     int    `json:"blancos"`
		Nulos       int    `json:"nulos"`
		Habilitados int    `json:"habilitados"`
		Papeletas   int    `json:"papeletas"`
		ImagenURL   string `json:"imagen_url"`
		FuenteScan  string `json:"fuente_scan"`
		Razon       string `json:"razon"`
		Observaciones []struct {
			Tipo        string `json:"tipo"`
			Descripcion string `json:"descripcion"`
		} `json:"observaciones_detectadas"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logDir := h.LogPath
	if logDir == "" {
		logDir = "data/logs"
	}
	if err := os.MkdirAll(logDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error creando directorio de logs: " + err.Error()})
		return
	}

	f, err := os.OpenFile(logDir+"/rechazadas.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error abriendo log: " + err.Error()})
		return
	}
	defer f.Close()

	var obsTexts []string
	for _, o := range req.Observaciones {
		obsTexts = append(obsTexts, fmt.Sprintf("[%s] %s", o.Tipo, o.Descripcion))
	}
	obsStr := strings.Join(obsTexts, " | ")
	if obsStr == "" {
		obsStr = "(sin observaciones)"
	}

	fuente := req.FuenteScan
	if fuente == "" {
		fuente = "web_scanner"
	}

	line := fmt.Sprintf("[%s] RECHAZADA | mesa=%d | numero_mesa=%d | fuente=%s | razon=%q | p1=%d (%s) | p2=%d (%s) | p3=%d (%s) | p4=%d (%s) | validos=%d | blancos=%d | nulos=%d | habilitados=%d | papeletas=%d | observaciones=%s | imagen=%s\n",
		time.Now().Format(time.RFC3339),
		req.CodigoMesa,
		req.NroMesa,
		fuente,
		req.Razon,
		req.P1, req.Partido1,
		req.P2, req.Partido2,
		req.P3, req.Partido3,
		req.P4, req.Partido4,
		req.Validos, req.Blancos, req.Nulos,
		req.Habilitados, req.Papeletas,
		obsStr,
		req.ImagenURL,
	)
	if _, err := f.WriteString(line); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error escribiendo log: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "logged": true})
}

// CheckDuplicate verifica si ya existe un acta para el codigo_mesa dado.
// GET /api/v1/rrv/scan/check?mesa=1030400115004
func (h *ScanHandler) CheckDuplicate(c *gin.Context) {
	mesaStr := strings.TrimSpace(c.Query("mesa"))
	if mesaStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "se requiere parámetro mesa"})
		return
	}

	col := h.MongoClient.Database(h.DBName).Collection(h.Collection)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var existing actaScanDoc
	err := col.FindOne(ctx, bson.M{"mesa": mesaStr}).Decode(&existing)
	if err == nil {
		c.JSON(http.StatusOK, gin.H{
			"existe":    true,
			"duplicada": true,
			"acta_id":   existing.ActaID,
			"fuente":    existing.Fuente,
			"fecha":     existing.FechaRecepcion,
			"estado":    existing.Estado,
			"mesa":      existing.Mesa,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"existe": false, "duplicada": false})
}

// SaveObservada registra un acta observada en el log de texto del proyecto.
// POST /api/v1/rrv/scan/observada
func (h *ScanHandler) SaveObservada(c *gin.Context) {
	var req struct {
		CodigoMesa   int64  `json:"codigo_mesa"`
		NumeroMesa   int    `json:"numero_mesa"`
		P1           int    `json:"p1"`
		P2           int    `json:"p2"`
		P3           int    `json:"p3"`
		P4           int    `json:"p4"`
		Validos      int    `json:"validos"`
		Blancos      int    `json:"blancos"`
		Nulos        int    `json:"nulos"`
		Habilitados  int    `json:"habilitados"`
		Papeletas    int    `json:"papeletas"`
		Partido1     string `json:"partido_1"`
		Partido2     string `json:"partido_2"`
		Partido3     string `json:"partido_3"`
		Partido4     string `json:"partido_4"`
		FuenteScan   string `json:"fuente_scan"`
		Anulada      bool   `json:"anulada"`
		RazonPrincipal string `json:"razon_principal"`
		ImagenURL    string `json:"imagen_url"`
		Observaciones []struct {
			Tipo        string `json:"tipo"`
			Descripcion string `json:"descripcion"`
		} `json:"observaciones_detectadas"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logDir := h.LogPath
	if logDir == "" {
		logDir = "data/logs"
	}
	if err := os.MkdirAll(logDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error creando directorio de logs: " + err.Error()})
		return
	}

	f, err := os.OpenFile(logDir+"/observadas.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error abriendo log: " + err.Error()})
		return
	}
	defer f.Close()

	var obsTexts []string
	for _, o := range req.Observaciones {
		obsTexts = append(obsTexts, fmt.Sprintf("[%s] %s", o.Tipo, o.Descripcion))
	}
	obsStr := strings.Join(obsTexts, " | ")
	if obsStr == "" {
		obsStr = "(sin detalle)"
	}

	fuente := req.FuenteScan
	if fuente == "" {
		fuente = "web_scanner"
	}

	line := fmt.Sprintf("[%s] OBSERVADA | mesa=%d | numero_mesa=%d | fuente=%s | anulada=%v | razon=%q | p1=%d (%s) | p2=%d (%s) | p3=%d (%s) | p4=%d (%s) | validos=%d | blancos=%d | nulos=%d | habilitados=%d | papeletas=%d | observaciones=%s | imagen=%s\n",
		time.Now().Format(time.RFC3339),
		req.CodigoMesa,
		req.NumeroMesa,
		fuente,
		req.Anulada,
		req.RazonPrincipal,
		req.P1, req.Partido1,
		req.P2, req.Partido2,
		req.P3, req.Partido3,
		req.P4, req.Partido4,
		req.Validos, req.Blancos, req.Nulos,
		req.Habilitados, req.Papeletas,
		obsStr,
		req.ImagenURL,
	)
	if _, err := f.WriteString(line); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error escribiendo log: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "logged": true})
}
