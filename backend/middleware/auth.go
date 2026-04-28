package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/srvof/votos-backend/utils"
)

const (
	ContextUserID  = "userID"
	ContextEsAdmin = "esAdmin"
)

// AuthJWT valida Bearer token y guarda userID y esAdmin en el contexto.
func AuthJWT(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if h == "" || !strings.HasPrefix(h, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token requerido"})
			return
		}
		raw := strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
		claims, err := utils.ParseToken(raw, secret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token inválido"})
			return
		}
		c.Set(ContextUserID, claims.UserID)
		c.Set(ContextEsAdmin, claims.EsAdmin)
		c.Next()
	}
}

// RequireAdmin exige que el usuario sea administrador.
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !c.GetBool(ContextEsAdmin) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "requiere rol administrador"})
			return
		}
		c.Next()
	}
}
