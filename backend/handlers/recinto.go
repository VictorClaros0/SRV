package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/srvof/votos-backend/models"
	"gorm.io/gorm"
)

// RecintoHandler CRUD recintos electorales.
type RecintoHandler struct {
	DB *gorm.DB
}

func (h *RecintoHandler) List(c *gin.Context) {
	var list []models.RecintoElectoral
	q := h.DB.Preload("Distribucion")
	if distID := c.Query("distribucionId"); distID != "" {
		q = q.Where("id_distribucion_territorial = ?", distID)
	}
	if err := q.Order("recinto_id").Find(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *RecintoHandler) Get(c *gin.Context) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}
	var r models.RecintoElectoral
	if err := h.DB.Preload("Distribucion").First(&r, uint(id64)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no encontrado"})
		return
	}
	c.JSON(http.StatusOK, r)
}

func (h *RecintoHandler) Create(c *gin.Context) {
	var r models.RecintoElectoral
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	r.RecintoID = 0
	if err := h.DB.Create(&r).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, r)
}

func (h *RecintoHandler) Update(c *gin.Context) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}
	var r models.RecintoElectoral
	if err := h.DB.First(&r, uint(id64)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no encontrado"})
		return
	}
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	r.RecintoID = uint(id64)
	if err := h.DB.Save(&r).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, r)
}

func (h *RecintoHandler) Delete(c *gin.Context) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}
	if err := h.DB.Delete(&models.RecintoElectoral{}, uint(id64)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
