package config

import (
	"os"
	"strings"

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
	MongoURI              string
	MongoDBName           string
	MongoRRVCollections   []string
	MongoEventsCollection string
	// SMS
	SMSDefaultNumbers  []string
	SMSConfigCollection string
	// n8n
	N8NWebhookURL string
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
		MongoURI:              getEnv("MONGO_URI", "mongodb://admin:admin123@localhost:27017/?authSource=admin"),
		MongoDBName:           getEnv("MONGO_DATABASE", "rrv"),
		MongoRRVCollections:   getEnvList("MONGO_RRV_COLLECTIONS", getEnv("MONGO_RRV_COLLECTION", "actas_rrv")),
		MongoEventsCollection: getEnv("MONGO_EVENTS_COLLECTION", "rrv_eventos"),
		SMSDefaultNumbers:     getEnvList("SMS_AUTHORIZED_NUMBERS", "65707079"),
		SMSConfigCollection:   getEnv("SMS_CONFIG_COLLECTION", "sms_config"),
		N8NWebhookURL:         getEnv("N8N_TRIGGER_WEBHOOK_URL", "http://n8n:5678/webhook/trigger-transcripcion"),
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvList(key, def string) []string {
	raw := getEnv(key, def)
	var out []string
	seen := map[string]bool{}
	for _, part := range strings.Split(raw, ",") {
		item := strings.TrimSpace(part)
		if item == "" || seen[item] {
			continue
		}
		seen[item] = true
		out = append(out, item)
	}
	if len(out) == 0 {
		return []string{"actas_rrv"}
	}
	return out
}
