package handlers

import (
	"fmt"
	"net/http"
	"time"

	"srrv/internal/models"
	"srrv/internal/repository"
	"srrv/internal/services"

	"github.com/gin-gonic/gin"
)

type ScanHandler struct {
	actaRepo   *repository.RRVActaRepository
	eventoRepo *repository.EventoRepository
}

func NewScanHandler(ar *repository.RRVActaRepository, er *repository.EventoRepository) *ScanHandler {
	return &ScanHandler{actaRepo: ar, eventoRepo: er}
}

// CheckMesa verifica si ya existe un acta para la mesa indicada.
// GET /api/v1/rrv/scan/check?mesa=X
func (h *ScanHandler) CheckMesa(c *gin.Context) {
	mesa := c.Query("mesa")
	if mesa == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "parámetro 'mesa' requerido"})
		return
	}
	existing, err := h.actaRepo.GetByMesa(c.Request.Context(), mesa)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existing == nil {
		c.JSON(http.StatusOK, gin.H{"existe": false})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"existe":   true,
		"mesa":     existing.Mesa,
		"acta_id":  existing.ActaID,
		"estado":   existing.Estado,
		"fuente":   existing.Fuente,
	})
}

// SaveScan guarda un acta escaneada y aceptada desde la app móvil.
// POST /api/v1/rrv/scan
func (h *ScanHandler) SaveScan(c *gin.Context) {
	var body struct {
		CodigoMesa int     `json:"codigo_mesa"`
		NumeroMesa int     `json:"numero_mesa"`
		P1         int     `json:"p1"`
		P2         int     `json:"p2"`
		P3         int     `json:"p3"`
		P4         int     `json:"p4"`
		Partido1   *string `json:"partido_1"`
		Partido2   *string `json:"partido_2"`
		Partido3   *string `json:"partido_3"`
		Partido4   *string `json:"partido_4"`
		Validos    int     `json:"validos"`
		Blancos    int     `json:"blancos"`
		Nulos      int     `json:"nulos"`
		Habilitados int    `json:"habilitados"`
		Papeletas  int     `json:"papeletas"`
		FuenteScan string  `json:"fuente_scan"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mesa := fmt.Sprintf("%d", body.CodigoMesa)
	actaID := fmt.Sprintf("SCAN-%s-%d", mesa, time.Now().UnixMilli())

	nombre := func(p *string, def string) string {
		if p != nil && *p != "" { return *p }
		return def
	}

	acta := &models.RRVActa{
		ActaID: actaID,
		Mesa:   mesa,
		Candidatos: []models.Candidato{
			{CandidatoID: "CAND-01", Nombre: nombre(body.Partido1, "Candidato 1"), Votos: body.P1},
			{CandidatoID: "CAND-02", Nombre: nombre(body.Partido2, "Candidato 2"), Votos: body.P2},
			{CandidatoID: "CAND-03", Nombre: nombre(body.Partido3, "Candidato 3"), Votos: body.P3},
			{CandidatoID: "CAND-04", Nombre: nombre(body.Partido4, "Candidato 4"), Votos: body.P4},
		},
		VotosNulos:     body.Nulos,
		VotosBlancos:   body.Blancos,
		TotalVotos:     body.Validos + body.Blancos + body.Nulos,
		Estado:         "PROCESADA",
		Fuente:         "SCAN_MOVIL",
		TipoEntrada:    "IMAGEN",
		HashOrigen:     services.HashBytes([]byte(actaID)),
		FechaRecepcion: time.Now(),
	}

	saved, err := h.actaRepo.Create(c.Request.Context(), acta)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, saved)
}

// SaveObservada guarda un acta observada o anulada desde la app móvil.
// POST /api/v1/rrv/scan/observada
func (h *ScanHandler) SaveObservada(c *gin.Context) {
	var body struct {
		CodigoMesa            int         `json:"codigo_mesa"`
		FuenteScan            string      `json:"fuente_scan"`
		Anulada               bool        `json:"anulada"`
		RazonPrincipal        string      `json:"razon_principal"`
		ObservacionesDetectadas interface{} `json:"observaciones_detectadas"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mesa := fmt.Sprintf("%d", body.CodigoMesa)
	estado := "OBSERVADA"
	if body.Anulada {
		estado = "ANULADA"
	}
	actaID := fmt.Sprintf("SCAN-%s-%d", mesa, time.Now().UnixMilli())

	acta := &models.RRVActa{
		ActaID:         actaID,
		Mesa:           mesa,
		Candidatos:     []models.Candidato{},
		Estado:         estado,
		Fuente:         "SCAN_MOVIL",
		TipoEntrada:    "IMAGEN",
		HashOrigen:     services.HashBytes([]byte(actaID)),
		FechaRecepcion: time.Now(),
	}

	saved, err := h.actaRepo.Create(c.Request.Context(), acta)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, saved)
}

// SaveRechazada guarda un acta rechazada desde la app móvil.
// POST /api/v1/rrv/scan/rechazada
func (h *ScanHandler) SaveRechazada(c *gin.Context) {
	var body struct {
		CodigoMesa            int         `json:"codigo_mesa"`
		P1, P2, P3, P4        int
		Validos, Blancos, Nulos int
		FuenteScan            string      `json:"fuente_scan"`
		Razon                 string      `json:"razon"`
		ObservacionesDetectadas interface{} `json:"observaciones_detectadas"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mesa := fmt.Sprintf("%d", body.CodigoMesa)
	actaID := fmt.Sprintf("SCAN-%s-%d", mesa, time.Now().UnixMilli())

	acta := &models.RRVActa{
		ActaID:         actaID,
		Mesa:           mesa,
		Candidatos:     []models.Candidato{},
		Estado:         "RECHAZADA",
		Fuente:         "SCAN_MOVIL",
		TipoEntrada:    "IMAGEN",
		HashOrigen:     services.HashBytes([]byte(actaID)),
		FechaRecepcion: time.Now(),
	}

	saved, err := h.actaRepo.Create(c.Request.Context(), acta)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, saved)
}
