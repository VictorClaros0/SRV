package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"srrv/internal/config"
	"srrv/internal/data"
	"srrv/internal/database"
	"srrv/internal/handlers"
	"srrv/internal/repository"
	"srrv/internal/services"

	"github.com/gin-gonic/gin"
)

func main() {
	// ── Flags ─────────────────────────────────────────────────────────────────
	seedFlag := flag.Bool("seed", false, "Poblar la base de datos con datos iniciales desde CSV y salir")
	seedPath := flag.String("seed-path", "internal/data", "Ruta a la carpeta con los archivos CSV")
	flag.Parse()

	// ── Configuración ─────────────────────────────────────────────────────────
	cfg := config.Load()

	// ── MongoDB ───────────────────────────────────────────────────────────────
	client, err := database.Connect(cfg.MongoURI)
	if err != nil {
		log.Fatalf("❌ No se pudo conectar a MongoDB: %v", err)
	}
	defer database.Disconnect(client)

	db := client.Database(cfg.DBName)

	// ── Seed (opcional) ───────────────────────────────────────────────────────
	if *seedFlag {
		log.Println("🌱 Ejecutando seed de la base de datos...")
		if err := data.SeedDatabase(db, *seedPath); err != nil {
			log.Fatalf("❌ Error al hacer seed: %v", err)
		}
		log.Println("✅ Seed completado. Saliendo...")
		return
	}

	// ── Repositories ──────────────────────────────────────────────────────────
	distribucionRepo := repository.NewDistribucionRepository(db)
	recintoRepo      := repository.NewRecintoRepository(db)
	mesaRepo         := repository.NewMesaRepository(db)
	actaRepo         := repository.NewActaRepository(db)
	rrvActaRepo      := repository.NewRRVActaRepository(db, cfg.RRVCollection)
	eventoRepo       := repository.NewEventoRepository(db, cfg.EventsCollection)

	// ── Handlers ──────────────────────────────────────────────────────────────
	distribucionH := handlers.NewDistribucionHandler(distribucionRepo)
	recintoH      := handlers.NewRecintoHandler(recintoRepo)
	mesaH         := handlers.NewMesaHandler(mesaRepo)
	actaH         := handlers.NewActaHandler(actaRepo)
	twilioClient  := services.NewTwilioClient(cfg.TwilioAccountSID, cfg.TwilioAuthToken, cfg.TwilioMessagingSID, cfg.TwilioFromNumber)
	rrvH          := handlers.NewRRVHandler(rrvActaRepo, eventoRepo, twilioClient)

	// ── Router ────────────────────────────────────────────────────────────────
	r := gin.Default()

	// CORS permisivo para el dashboard de Pablo
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type,Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "db": cfg.DBName})
	})

	v1 := r.Group("/api/v1")
	{
		// DistribucionTerritorial
		dist := v1.Group("/distribuciones")
		{
			dist.GET("",      distribucionH.GetAll)
			dist.GET("/:id",  distribucionH.GetByID)
			dist.POST("",     distribucionH.Create)
			dist.PUT("/:id",  distribucionH.Update)
			dist.DELETE("/:id", distribucionH.Delete)
		}

		// RecintoElectoral
		rec := v1.Group("/recintos")
		{
			rec.GET("",      recintoH.GetAll)
			rec.GET("/:id",  recintoH.GetByID)
			rec.POST("",     recintoH.Create)
			rec.PUT("/:id",  recintoH.Update)
			rec.DELETE("/:id", recintoH.Delete)
		}

		// Mesa
		mesa := v1.Group("/mesas")
		{
			mesa.GET("",      mesaH.GetAll)
			mesa.GET("/:id",  mesaH.GetByID)
			mesa.POST("",     mesaH.Create)
			mesa.PUT("/:id",  mesaH.Update)
			mesa.DELETE("/:id", mesaH.Delete)
		}

		// Acta
		acta := v1.Group("/actas")
		{
			acta.GET("",      actaH.GetAll)
			acta.GET("/:id",  actaH.GetByID)
			acta.POST("",     actaH.Create)
			acta.PUT("/:id",  actaH.Update)
			acta.DELETE("/:id", actaH.Delete)
		}
	}

	// ── RRV (Sebastian) ───────────────────────────────────────────────────────
	rrv := r.Group("/api/rrv")
	{
		rrv.POST("/actas/upload",    rrvH.Upload)
		rrv.POST("/sms",             rrvH.SMS)
		rrv.POST("/webhook/sms",     rrvH.WebhookSMS) // Twilio envía aquí cuando llega un SMS al número virtual
		rrv.GET("/actas",            rrvH.GetAll)
		rrv.GET("/actas/:acta_id",   rrvH.GetByID)
		rrv.GET("/eventos",          rrvH.GetEventos)
	}

	// ── Arrancar servidor ─────────────────────────────────────────────────────
	log.Printf("🚀 Servidor iniciado en :%s", cfg.Port)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := r.Run(":" + cfg.Port); err != nil {
			log.Fatalf("❌ Error al iniciar el servidor: %v", err)
		}
	}()

	<-quit
	log.Println("🛑 Apagando servidor...")
}
