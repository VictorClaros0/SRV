package config

import "os"

// Config contiene la configuración de la aplicación leída desde variables de entorno.
type Config struct {
	MongoURI string
	DBName   string
	Port     string
}

// Load carga la configuración desde variables de entorno con valores por defecto.
func Load() *Config {
	return &Config{
		MongoURI: getEnv("MONGO_URI", "mongodb://localhost:27017"),
		DBName:   getEnv("DB_NAME", "electoral_db"),
		Port:     getEnv("PORT", "8080"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
