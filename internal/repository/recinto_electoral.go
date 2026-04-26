package repository

import (
	"context"
	"time"

	"srrv/internal/models"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

const colRecinto = "recintos_electorales"

// RecintoRepository gestiona las operaciones CRUD sobre RecintoElectoral.
type RecintoRepository struct {
	col *mongo.Collection
}

// NewRecintoRepository crea una nueva instancia del repositorio.
func NewRecintoRepository(db *mongo.Database) *RecintoRepository {
	return &RecintoRepository{col: db.Collection(colRecinto)}
}

func (r *RecintoRepository) GetAll(ctx context.Context) ([]models.RecintoElectoral, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var results []models.RecintoElectoral
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

func (r *RecintoRepository) GetByID(ctx context.Context, id bson.ObjectID) (*models.RecintoElectoral, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var result models.RecintoElectoral
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *RecintoRepository) Create(ctx context.Context, rec *models.RecintoElectoral) (*models.RecintoElectoral, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rec.ID = bson.NewObjectID()
	_, err := r.col.InsertOne(ctx, rec)
	if err != nil {
		return nil, err
	}
	return rec, nil
}

func (r *RecintoRepository) Update(ctx context.Context, id bson.ObjectID, rec *models.RecintoElectoral) (*models.RecintoElectoral, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	update := bson.M{"$set": bson.M{
		"recintoId":                 rec.RecintoID,
		"recinto":                   rec.Recinto,
		"direccion":                 rec.Direccion,
		"mesas":                     rec.Mesas,
		"idDistribucionTerritorial": rec.IDDistribucionTerritorial,
	}}
	_, err := r.col.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return nil, err
	}
	rec.ID = id
	return rec, nil
}

func (r *RecintoRepository) Delete(ctx context.Context, id bson.ObjectID) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := r.col.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
