package comparacion

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

// RegisterRoutes registra los endpoints del módulo de comparación.
// Todos son GET (solo lectura — patrón CQRS/Query).
func RegisterRoutes(protected *gin.RouterGroup, db *gorm.DB, mongoClient *mongo.Client, dbName string) {
	h := newHandler(db, mongoClient, dbName)

	cmp := protected.Group("/comparacion")
	cmp.GET("", h.Comparar)
	cmp.GET("/resumen", h.Resumen)
	cmp.GET("/inconsistencias", h.Inconsistencias)
	cmp.GET("/:acta_id", h.GetActaComparada)
}
