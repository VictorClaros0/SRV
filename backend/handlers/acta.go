package handlers

import (
	"encoding/csv"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/srvof/votos-backend/middleware"
	"github.com/srvof/votos-backend/models"
	"gorm.io/gorm"
)

func parseCampo(s string) int {
	n, _ := strconv.Atoi(strings.TrimSpace(s))
	return n
}

// ActaHandler CRUD actas y resumen.
type ActaHandler struct {
	DB *gorm.DB
}

func validateActa(a *models.Acta) error {
	if a.P1 < 0 || a.P2 < 0 || a.P3 < 0 || a.P4 < 0 || a.VotosNulos < 0 || a.VotosBlanco < 0 || a.VotosValidos < 0 {
		return errors.New("los conteos no pueden ser negativos")
	}
	if a.VotosValidos > 0 {
		sumP := a.P1 + a.P2 + a.P3 + a.P4
		if sumP != a.VotosValidos {
			return errors.New("p1+p2+p3+p4 debe coincidir con votosValidos")
		}
	}
	return nil
}

func (h *ActaHandler) List(c *gin.Context) {
	var list []models.Acta
	q := h.DB
	if search := c.Query("search"); search != "" {
		like := "%" + search + "%"
		q = q.Where("CAST(codigo_acta AS TEXT) LIKE ? OR CAST(codigo_recinto AS TEXT) LIKE ?", like, like)
	}
	if err := q.Order("id").Find(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *ActaHandler) Get(c *gin.Context) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}
	var a models.Acta
	if err := h.DB.First(&a, uint(id64)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no encontrado"})
		return
	}
	c.JSON(http.StatusOK, a)
}

func (h *ActaHandler) Create(c *gin.Context) {
	var a models.Acta
	if err := c.ShouldBindJSON(&a); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validateActa(&a); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Verificar duplicado por codigoActa
	var existing models.Acta
	if err := h.DB.Where("codigo_acta = ?", a.CodigoActa).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Ya existe un acta con ese código. No se permiten actas duplicadas."})
		return
	}
	uid, ok := middleware.UserIDFromContext(c)
	if ok {
		a.CreadoPorID = &uid
	}
	a.ID = 0
	if err := h.DB.Create(&a).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, a)
}

func (h *ActaHandler) Update(c *gin.Context) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}
	var a models.Acta
	if err := h.DB.First(&a, uint(id64)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no encontrado"})
		return
	}
	if err := c.ShouldBindJSON(&a); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	a.ID = uint(id64)
	if err := validateActa(&a); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	uid, ok := middleware.UserIDFromContext(c)
	if ok {
		a.ModificadoPorID = &uid
		now := time.Now()
		a.FechaModificacion = &now
	}
	if err := h.DB.Save(&a).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, a)
}

