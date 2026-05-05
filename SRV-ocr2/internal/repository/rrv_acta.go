package repository

import (
	"context"
	"time"

	"srrv/internal/models"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
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

// GetBySMSMesa busca la primera acta enviada por SMS para esa mesa.
func (r *RRVActaRepository) GetBySMSMesa(ctx context.Context, mesa string) (*models.RRVActa, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var result models.RRVActa
	err := r.col.FindOne(ctx, bson.M{"mesa": mesa, "fuente": "SMS"}).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetByMesa busca cualquier acta existente para esa mesa (cualquier fuente).
func (r *RRVActaRepository) GetByMesa(ctx context.Context, mesa string) (*models.RRVActa, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var result models.RRVActa
	err := r.col.FindOne(ctx, bson.M{"mesa": mesa}).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetRecentSMS obtiene las últimas actas recibidas por SMS ordenadas por fecha descendente.
func (r *RRVActaRepository) GetRecentSMS(ctx context.Context, limit int) ([]models.RRVActa, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.D{{Key: "fecha_recepcion", Value: -1}}).SetLimit(int64(limit))

	cursor, err := r.col.Find(ctx, bson.M{"fuente": "SMS"}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []models.RRVActa
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	if results == nil {
		results = []models.RRVActa{}
	}
	return results, nil
}
