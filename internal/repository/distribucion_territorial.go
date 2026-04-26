package repository

import (
	"context"
	"time"

	"srrv/internal/models"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

const colDistribucion = "distribuciones_territoriales"

// DistribucionRepository gestiona las operaciones CRUD sobre DistribucionTerritorial.
type DistribucionRepository struct {
	col *mongo.Collection
}

// NewDistribucionRepository crea una nueva instancia del repositorio.
func NewDistribucionRepository(db *mongo.Database) *DistribucionRepository {
	return &DistribucionRepository{col: db.Collection(colDistribucion)}
}

func (r *DistribucionRepository) GetAll(ctx context.Context) ([]models.DistribucionTerritorial, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var results []models.DistribucionTerritorial
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

func (r *DistribucionRepository) GetByID(ctx context.Context, id bson.ObjectID) (*models.DistribucionTerritorial, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var result models.DistribucionTerritorial
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *DistribucionRepository) Create(ctx context.Context, d *models.DistribucionTerritorial) (*models.DistribucionTerritorial, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	d.ID = bson.NewObjectID()
	_, err := r.col.InsertOne(ctx, d)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (r *DistribucionRepository) Update(ctx context.Context, id bson.ObjectID, d *models.DistribucionTerritorial) (*models.DistribucionTerritorial, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	update := bson.M{"$set": bson.M{
		"departamento": d.Departamento,
		"municipio":    d.Municipio,
		"provincia":    d.Provincia,
	}}
	_, err := r.col.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return nil, err
	}
	d.ID = id
	return d, nil
}

func (r *DistribucionRepository) Delete(ctx context.Context, id bson.ObjectID) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := r.col.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
