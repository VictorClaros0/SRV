package repository

import (
	"context"
	"time"

	"srrv/internal/models"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

const colRRVActa = "rrv_actas"

type RRVActaRepository struct {
	col *mongo.Collection
}

func NewRRVActaRepository(db *mongo.Database) *RRVActaRepository {
	return &RRVActaRepository{col: db.Collection(colRRVActa)}
}

func (r *RRVActaRepository) ExistsByActaID(ctx context.Context, actaID string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	count, err := r.col.CountDocuments(ctx, bson.M{"acta_id": actaID})
	return count > 0, err
}

func (r *RRVActaRepository) ExistsByHash(ctx context.Context, hash string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	count, err := r.col.CountDocuments(ctx, bson.M{"hash_origen": hash})
	return count > 0, err
}

func (r *RRVActaRepository) Create(ctx context.Context, a *models.RRVActa) (*models.RRVActa, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	a.ID = bson.NewObjectID()
	_, err := r.col.InsertOne(ctx, a)
	return a, err
}

func (r *RRVActaRepository) GetAll(ctx context.Context) ([]models.RRVActa, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	var results []models.RRVActa
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

func (r *RRVActaRepository) GetByActaID(ctx context.Context, actaID string) (*models.RRVActa, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var result models.RRVActa
	if err := r.col.FindOne(ctx, bson.M{"acta_id": actaID}).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}
