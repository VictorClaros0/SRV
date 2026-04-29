package repository

import (
	"context"
	"time"

	"github.com/srvof/votos-backend/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

const colEvento = "rrv_eventos"

type EventoRepository struct {
	col *mongo.Collection
}

func NewEventoRepository(db *mongo.Database) *EventoRepository {
	return &EventoRepository{col: db.Collection(colEvento)}
}

func (r *EventoRepository) Create(ctx context.Context, e *models.Evento) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	e.ID = bson.NewObjectID()
	_, err := r.col.InsertOne(ctx, e)
	return err
}

func (r *EventoRepository) GetAll(ctx context.Context) ([]models.Evento, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	var results []models.Evento
	cursor, err := r.col.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}
