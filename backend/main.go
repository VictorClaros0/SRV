package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/srvof/votos-backend/config"
	"github.com/srvof/votos-backend/database"
	"github.com/srvof/votos-backend/models"
	"github.com/srvof/votos-backend/routes"
	"github.com/srvof/votos-backend/utils"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()

	// PostgreSQL — cómputo oficial
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("conexión PostgreSQL: %v", err)
	}
	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("migración: %v", err)
	}
	seedCatalogData(db)
	seedAdmin(db, cfg)

	// MongoDB — módulo RRV (conteo rápido)
	mongoClient, err := database.ConnectMongo(cfg.MongoURI)
	if err != nil {
		log.Fatalf("conexión MongoDB: %v", err)
	}
	defer database.DisconnectMongo(mongoClient)
	mongoDB := mongoClient.Database(cfg.MongoDB)
	database.SeedRRVActas(mongoDB, "actas")

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())
	routes.Setup(r, db, mongoDB, cfg)

	addr := ":" + cfg.Port
	log.Printf("API escuchando en %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("servidor: %v", err)
	}
}

func seedAdmin(db *gorm.DB, cfg *config.Config) {
	var n int64
	if err := db.Model(&models.Usuario{}).Count(&n).Error; err != nil {
		log.Printf("seed: no se pudo contar usuarios: %v", err)
		return
	}
	if n > 0 {
		return
	}
	hash, err := utils.HashPassword(cfg.AdminPass)
	if err != nil {
		log.Fatalf("seed: hash admin: %v", err)
	}
	u := models.Usuario{
		Nombre:        "Administrador",
		Apellido:      "Sistema",
		NombreUsuario: cfg.AdminUser,
		Contrasena:    hash,
		EsAdmin:       true,
	}
	if err := db.Create(&u).Error; err != nil {
		log.Fatalf("seed: crear admin: %v", err)
	}
	log.Printf("Usuario admin inicial creado: %s", cfg.AdminUser)
}
