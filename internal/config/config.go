package config

import "os"

// Config contiene la configuración de la aplicación leída desde variables de entorno.
type Config struct {
	MongoURI           string
	DBName             string
	Port               string
	TwilioAccountSID   string
	TwilioAuthToken    string
	TwilioMessagingSID string
	// Número verificado en la cuenta Twilio trial para enviar confirmaciones.
	// En producción puede ser cualquier número habilitado por Twilio.
	TwilioFromNumber string
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
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
