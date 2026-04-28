package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/srvof/votos-backend/config"
	"github.com/srvof/votos-backend/dto"
	"github.com/srvof/votos-backend/middleware"
	"github.com/srvof/votos-backend/models"
	"github.com/srvof/votos-backend/utils"
	"gorm.io/gorm"
)

const jwtTTL = 24 * time.Hour

// AuthHandler agrupa login y registro.
type AuthHandler struct {
	DB     *gorm.DB
	Config *config.Config
}

// Login emite JWT si las credenciales son válidas.
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var u models.Usuario
	if err := h.DB.Where("usuario = ?", req.Usuario).First(&u).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "credenciales inválidas"})
		return
	}
	if !utils.CheckPassword(u.Contrasena, req.Contrasena) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "credenciales inválidas"})
		return
	}
	token, err := utils.GenerateToken(u.ID, u.EsAdmin, h.Config.JWTSecret, jwtTTL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo generar token"})
		return
	}
	c.JSON(http.StatusOK, dto.LoginResponse{
		Token:     token,
		ExpiresIn: int(jwtTTL.Seconds()),
		EsAdmin:   u.EsAdmin,
	})
}

// Register crea un usuario (solo admin).
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	adminID, ok := middleware.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "sin usuario"})
		return
	}
	hash, err := utils.HashPassword(req.Contrasena)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "hash de contraseña"})
		return
	}
	u := models.Usuario{
		Nombre:          req.Nombre,
		Apellido:        req.Apellido,
		SegundoApellido: req.SegundoApellido,
		NombreUsuario:   req.Usuario,
		Contrasena:      hash,
		EsAdmin:         req.EsAdmin,
		CreadoPorID:     &adminID,
	}
	if err := h.DB.Create(&u).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "no se pudo crear usuario (¿duplicado?)"})
		return
	}
	u.Contrasena = ""
	c.JSON(http.StatusCreated, u)
}

// Me devuelve el usuario autenticado.
func (h *AuthHandler) Me(c *gin.Context) {
	id, ok := middleware.UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "sin usuario"})
		return
	}
	var u models.Usuario
	if err := h.DB.First(&u, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "usuario no encontrado"})
		return
	}
	u.Contrasena = ""
	c.JSON(http.StatusOK, u)
}
