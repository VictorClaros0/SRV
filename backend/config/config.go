package config

import (
	"os"

	"github.com/joho/godotenv"
)

// Config carga la configuración de la API y las bases de datos.
type Config struct {
	Port        string
	DBHost      string
	DBPort      string
	DBUser      string
	DBPassword  string
	DBName      string
	SSLMode     string
	JWTSecret   string
	AdminUser   string
	AdminPass   string
	CORSOrigins string
	// MongoDB
	MongoURI    string
	MongoDBName string
}

// Load lee variables de entorno; intenta cargar .env si existe.
func Load() *Config {
	_ = godotenv.Load()

	return &Config{
		Port:        getEnv("PORT", "8080"),
		DBHost:      getEnv("DB_HOST", "localhost"),
		DBPort:      getEnv("DB_PORT", "5432"),
		DBUser:      getEnv("DB_USER", "votos"),
		DBPassword:  getEnv("DB_PASSWORD", "votospass"),
		DBName:      getEnv("DB_NAME", "votos"),
		SSLMode:     getEnv("DB_SSLMODE", "disable"),
		JWTSecret:   getEnv("JWT_SECRET", "cambiar-en-produccion"),
		AdminUser:   getEnv("ADMIN_USER", "admin"),
		AdminPass:   getEnv("ADMIN_PASSWORD", "admin123"),
		CORSOrigins: getEnv("CORS_ORIGINS", "*"),
		MongoURI:    getEnv("MONGO_URI", "mongodb://admin:admin123@localhost:27017/?authSource=admin"),
		MongoDBName: getEnv("MONGO_DATABASE", "rrv"),
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
