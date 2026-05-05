package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Evento struct {
	ID        bson.ObjectID `bson:"_id,omitempty"   json:"id"`
	ActaID    string        `bson:"acta_id"         json:"acta_id"`
	Tipo      string        `bson:"tipo"            json:"tipo"`
	Fuente    string        `bson:"fuente"          json:"fuente"`
	Payload   bson.M        `bson:"payload"         json:"payload"`
	Error     string        `bson:"error,omitempty" json:"error,omitempty"`
	Timestamp time.Time     `bson:"timestamp"       json:"timestamp"`
}
