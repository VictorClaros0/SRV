package config

import (
	"os"

	"github.com/joho/godotenv"
)

// Config carga la configuración de la API y la base de datos.
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

	MongoURI string
	MongoDB  string

	TwilioAccountSID   string
	TwilioAuthToken    string
	TwilioMessagingSID string
	TwilioFromNumber   string
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

		MongoURI: getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDB:  getEnv("MONGO_DB", "electoral_db"),

		TwilioAccountSID:   getEnv("TWILIO_ACCOUNT_SID", ""),
		TwilioAuthToken:    getEnv("TWILIO_AUTH_TOKEN", ""),
		TwilioMessagingSID: getEnv("TWILIO_MESSAGING_SERVICE_SID", ""),
		TwilioFromNumber:   getEnv("TWILIO_VERIFIED_NUMBER", ""),
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
