package models

import "go.mongodb.org/mongo-driver/v2/bson"

// DistribucionTerritorial representa una unidad geográfica territorial.
type DistribucionTerritorial struct {
	ID           bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Departamento string        `bson:"departamento"  json:"departamento" binding:"required"`
	Municipio    string        `bson:"municipio"     json:"municipio"     binding:"required"`
	Provincia    string        `bson:"provincia"     json:"provincia"     binding:"required"`
}
