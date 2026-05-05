package repository

import (
	"context"
	"time"

	"srrv/internal/models"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

const colActa = "actas"

// ActaRepository gestiona las operaciones CRUD sobre Acta.
type ActaRepository struct {
	col *mongo.Collection
}

// NewActaRepository crea una nueva instancia del repositorio.
func NewActaRepository(db *mongo.Database) *ActaRepository {
	return &ActaRepository{col: db.Collection(colActa)}
}

func (r *ActaRepository) GetAll(ctx context.Context) ([]models.Acta, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var results []models.Acta
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

func (r *ActaRepository) GetByID(ctx context.Context, id bson.ObjectID) (*models.Acta, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var result models.Acta
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *ActaRepository) Create(ctx context.Context, a *models.Acta) (*models.Acta, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	a.ID = bson.NewObjectID()
	_, err := r.col.InsertOne(ctx, a)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (r *ActaRepository) ExistsByCodigoMesa(ctx context.Context, codigoMesa string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	count, err := r.col.CountDocuments(ctx, bson.M{"codigoMesa": codigoMesa})
	return count > 0, err
}

func (r *ActaRepository) Update(ctx context.Context, id bson.ObjectID, a *models.Acta) (*models.Acta, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	update := bson.M{"$set": bson.M{
		"papeletasNoUsadas": a.PapeletasNoUsadas,
		"p1":                a.P1,
		"p2":                a.P2,
		"p3":                a.P3,
		"p4":                a.P4,
		"votosNulos":        a.VotosNulos,
		"votosBlanco":       a.VotosBlanco,
		"votosValidos":      a.VotosValidos,
		"idMesa":            a.IDMesa,
		"tipoCliente":       a.TipoCliente,
	}}
	_, err := r.col.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return nil, err
	}
	a.ID = id
	return a, nil
}

func (r *ActaRepository) Delete(ctx context.Context, id bson.ObjectID) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := r.col.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
