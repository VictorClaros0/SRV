package models

import "go.mongodb.org/mongo-driver/v2/bson"

// RecintoElectoral representa un recinto donde se realiza la votación.
type RecintoElectoral struct {
	ID                        bson.ObjectID `bson:"_id,omitempty"             json:"id"`
	RecintoID                 int           `bson:"recintoId"                 json:"recintoId"                 binding:"required"`
	Recinto                   string        `bson:"recinto"                   json:"recinto"                   binding:"required"`
	Direccion                 string        `bson:"direccion"                 json:"direccion"                 binding:"required"`
	Mesas                     int           `bson:"mesas"                     json:"mesas"                     binding:"required"`
	IDDistribucionTerritorial bson.ObjectID `bson:"idDistribucionTerritorial" json:"idDistribucionTerritorial"`
}
