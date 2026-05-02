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
	ID             bson.ObjectID `bson:"_id,omitempty"          json:"id"`
	ActaID         string        `bson:"acta_id"                json:"acta_id"`
	Departamento   string        `bson:"departamento"           json:"departamento"`
	Provincia      string        `bson:"provincia"              json:"provincia"`
	Municipio      string        `bson:"municipio"              json:"municipio"`
	Recinto        string        `bson:"recinto"                json:"recinto"`
	Mesa           string        `bson:"mesa"                   json:"mesa"`
	NroMesa        string        `bson:"nro_mesa"               json:"nro_mesa"`

	ElectoresHabilitados int `bson:"electores_habilitados" json:"electores_habilitados"`
	PapeletasAnfora      int `bson:"papeletas_anfora"      json:"papeletas_anfora"`
	PapeletasNoUsadas    int `bson:"papeletas_no_usadas"   json:"papeletas_no_usadas"`
	AperturaHora         string `bson:"apertura_hora"      json:"apertura_hora"`
	CierreHora           string `bson:"cierre_hora"        json:"cierre_hora"`

	Candidatos   []Candidato `bson:"candidatos"    json:"candidatos"`
	VotosValidos int         `bson:"votos_validos" json:"votos_validos"`
	VotosNulos   int         `bson:"votos_nulos"   json:"votos_nulos"`
	VotosBlancos int         `bson:"votos_blancos" json:"votos_blancos"`
	TotalVotos   int         `bson:"total_votos"   json:"total_votos"`

	Estado      string `bson:"estado"       json:"estado"`
	Fuente      string `bson:"fuente"       json:"fuente"`
	TipoEntrada string `bson:"tipo_entrada" json:"tipo_entrada"`

	OCRMode        string   `bson:"ocr_mode"        json:"ocr_mode"`
	Confidence     float64  `bson:"confidence"      json:"confidence"`
	RequiresReview bool     `bson:"requires_review" json:"requires_review"`
	Warnings       []string `bson:"warnings"        json:"warnings"`
	Anomalies      []string `bson:"anomalies"       json:"anomalies"`

	HashOrigen     string    `bson:"hash_origen"      json:"hash_origen"`
	FechaRecepcion time.Time `bson:"fecha_recepcion"  json:"fecha_recepcion"`
}
