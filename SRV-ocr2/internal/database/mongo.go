package database

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Connect establece la conexión con MongoDB y verifica la conectividad con Ping.
// En v2 de mongo-driver, Connect ya no recibe un context como primer argumento.
func Connect(uri string) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(clientOpts)
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	log.Println("✅ Conectado a MongoDB exitosamente")
	return client, nil
}

// Disconnect cierra la conexión con MongoDB de forma limpia.
func Disconnect(client *mongo.Client) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Disconnect(ctx); err != nil {
		log.Printf("⚠️  Error al desconectar MongoDB: %v", err)
	} else {
		log.Println("🔌 MongoDB desconectado correctamente")
	}
}
