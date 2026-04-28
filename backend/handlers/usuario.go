package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/srvof/votos-backend/middleware"
	"github.com/srvof/votos-backend/models"
	"github.com/srvof/votos-backend/utils"
	"gorm.io/gorm"
)

// UsuarioHandler CRUD de usuarios.
type UsuarioHandler struct {
	DB *gorm.DB
}

// List lista todos los usuarios (solo admin en rutas).
func (h *UsuarioHandler) List(c *gin.Context) {
	var list []models.Usuario
	if err := h.DB.Omit("contrasena").Find(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

// Get obtiene un usuario por id (admin o el mismo usuario).
func (h *UsuarioHandler) Get(c *gin.Context) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}
	selfID, _ := middleware.UserIDFromContext(c)
	isAdmin := c.GetBool(middleware.ContextEsAdmin)
	if !isAdmin && uint(id64) != selfID {
		c.JSON(http.StatusForbidden, gin.H{"error": "no autorizado"})
		return
	}
	var u models.Usuario
	if err := h.DB.Omit("contrasena").First(&u, uint(id64)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no encontrado"})
		return
	}
	c.JSON(http.StatusOK, u)
}

type updateUsuarioBody struct {
	Nombre          *string `json:"nombre"`
	Apellido        *string `json:"apellido"`
	SegundoApellido *string `json:"segundoApellido"`
	Contrasena      *string `json:"contrasena"`
	EsAdmin         *bool   `json:"esAdmin"`
}

// Update actualiza usuario (admin o self; solo admin puede cambiar EsAdmin).
func (h *UsuarioHandler) Update(c *gin.Context) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}
	id := uint(id64)
	selfID, _ := middleware.UserIDFromContext(c)
	isAdmin := c.GetBool(middleware.ContextEsAdmin)
	if !isAdmin && id != selfID {
		c.JSON(http.StatusForbidden, gin.H{"error": "no autorizado"})
		return
	}
	var body updateUsuarioBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var u models.Usuario
	if err := h.DB.First(&u, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no encontrado"})
		return
	}
	if body.Nombre != nil {
		u.Nombre = *body.Nombre
	}
	if body.Apellido != nil {
		u.Apellido = *body.Apellido
	}
	if body.SegundoApellido != nil {
		u.SegundoApellido = *body.SegundoApellido
	}
	if body.Contrasena != nil {
		hash, err := utils.HashPassword(*body.Contrasena)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "hash"})
			return
		}
		u.Contrasena = hash
	}
	if isAdmin && body.EsAdmin != nil {
		u.EsAdmin = *body.EsAdmin
	}
	now := time.Now()
	u.FechaModificacion = &now
	u.ModificadoPorID = &selfID
	if err := h.DB.Save(&u).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	u.Contrasena = ""
	c.JSON(http.StatusOK, u)
}

// Delete borrado lógico (solo admin).
func (h *UsuarioHandler) Delete(c *gin.Context) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}
	actorID, _ := middleware.UserIDFromContext(c)
	if err := h.DB.Model(&models.Usuario{}).Where("id = ?", uint(id64)).Update("eliminado_por", actorID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := h.DB.Delete(&models.Usuario{}, uint(id64)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
