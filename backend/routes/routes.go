package routes

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/srvof/votos-backend/comparacion"
	"github.com/srvof/votos-backend/config"
	"github.com/srvof/votos-backend/dashboard"
	"github.com/srvof/votos-backend/handlers"
	"github.com/srvof/votos-backend/middleware"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

// Setup registra middlewares y rutas.
func Setup(r *gin.Engine, db *gorm.DB, cfg *config.Config, mongoClient *mongo.Client) {
	r.Use(corsMiddleware(cfg.CORSOrigins))

	authH := &handlers.AuthHandler{DB: db, Config: cfg}
	userH := &handlers.UsuarioHandler{DB: db}
	distH := &handlers.DistribucionHandler{DB: db}
	recH := &handlers.RecintoHandler{DB: db}
	mesaH := &handlers.MesaHandler{DB: db}
	actaH := &handlers.ActaHandler{DB: db}
	resH := &handlers.ResultadosHandler{DB: db}

	r.GET("/health", health(db))

	// Endpoints públicos consumidos por n8n (sin JWT)
	r.POST("/webhook/n8n/transcripcion", actaH.WebhookN8N)
	r.GET("/api/v1/actas/para-transcribir", actaH.ParaTranscribir)
	r.POST("/api/v1/actas/procesar-transcripciones", actaH.ProcesarTranscripciones)

	v1 := r.Group("/api/v1")

	v1.POST("/auth/login", authH.Login)

	protected := v1.Group("")
	protected.Use(middleware.AuthJWT(cfg.JWTSecret))
	protected.GET("/auth/me", authH.Me)
	protected.POST("/auth/register", middleware.RequireAdmin(), authH.Register)

	protected.GET("/usuarios", middleware.RequireAdmin(), userH.List)
	protected.GET("/usuarios/:id", userH.Get)
	protected.PUT("/usuarios/:id", userH.Update)
	protected.DELETE("/usuarios/:id", middleware.RequireAdmin(), userH.Delete)

	protected.GET("/distribuciones", distH.List)
	protected.GET("/distribuciones/:id", distH.Get)
	protected.POST("/distribuciones", middleware.RequireAdmin(), distH.Create)
	protected.PUT("/distribuciones/:id", middleware.RequireAdmin(), distH.Update)
	protected.DELETE("/distribuciones/:id", middleware.RequireAdmin(), distH.Delete)

	protected.GET("/recintos", recH.List)
	protected.GET("/recintos/:id", recH.Get)
	protected.POST("/recintos", middleware.RequireAdmin(), recH.Create)
	protected.PUT("/recintos/:id", middleware.RequireAdmin(), recH.Update)
	protected.DELETE("/recintos/:id", middleware.RequireAdmin(), recH.Delete)

	protected.GET("/mesas", mesaH.List)
	protected.GET("/mesas/:id", mesaH.Get)
	protected.POST("/mesas", middleware.RequireAdmin(), mesaH.Create)
	protected.PUT("/mesas/:id", middleware.RequireAdmin(), mesaH.Update)
	protected.DELETE("/mesas/:id", middleware.RequireAdmin(), mesaH.Delete)

	protected.GET("/actas", actaH.List)
	protected.GET("/actas/resumen", actaH.Resumen)
	protected.GET("/actas/:id", actaH.Get)
	protected.POST("/actas", actaH.Create)
	protected.PUT("/actas/:id", actaH.Update)
	protected.DELETE("/actas/:id", actaH.Delete)

	protected.GET("/resultados", resH.Resultados)
	protected.GET("/auditoria", resH.Auditoria)

	// Módulo de comparación RRV vs Oficial (solo lectura / CQRS Query)
	comparacion.RegisterRoutes(protected, db, mongoClient, cfg.MongoDBName)

	// Dashboard endpoints listos para Chart.js (solo lectura / CQRS Query)
	dashboard.RegisterRoutes(protected, db, mongoClient, cfg.MongoDBName)
}

func corsMiddleware(allowed string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := allowed
		if origin == "" {
			origin = "*"
		}
		reqOrigin := c.GetHeader("Origin")
		if origin != "*" && reqOrigin != "" {
			for _, o := range strings.Split(origin, ",") {
				o = strings.TrimSpace(o)
				if o == reqOrigin {
					c.Header("Access-Control-Allow-Origin", reqOrigin)
					break
				}
			}
		} else {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization,Content-Type")
		c.Header("Access-Control-Expose-Headers", "Authorization")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func health(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "error", "database": "unavailable"})
			return
		}
		if err := sqlDB.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "degraded", "database": "down", "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok", "database": "up"})
	}
}
