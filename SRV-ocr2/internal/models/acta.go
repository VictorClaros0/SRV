package models

import "go.mongodb.org/mongo-driver/v2/bson"

// Acta representa el acta de resultados de una mesa electoral.
type Acta struct {
	ID                bson.ObjectID `bson:"_id,omitempty"    json:"id"`
	PapeletasNoUsadas int           `bson:"papeletasNoUsadas" json:"papeletasNoUsadas"`
	P1                int           `bson:"p1"               json:"p1"`
	P2                int           `bson:"p2"               json:"p2"`
	P3                int           `bson:"p3"               json:"p3"`
	P4                int           `bson:"p4"               json:"p4"`
	VotosNulos        int           `bson:"votosNulos"       json:"votosNulos"`
	VotosBlanco       int           `bson:"votosBlanco"      json:"votosBlanco"`
	VotosValidos      int           `bson:"votosValidos"     json:"votosValidos"`
	IDMesa            bson.ObjectID `bson:"idMesa"           json:"idMesa"`
	CodigoMesa        string        `bson:"codigoMesa"       json:"codigoMesa"`
	Mesa              int           `bson:"mesa"             json:"mesa"`
	Departamento      string        `bson:"departamento"     json:"departamento"`
	Provincia         string        `bson:"provincia"        json:"provincia"`
	Municipio         string        `bson:"municipio"        json:"municipio"`
	Recinto           string        `bson:"recinto"          json:"recinto"`
	VotantesHabilitados int         `bson:"votantesHabilitados" json:"votantesHabilitados"`
	PapeletasAnfora   int           `bson:"papeletasAnfora"   json:"papeletasAnfora"`
	AperturaHora      int           `bson:"aperturaHora"      json:"aperturaHora"`
	AperturaMinutos   int           `bson:"aperturaMinutos"   json:"aperturaMinutos"`
	CierreHora        int           `bson:"cierreHora"        json:"cierreHora"`
	CierreMinutos     int           `bson:"cierreMinutos"     json:"cierreMinutos"`
	TipoCliente       string        `bson:"tipoCliente"      json:"tipoCliente"      binding:"required"`
}
