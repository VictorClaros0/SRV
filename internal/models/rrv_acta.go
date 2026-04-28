package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Candidato struct {
	CandidatoID string `bson:"candidato_id" json:"candidato_id"`
	Nombre      string `bson:"nombre"       json:"nombre"`
	Votos       int    `bson:"votos"        json:"votos"`
}

type RRVActa struct {
	ID             bson.ObjectID `bson:"_id,omitempty"   json:"id"`
	ActaID         string        `bson:"acta_id"         json:"acta_id"`
	Departamento   string        `bson:"departamento"    json:"departamento"`
	Municipio      string        `bson:"municipio"       json:"municipio"`
	Recinto        string        `bson:"recinto"         json:"recinto"`
	Mesa           string        `bson:"mesa"            json:"mesa"`
	Candidatos     []Candidato   `bson:"candidatos"      json:"candidatos"`
	VotosNulos     int           `bson:"votos_nulos"     json:"votos_nulos"`
	VotosBlancos   int           `bson:"votos_blancos"   json:"votos_blancos"`
	TotalVotos     int           `bson:"total_votos"     json:"total_votos"`
	Estado         string        `bson:"estado"          json:"estado"`
	Fuente         string        `bson:"fuente"          json:"fuente"`
	TipoEntrada    string        `bson:"tipo_entrada"    json:"tipo_entrada"`
	HashOrigen     string        `bson:"hash_origen"     json:"hash_origen"`
	FechaRecepcion time.Time     `bson:"fecha_recepcion" json:"fecha_recepcion"`
}
