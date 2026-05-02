package repository

import (
	"context"
	"time"

	"srrv/internal/models"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

const defaultRRVActaCollection = "actas_rrv"

type RRVActaRepository struct {
	col *mongo.Collection
}

func NewRRVActaRepository(db *mongo.Database, collectionName string) *RRVActaRepository {
	if collectionName == "" {
		collectionName = defaultRRVActaCollection
	}
	return &RRVActaRepository{col: db.Collection(collectionName)}
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

// DeleteByHash removes the acta identified by hash_origen. Used only when
// OCR_ALLOW_REPROCESS=true to allow re-uploading the same file after improving OCR.
func (r *RRVActaRepository) DeleteByHash(ctx context.Context, hash string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	_, err := r.col.DeleteOne(ctx, bson.M{"hash_origen": hash})
	return err
}

// DeleteByActaID removes the acta identified by acta_id. Used as cleanup
// companion to DeleteByHash during reprocessing.
func (r *RRVActaRepository) DeleteByActaID(ctx context.Context, actaID string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	_, err := r.col.DeleteOne(ctx, bson.M{"acta_id": actaID})
	return err
}
