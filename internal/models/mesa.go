package models

import "go.mongodb.org/mongo-driver/v2/bson"

// Mesa representa una mesa de votación dentro de un recinto electoral.
type Mesa struct {
	ID                 bson.ObjectID `bson:"_id,omitempty"      json:"id"`
	Codigo             int           `bson:"codigo"             json:"codigo"             binding:"required"`
	CantidadHabilitada int           `bson:"cantidadHabilitada" json:"cantidadHabilitada" binding:"required"`
	Mesa               int           `bson:"mesa"               json:"mesa"               binding:"required"`
	IDRecintoElectoral bson.ObjectID `bson:"idRecintoElectoral" json:"idRecintoElectoral" binding:"required"`
}
