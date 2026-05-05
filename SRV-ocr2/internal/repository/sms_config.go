package repository

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

const colSMSConfig = "sms_config"

type smsConfigDoc struct {
	ID                 string   `bson:"_id"`
	NumerosAutorizados []string `bson:"numeros_autorizados"`
}

type SMSConfigRepository struct {
	col            *mongo.Collection
	defaultNumbers []string
}

func NewSMSConfigRepository(db *mongo.Database, defaultNumbers []string) *SMSConfigRepository {
	return &SMSConfigRepository{
		col:            db.Collection(colSMSConfig),
		defaultNumbers: defaultNumbers,
	}
}

// GetAuthorizedNumbers retorna los números autorizados; si no hay documento devuelve los defaults.
func (r *SMSConfigRepository) GetAuthorizedNumbers(ctx context.Context) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var doc smsConfigDoc
	err := r.col.FindOne(ctx, bson.M{"_id": "config"}).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		return r.defaultNumbers, nil
	}
	if err != nil {
		return nil, fmt.Errorf("leer sms_config: %w", err)
	}
	return doc.NumerosAutorizados, nil
}

// IsAuthorized comprueba si el número normalizado está autorizado.
func (r *SMSConfigRepository) IsAuthorized(ctx context.Context, numero string) (bool, error) {
	numeros, err := r.GetAuthorizedNumbers(ctx)
	if err != nil {
		return false, err
	}
	for _, n := range numeros {
		if n == numero {
			return true, nil
		}
	}
	return false, nil
}

// AddAuthorizedNumber agrega un número a la lista; la crea desde defaults si no existe.
func (r *SMSConfigRepository) AddAuthorizedNumber(ctx context.Context, numero string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var existing smsConfigDoc
	err := r.col.FindOne(ctx, bson.M{"_id": "config"}).Decode(&existing)
	if err == mongo.ErrNoDocuments {
		nums := deduplicarNums(append(r.defaultNumbers, numero))
		_, err = r.col.InsertOne(ctx, smsConfigDoc{ID: "config", NumerosAutorizados: nums})
		return err
	}
	if err != nil {
		return fmt.Errorf("leer sms_config: %w", err)
	}
	_, err = r.col.UpdateOne(ctx,
		bson.M{"_id": "config"},
		bson.M{"$addToSet": bson.M{"numeros_autorizados": numero}},
	)
	return err
}

// RemoveAuthorizedNumber elimina un número de la lista autorizada.
func (r *SMSConfigRepository) RemoveAuthorizedNumber(ctx context.Context, numero string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var existing smsConfigDoc
	err := r.col.FindOne(ctx, bson.M{"_id": "config"}).Decode(&existing)
	if err == mongo.ErrNoDocuments {
		nums := deduplicarNums(r.defaultNumbers)
		filtered := make([]string, 0, len(nums))
		for _, n := range nums {
			if n != numero {
				filtered = append(filtered, n)
			}
		}
		_, err = r.col.InsertOne(ctx, smsConfigDoc{ID: "config", NumerosAutorizados: filtered})
		return err
	}
	if err != nil {
		return fmt.Errorf("leer sms_config: %w", err)
	}
	_, err = r.col.UpdateOne(ctx,
		bson.M{"_id": "config"},
		bson.M{"$pull": bson.M{"numeros_autorizados": numero}},
	)
	return err
}

func deduplicarNums(nums []string) []string {
	seen := make(map[string]bool, len(nums))
	out := make([]string, 0, len(nums))
	for _, n := range nums {
		if n != "" && !seen[n] {
			seen[n] = true
			out = append(out, n)
		}
	}
	return out
}
