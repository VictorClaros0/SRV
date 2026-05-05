package sms

import (
	"github.com/gin-gonic/gin"
	"github.com/srvof/votos-backend/config"
	"go.mongodb.org/mongo-driver/mongo"
)

// RegisterRoutes registra los endpoints del módulo SMS.
//
// Rutas públicas (sin JWT):
//   POST /api/v1/rrv/sms/inbound   — webhook de SMS Forwarder
//
// Rutas protegidas (JWT requerido):
//   GET    /api/v1/rrv/sms/numeros          — listar números autorizados
//   POST   /api/v1/rrv/sms/numeros          — agregar número autorizado
//   DELETE /api/v1/rrv/sms/numeros/:numero  — quitar número autorizado
func RegisterRoutes(r *gin.Engine, protected *gin.RouterGroup, mongoClient *mongo.Client, cfg *config.Config) {
	store := NewStore(
		mongoClient,
		cfg.MongoDBName,
		firstCollection(cfg.MongoRRVCollections),
		cfg.MongoEventsCollection,
		cfg.SMSConfigCollection,
		cfg.SMSDefaultNumbers,
	)
	h := NewHandler(store, cfg.OCR2SMSForwardURL)

	// Público: no requiere JWT (igual que /webhook/n8n/transcripcion)
	r.POST("/api/v1/rrv/sms/inbound", h.InboundSMS)

	// Protegidos: requieren JWT
	smsGroup := protected.Group("/rrv/sms")
	smsGroup.GET("/numeros", h.ListNumeros)
	smsGroup.POST("/numeros", h.AddNumero)
	smsGroup.DELETE("/numeros/:numero", h.RemoveNumero)
	smsGroup.GET("/historial", h.GetHistorial)
}

func firstCollection(collections []string) string {
	if len(collections) > 0 {
		return collections[0]
	}
	return "actas_rrv"
}
