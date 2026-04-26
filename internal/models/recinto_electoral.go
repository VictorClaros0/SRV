package models

import "go.mongodb.org/mongo-driver/v2/bson"

// RecintoElectoral representa un recinto donde se realiza la votación.
type RecintoElectoral struct {
	ID                        bson.ObjectID `bson:"_id,omitempty"             json:"id"`
	Nombre                    string        `bson:"nombre"                    json:"nombre"                    binding:"required"`
	IDDistribucionTerritorial bson.ObjectID `bson:"idDistribucionTerritorial" json:"idDistribucionTerritorial" binding:"required"`
}
