package dashboard

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

// RegisterRoutes registra todos los endpoints del dashboard bajo /api/v1/dashboard.
// Todos requieren JWT (el grupo `protected` ya lleva el middleware AuthJWT).
// Todos son GET (solo lectura — CQRS/Query).
func RegisterRoutes(protected *gin.RouterGroup, db *gorm.DB, mongoClient *mongo.Client, dbName string, rrvCollections []string, eventsCollection string) {
	h := NewHandler(db, mongoClient, dbName, rrvCollections, eventsCollection)

	d := protected.Group("/dashboard")
	d.GET("/kpis", h.KPIs)
	d.GET("/rrv-vs-oficial", h.RRVvsOficial)
	d.GET("/votos-candidato", h.VotosCandidato)
	d.GET("/participacion", h.Participacion)
	d.GET("/geografico", h.Geografico)
	d.GET("/tecnico", h.Tecnico)
	d.GET("/inconsistencias", h.Inconsistencias)
	d.GET("/eventos-rrv", h.EventosRRV)
	d.GET("/mobile-scans", h.MobileScans)
}