func (h *ActaHandler) Delete(c *gin.Context) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}
	uid, ok := middleware.UserIDFromContext(c)
	if ok {
		_ = h.DB.Model(&models.Acta{}).Where("id = ?", uint(id64)).Update("eliminado_por", uid).Error
	}
	if err := h.DB.Delete(&models.Acta{}, uint(id64)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// WebhookN8N recibe el resultado de transcripción desde n8n y actualiza el acta.
// Si el campo observaciones viene con contenido, el acta se marca como "observada"
// (no cuenta en los resultados oficiales). De lo contrario pasa a "transcrita".
func (h *ActaHandler) WebhookN8N(c *gin.Context) {
	var payload struct {
		CodigoActa        int64  `json:"codigoActa" binding:"required"`
		P1                int    `json:"p1"`
		P2                int    `json:"p2"`
		P3                int    `json:"p3"`
		P4                int    `json:"p4"`
		VotosValidos      int    `json:"votosValidos"`
		VotosNulos        int    `json:"votosNulos"`
		VotosBlanco       int    `json:"votosBlanco"`
		PapeletasAnfora   int    `json:"papeletasAnfora"`
		PapeletasNoUsadas int    `json:"papeletasNoUsadas"`
		Observaciones     string `json:"observaciones"`
		AperturaHora      int    `json:"aperturaHora"`
		AperturaMinutos   int    `json:"aperturaMinutos"`
		CierreHora        int    `json:"cierreHora"`
		CierreMinutos     int    `json:"cierreMinutos"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var acta models.Acta
	if err := h.DB.Where("codigo_acta = ?", payload.CodigoActa).First(&acta).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "acta no encontrada"})
		return
	}
	estado := "transcrita"
	if strings.TrimSpace(payload.Observaciones) != "" {
		estado = "observada"
	}
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
	if err := h.DB.Model(&acta).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "id": acta.ID, "estado": estado})
}

// ProcesarTranscripciones lee Transcripciones.csv y actualiza todas las actas con sus datos reales.
func (h *ActaHandler) ProcesarTranscripciones(c *gin.Context) {
	f, err := os.Open(filepath.Join("data", "Transcripciones.csv"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transcripciones.csv no disponible"})
		return
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.FieldsPerRecord = -1
	if _, err := r.Read(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error leyendo CSV"})
		return
	}

	updated, observadas, errors := 0, 0, 0
	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil || len(row) < 26 {
			continue
		}
		codigoActa, err := strconv.ParseInt(strings.TrimSpace(row[8]), 10, 64)
		if err != nil {
			continue
		}
		obs := strings.TrimSpace(row[20])
		estado := "transcrita"
		if obs != "" {
			estado = "observada"
			observadas++
		}
		updates := map[string]interface{}{
			"estado":               estado,
			"p1":                   parseCampo(row[13]),
			"p2":                   parseCampo(row[14]),
			"p3":                   parseCampo(row[15]),
			"p4":                   parseCampo(row[16]),
			"votos_validos":        parseCampo(row[17]),
			"votos_blanco":         parseCampo(row[18]),
			"votos_nulos":          parseCampo(row[19]),
			"papeletas_anfora":     parseCampo(row[11]),
			"papeletas_no_usadas":  parseCampo(row[12]),
			"observaciones":        obs,
			"apertura_hora":        parseCampo(row[22]),
			"apertura_minutos":     parseCampo(row[23]),
			"cierre_hora":          parseCampo(row[24]),
			"cierre_minutos":       parseCampo(row[25]),
		}
		if err := h.DB.Model(&models.Acta{}).Where("codigo_acta = ?", codigoActa).Updates(updates).Error; err != nil {
			errors++
		} else {
			updated++
		}
	}
	c.JSON(http.StatusOK, gin.H{"actualizadas": updated, "observadas": observadas, "errores": errors})
}

// ParaTranscribir devuelve los datos del CSV de Transcripciones como JSON.
// Es consumido por el workflow de n8n para obtener los datos a transcribir.
// Endpoint público (sin JWT) ya que lo llama n8n internamente.
func (h *ActaHandler) ParaTranscribir(c *gin.Context) {
	f, err := os.Open(filepath.Join("data", "Transcripciones.csv"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "archivo Transcripciones.csv no disponible"})
		return
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.FieldsPerRecord = -1
	if _, err := r.Read(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error leyendo CSV"})
		return
	}

	type transcripcionPayload struct {
		CodigoActa        int64  `json:"codigoActa"`
		P1                int    `json:"p1"`
		P2                int    `json:"p2"`
		P3                int    `json:"p3"`
		P4                int    `json:"p4"`
		VotosValidos      int    `json:"votosValidos"`
		VotosBlanco       int    `json:"votosBlanco"`
		VotosNulos        int    `json:"votosNulos"`
		PapeletasAnfora   int    `json:"papeletasAnfora"`
		PapeletasNoUsadas int    `json:"papeletasNoUsadas"`
		Observaciones     string `json:"observaciones"`
		AperturaHora      int    `json:"aperturaHora"`
		AperturaMinutos   int    `json:"aperturaMinutos"`
		CierreHora        int    `json:"cierreHora"`
		CierreMinutos     int    `json:"cierreMinutos"`
	}

	var result []transcripcionPayload
	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil || len(row) < 26 {
			continue
		}
		codigoActa, err := strconv.ParseInt(strings.TrimSpace(row[8]), 10, 64)
		if err != nil {
			continue
		}
		result = append(result, transcripcionPayload{
			CodigoActa:        codigoActa,
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
		})
	}
	c.JSON(http.StatusOK, result)
}

// ResumenPorRecinto filas agregadas por recinto.
type ResumenPorRecinto struct {
	CodigoRecinto int64  `gorm:"column:codigo_recinto" json:"codigoRecinto"`
	P1            int64  `gorm:"column:p1" json:"p1"`
	P2            int64  `gorm:"column:p2" json:"p2"`
	P3            int64  `gorm:"column:p3" json:"p3"`
	P4            int64  `gorm:"column:p4" json:"p4"`
	VotosNulos    int64  `gorm:"column:votos_nulos" json:"votosNulos"`
	VotosBlanco   int64  `gorm:"column:votos_blanco" json:"votosBlanco"`
	VotosValidos  int64  `gorm:"column:votos_validos" json:"votosValidos"`
	TotalActas    int64  `gorm:"column:total_actas" json:"totalActas"`
}

// Resumen totales agregados por código de recinto.
func (h *ActaHandler) Resumen(c *gin.Context) {
	var resultado []ResumenPorRecinto
	sql := `
SELECT codigo_recinto,
  COALESCE(SUM(p1),0) AS p1, COALESCE(SUM(p2),0) AS p2,
  COALESCE(SUM(p3),0) AS p3, COALESCE(SUM(p4),0) AS p4,
  COALESCE(SUM(votos_nulos),0) AS votos_nulos,
  COALESCE(SUM(votos_blanco),0) AS votos_blanco,
  COALESCE(SUM(votos_validos),0) AS votos_validos,
  COUNT(*) AS total_actas
FROM acta
WHERE fecha_eliminado IS NULL
GROUP BY codigo_recinto
ORDER BY codigo_recinto
`
	if err := h.DB.Raw(sql).Scan(&resultado).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resultado)
}
