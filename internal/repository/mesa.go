package repository

import (
	"context"
	"time"

	"srrv/internal/models"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

const colMesa = "mesas"

// MesaRepository gestiona las operaciones CRUD sobre Mesa.
type MesaRepository struct {
	col *mongo.Collection
}

// NewMesaRepository crea una nueva instancia del repositorio.
func NewMesaRepository(db *mongo.Database) *MesaRepository {
	return &MesaRepository{col: db.Collection(colMesa)}
}

func (r *MesaRepository) GetAll(ctx context.Context) ([]models.Mesa, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var results []models.Mesa
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

func (r *MesaRepository) GetByID(ctx context.Context, id bson.ObjectID) (*models.Mesa, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var result models.Mesa
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *MesaRepository) Create(ctx context.Context, m *models.Mesa) (*models.Mesa, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	m.ID = bson.NewObjectID()
	_, err := r.col.InsertOne(ctx, m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (r *MesaRepository) Update(ctx context.Context, id bson.ObjectID, m *models.Mesa) (*models.Mesa, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	update := bson.M{"$set": bson.M{
		"codigo":             m.Codigo,
		"cantidadHabilitada": m.CantidadHabilitada,
		"idRecintoElectoral": m.IDRecintoElectoral,
	}}
	_, err := r.col.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return nil, err
	}
	m.ID = id
	return m, nil
}

func (r *MesaRepository) Delete(ctx context.Context, id bson.ObjectID) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := r.col.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
