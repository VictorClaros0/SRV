package comparacion

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/gorm"
)

const coleccionActasRRV = "actas_rrv"

// actaRRVMongo es el struct de deserialización desde MongoDB.
type actaRRVMongo struct {
	ActaID       string            `bson:"acta_id"`
	Departamento string            `bson:"departamento"`
	Provincia    string            `bson:"provincia"`
	Municipio    string            `bson:"municipio"`
	Recinto      string            `bson:"recinto"`
	Mesa         string            `bson:"mesa"`
	Candidatos   []candidatoMongo  `bson:"candidatos"`
	VotosNulos   int               `bson:"votos_nulos"`
	VotosBlancos int               `bson:"votos_blancos"`
	TotalVotos   int               `bson:"total_votos"`
	Estado       string            `bson:"estado"`
	Fuente       string            `bson:"fuente"`
}

type candidatoMongo struct {
	CandidatoID string `bson:"candidato_id"`
	Nombre      string `bson:"nombre"`
	Votos       int    `bson:"votos"`
}

// actaOficialPG es el resultado del query agregado a PostgreSQL.
type actaOficialPG struct {
	CodigoActa int64  `gorm:"column:codigo_acta"`
	NroMesa    int    `gorm:"column:nro_mesa"`
	P1         int    `gorm:"column:p1"`
	P2         int    `gorm:"column:p2"`
	P3         int    `gorm:"column:p3"`
	P4         int    `gorm:"column:p4"`
	VotosNulos int    `gorm:"column:votos_nulos"`
	VotosBlanco int   `gorm:"column:votos_blanco"`
	VotosValidos int  `gorm:"column:votos_validos"`
	Estado     string `gorm:"column:estado"`
}

// Service orquesta la lectura de datos desde ambas fuentes y los normaliza.
type Service struct {
	DB          *gorm.DB
	MongoClient *mongo.Client
	DBName      string
}

// FetchRRV lee todas las actas de MongoDB y las normaliza a ActaFuente.
func (s *Service) FetchRRV(ctx context.Context) ([]ActaFuente, error) {
	col := s.MongoClient.Database(s.DBName).Collection(coleccionActasRRV)

	queryCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cur, err := col.Find(queryCtx, bson.D{}, options.Find().SetSort(bson.D{{Key: "acta_id", Value: 1}}))
	if err != nil {
		return nil, fmt.Errorf("mongo find actas_rrv: %w", err)
	}
	defer cur.Close(queryCtx)

	var result []ActaFuente
	for cur.Next(queryCtx) {
		var raw actaRRVMongo
		if err := cur.Decode(&raw); err != nil {
			continue
		}
		result = append(result, normalizarRRV(raw))
	}
	if err := cur.Err(); err != nil {
		return nil, fmt.Errorf("mongo cursor actas_rrv: %w", err)
	}
	return result, nil
}

// FetchOficiales lee las actas de PostgreSQL y las normaliza a ActaFuente.
// Sólo trae actas cuyo codigo_acta no es nulo (datos con identidad real).
func (s *Service) FetchOficiales(ctx context.Context) ([]ActaFuente, error) {
	var rows []actaOficialPG
	err := s.DB.WithContext(ctx).
		Table("acta").
		Select("codigo_acta, nro_mesa, p1, p2, p3, p4, votos_nulos, votos_blanco, votos_validos, estado").
		Where("codigo_acta IS NOT NULL AND fecha_eliminado IS NULL").
		Order("codigo_acta").
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("postgres fetch actas: %w", err)
	}

	result := make([]ActaFuente, 0, len(rows))
	for _, row := range rows {
		result = append(result, normalizarOficial(row))
	}
	return result, nil
}

// ─── Normalización ────────────────────────────────────────────────────────────

func normalizarRRV(raw actaRRVMongo) ActaFuente {
	candidatos := make([]CandidatoVotos, len(raw.Candidatos))
	for i, c := range raw.Candidatos {
		candidatos[i] = CandidatoVotos{
			CandidatoID: c.CandidatoID,
			Nombre:      c.Nombre,
			Votos:       c.Votos,
		}
	}
	fuente := raw.Fuente
	if fuente == "" {
		fuente = "RRV"
	}
	return ActaFuente{
		ActaID:       raw.ActaID,
		Departamento: raw.Departamento,
		Provincia:    raw.Provincia,
		Municipio:    raw.Municipio,
		Recinto:      raw.Recinto,
		Mesa:         raw.Mesa,
		Candidatos:   candidatos,
		VotosNulos:   raw.VotosNulos,
		VotosBlancos: raw.VotosBlancos,
		TotalVotos:   raw.TotalVotos,
		Estado:       raw.Estado,
		Fuente:       fuente,
	}
}

func normalizarOficial(row actaOficialPG) ActaFuente {
	// P1-P4 corresponden a los 4 partidos registrados en el sistema oficial.
	// Los nombres son genéricos porque PostgreSQL no almacena nombres de candidato.
	candidatos := []CandidatoVotos{
		{CandidatoID: "P1", Nombre: "Partido P1", Votos: row.P1},
		{CandidatoID: "P2", Nombre: "Partido P2", Votos: row.P2},
		{CandidatoID: "P3", Nombre: "Partido P3", Votos: row.P3},
		{CandidatoID: "P4", Nombre: "Partido P4", Votos: row.P4},
	}
	return ActaFuente{
		ActaID:       strconv.FormatInt(row.CodigoActa, 10),
		Mesa:         fmt.Sprintf("M-%02d", row.NroMesa),
		Candidatos:   candidatos,
		VotosNulos:   row.VotosNulos,
		VotosBlancos: row.VotosBlanco,
		TotalVotos:   row.VotosValidos,
		Estado:       row.Estado,
		Fuente:       "OFICIAL",
	}
}
