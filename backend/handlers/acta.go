package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/srvof/votos-backend/middleware"
	"github.com/srvof/votos-backend/models"
	"gorm.io/gorm"
)

// ActaHandler CRUD actas y resumen.
type ActaHandler struct {
	DB *gorm.DB
}

func validateActa(a *models.Acta, cantHabilitada int) error {
	sumP := a.P1 + a.P2 + a.P3 + a.P4
	if sumP != a.VotosValidos {
		return errors.New("p1+p2+p3+p4 debe coincidir con votosValidos")
	}
	total := sumP + a.VotosNulos + a.VotosBlanco
	if total > cantHabilitada {
		return fmt.Errorf("suma de votos (%d) supera cantidad habilitada de la mesa (%d)", total, cantHabilitada)
	}
	if a.PapeletasNoUsadas < 0 || a.P1 < 0 || a.P2 < 0 || a.P3 < 0 || a.P4 < 0 || a.VotosNulos < 0 || a.VotosBlanco < 0 {
		return errors.New("los conteos no pueden ser negativos")
	}
	return nil
}

func (h *ActaHandler) List(c *gin.Context) {
	var list []models.Acta
	q := h.DB.Preload("Mesa").Preload("Mesa.Recinto")
	if mesaID := c.Query("mesaId"); mesaID != "" {
		q = q.Where("id_mesa = ?", mesaID)
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
	if err := h.DB.Preload("Mesa").Preload("Mesa.Recinto").First(&a, uint(id64)).Error; err != nil {
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
	var mesa models.Mesa
	if err := h.DB.First(&mesa, a.IDMesa).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "mesa no existe"})
		return
	}
	if err := validateActa(&a, mesa.CantidadHabilitada); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
	var mesa models.Mesa
	if err := h.DB.First(&mesa, a.IDMesa).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "mesa no existe"})
		return
	}
	if err := validateActa(&a, mesa.CantidadHabilitada); err != nil {
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

// ResumenPorRecinto filas agregadas por recinto.
type ResumenPorRecinto struct {
	RecintoID      uint   `gorm:"column:recinto_id" json:"recintoId"`
	RecintoNombre  string `gorm:"column:recinto_nombre" json:"recinto"`
	P1             int64  `gorm:"column:p1" json:"p1"`
	P2             int64  `gorm:"column:p2" json:"p2"`
	P3             int64  `gorm:"column:p3" json:"p3"`
	P4             int64  `gorm:"column:p4" json:"p4"`
	VotosNulos     int64  `gorm:"column:votos_nulos" json:"votosNulos"`
	VotosBlanco    int64  `gorm:"column:votos_blanco" json:"votosBlanco"`
	VotosValidos   int64  `gorm:"column:votos_validos" json:"votosValidos"`
	PapeletasNoUs  int64  `gorm:"column:papeletas_no_us" json:"papeletasNoUsadas"`
}

// ResumenPorDistribucion agregado por departamento/municipio/provincia.
type ResumenPorDistribucion struct {
	DistribucionID uint   `gorm:"column:distribucion_id" json:"idDistribucionTerritorial"`
	Departamento   string `gorm:"column:departamento" json:"departamento"`
	Municipio      string `gorm:"column:municipio" json:"municipio"`
	Provincia      string `gorm:"column:provincia" json:"provincia"`
	P1             int64  `gorm:"column:p1" json:"p1"`
	P2             int64  `gorm:"column:p2" json:"p2"`
	P3             int64  `gorm:"column:p3" json:"p3"`
	P4             int64  `gorm:"column:p4" json:"p4"`
	VotosNulos     int64  `gorm:"column:votos_nulos" json:"votosNulos"`
	VotosBlanco    int64  `gorm:"column:votos_blanco" json:"votosBlanco"`
	VotosValidos   int64  `gorm:"column:votos_validos" json:"votosValidos"`
	PapeletasNoUs  int64  `gorm:"column:papeletas_no_us" json:"papeletasNoUsadas"`
}

// ResumenRespuesta agrupa dos vistas.
type ResumenRespuesta struct {
	PorRecinto      []ResumenPorRecinto      `json:"porRecinto"`
	PorDistribucion []ResumenPorDistribucion `json:"porDistribucion"`
}

// Resumen totales agregados (actas no eliminadas).
func (h *ActaHandler) Resumen(c *gin.Context) {
	var porRecinto []ResumenPorRecinto
	sqlRecinto := `
SELECT r.recinto_id AS recinto_id, r.recinto AS recinto_nombre,
  COALESCE(SUM(a.p1),0) AS p1, COALESCE(SUM(a.p2),0) AS p2,
  COALESCE(SUM(a.p3),0) AS p3, COALESCE(SUM(a.p4),0) AS p4,
  COALESCE(SUM(a.votos_nulos),0) AS votos_nulos,
  COALESCE(SUM(a.votos_blanco),0) AS votos_blanco,
  COALESCE(SUM(a.votos_validos),0) AS votos_validos,
  COALESCE(SUM(a.papeletas_no_usadas),0) AS papeletas_no_us
FROM acta a
JOIN mesa m ON m.id = a.id_mesa
JOIN recinto_electoral r ON r.recinto_id = m.id_recinto_electoral
WHERE a.fecha_eliminado IS NULL
GROUP BY r.recinto_id, r.recinto
ORDER BY r.recinto_id
`
	if err := h.DB.Raw(sqlRecinto).Scan(&porRecinto).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var porDist []ResumenPorDistribucion
	sqlDist := `
SELECT d.id AS distribucion_id, d.departamento, d.municipio, d.provincia,
  COALESCE(SUM(a.p1),0) AS p1, COALESCE(SUM(a.p2),0) AS p2,
  COALESCE(SUM(a.p3),0) AS p3, COALESCE(SUM(a.p4),0) AS p4,
  COALESCE(SUM(a.votos_nulos),0) AS votos_nulos,
  COALESCE(SUM(a.votos_blanco),0) AS votos_blanco,
  COALESCE(SUM(a.votos_validos),0) AS votos_validos,
  COALESCE(SUM(a.papeletas_no_usadas),0) AS papeletas_no_us
FROM acta a
JOIN mesa m ON m.id = a.id_mesa
JOIN recinto_electoral r ON r.recinto_id = m.id_recinto_electoral
JOIN distribucion_territorial d ON d.id = r.id_distribucion_territorial
WHERE a.fecha_eliminado IS NULL
GROUP BY d.id, d.departamento, d.municipio, d.provincia
ORDER BY d.id
`
	if err := h.DB.Raw(sqlDist).Scan(&porDist).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ResumenRespuesta{
		PorRecinto:      porRecinto,
		PorDistribucion: porDist,
	})
}
