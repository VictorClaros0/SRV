package middleware

import (
	"github.com/gin-gonic/gin"
)

// UserIDFromContext devuelve el id de usuario del JWT o false.
func UserIDFromContext(c *gin.Context) (uint, bool) {
	v, ok := c.Get(ContextUserID)
	if !ok {
		return 0, false
	}
	id, ok := v.(uint)
	return id, ok
}
