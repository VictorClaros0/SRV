package config

import (
	"os"
	"strings"
)

// Config contiene la configuración de la aplicación leída desde variables de entorno.
type Config struct {
	MongoURI           string
	DBName             string
	Port               string
	TwilioAccountSID   string
	TwilioAuthToken    string
	TwilioMessagingSID string
	TwilioFromNumber   string
	// Números de teléfono autorizados para enviar SMS (normalizados, sin prefijo +591).
	// Se pueden separar con comas en la variable SMS_DEFAULT_NUMBERS.
	SMSDefaultNumbers []string
}

// Load carga la configuración desde variables de entorno con valores por defecto.
func Load() *Config {
	return &Config{
		MongoURI:           getEnv("MONGO_URI", "mongodb://localhost:27017"),
		DBName:             getEnv("DB_NAME", "electoral_db"),
		Port:               getEnv("PORT", "8081"),
		TwilioAccountSID:   getEnv("TWILIO_ACCOUNT_SID", ""),
		TwilioAuthToken:    getEnv("TWILIO_AUTH_TOKEN", ""),
		TwilioMessagingSID: getEnv("TWILIO_MESSAGING_SERVICE_SID", ""),
		TwilioFromNumber:   getEnv("TWILIO_VERIFIED_NUMBER", ""),
		SMSDefaultNumbers:  parseCSV(getEnv("SMS_DEFAULT_NUMBERS", "")),
	}
}

func parseCSV(s string) []string {
	if s == "" {
		return []string{}
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
