package handlers

import (
	"net/http"

	"srrv/internal/models"
	"srrv/internal/repository"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// DistribucionHandler agrupa los handlers HTTP para DistribucionTerritorial.
type DistribucionHandler struct {
	repo *repository.DistribucionRepository
}

// NewDistribucionHandler crea un nuevo handler.
func NewDistribucionHandler(repo *repository.DistribucionRepository) *DistribucionHandler {
	return &DistribucionHandler{repo: repo}
}

// GetAll godoc — GET /api/v1/distribuciones
func (h *DistribucionHandler) GetAll(c *gin.Context) {
	items, err := h.repo.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

// GetByID godoc — GET /api/v1/distribuciones/:id
func (h *DistribucionHandler) GetByID(c *gin.Context) {
	id, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	item, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Distribución territorial no encontrada"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, item)
}

// Create godoc — POST /api/v1/distribuciones
func (h *DistribucionHandler) Create(c *gin.Context) {
	var body models.DistribucionTerritorial
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	created, err := h.repo.Create(c.Request.Context(), &body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, created)
}

// Update godoc — PUT /api/v1/distribuciones/:id
func (h *DistribucionHandler) Update(c *gin.Context) {
	id, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var body models.DistribucionTerritorial
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := h.repo.Update(c.Request.Context(), id, &body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updated)
}

// Delete godoc — DELETE /api/v1/distribuciones/:id
func (h *DistribucionHandler) Delete(c *gin.Context) {
	id, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Distribución territorial eliminada"})
}
