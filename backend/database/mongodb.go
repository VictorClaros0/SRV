package database

import (
	"context"
	"log"
	"time"

	"github.com/srvof/votos-backend/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// ConnectMongo abre la conexión a MongoDB con reintentos.
func ConnectMongo(cfg *config.Config) (*mongo.Client, error) {
	opts := options.Client().ApplyURI(cfg.MongoURI)

	var client *mongo.Client
	var lastErr error

	for i := 1; i <= 20; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		var err error
		client, err = mongo.Connect(ctx, opts)
		cancel()
		if err != nil {
			lastErr = err
			log.Printf("mongo connect falló (intento %d/20): %v", i, err)
			time.Sleep(3 * time.Second)
			continue
		}
		pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
		err = client.Ping(pingCtx, readpref.Primary())
		pingCancel()
		if err != nil {
			lastErr = err
			log.Printf("mongo ping falló (intento %d/20): %v", i, err)
			time.Sleep(3 * time.Second)
			continue
		}
		log.Printf("Conexión a MongoDB establecida")
		return client, nil
	}
	return nil, lastErr
}
