package database

import (
	"fmt"
	"log"
	"time"

	"github.com/srvof/votos-backend/config"
	"github.com/srvof/votos-backend/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connect abre la conexión a PostgreSQL (vía Pgpool) con reintentos.
func Connect(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort, cfg.SSLMode,
	)

	var db *gorm.DB
	var lastErr error
	maxAttempts := 30
	for i := 1; i <= maxAttempts; i++ {
		var err error
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger:                                   logger.Default.LogMode(logger.Warn),
			DisableForeignKeyConstraintWhenMigrating: true,
		})
		if err != nil {
			lastErr = err
			log.Printf("gorm open falló (intento %d/%d): %v", i, maxAttempts, err)
			time.Sleep(2 * time.Second)
			continue
		}
		sqlDB, err := db.DB()
		if err != nil {
			lastErr = err
			time.Sleep(2 * time.Second)
			continue
		}
		if err := sqlDB.Ping(); err != nil {
			lastErr = err
			log.Printf("ping falló (intento %d/%d): %v", i, maxAttempts, err)
			time.Sleep(2 * time.Second)
			continue
		}
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(50)
		sqlDB.SetConnMaxLifetime(time.Hour)
		return db, nil
	}
	return nil, fmt.Errorf("no se pudo conectar a la base de datos: %w", lastErr)
}

// AutoMigrate crea o actualiza el esquema.
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.Usuario{},
		&models.DistribucionTerritorial{},
		&models.RecintoElectoral{},
		&models.Mesa{},
		&models.Acta{},
		&models.OficialAuditoria{},
		&models.OficialError{},
		&models.OficialEvento{},
	)
}
