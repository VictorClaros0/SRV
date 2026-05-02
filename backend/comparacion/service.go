package comparacion

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/gorm"
)

const coleccionActasRRVDefault = "actas_rrv"

// actaRRVMongo es el struct de deserializacion desde MongoDB.
type actaRRVMongo struct {
	ActaID         string           `bson:"acta_id"`
	Departamento   string           `bson:"departamento"`
	Provincia      string           `bson:"provincia"`
	Municipio      string           `bson:"municipio"`
	Recinto        string           `bson:"recinto"`
	Mesa           string           `bson:"mesa"`
	Candidatos     []candidatoMongo `bson:"candidatos"`
	VotosNulos     int              `bson:"votos_nulos"`
	VotosBlancos   int              `bson:"votos_blancos"`
	VotosBlanco    int              `bson:"votos_blanco"`
	TotalVotos     int              `bson:"total_votos"`
	Estado         string           `bson:"estado"`
	Fuente         string           `bson:"fuente"`
	FechaRecepcion time.Time        `bson:"fecha_recepcion"`
	FechaProcesado time.Time        `bson:"fecha_procesado"`
}

type candidatoMongo struct {
	CandidatoID string `bson:"candidato_id"`
	Nombre      string `bson:"nombre"`
	Votos       int    `bson:"votos"`
}

// actaOficialPG es el resultado del query agregado a PostgreSQL.
type actaOficialPG struct {
	CodigoActa  int64  `gorm:"column:codigo_acta"`
	NroMesa     int    `gorm:"column:nro_mesa"`
	P1          int    `gorm:"column:p1"`
	P2          int    `gorm:"column:p2"`
	P3          int    `gorm:"column:p3"`
	P4          int    `gorm:"column:p4"`
	VotosNulos  int    `gorm:"column:votos_nulos"`
	VotosBlanco int    `gorm:"column:votos_blanco"`
	VotosValidos int   `gorm:"column:votos_validos"`
	Estado      string `gorm:"column:estado"`
}

type actaNormalizadaRRV struct {
	acta  ActaFuente
	fecha time.Time
	orden int
}

// Service orquesta la lectura de datos desde ambas fuentes y los normaliza.
type Service struct {
	DB             *gorm.DB
	MongoClient    *mongo.Client
	DBName         string
	RRVCollections []string
}

// FetchRRV lee una o mas colecciones MongoDB y normaliza actas RRV a ActaFuente.
func (s *Service) FetchRRV(ctx context.Context) ([]ActaFuente, error) {
	queryCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	porActa := map[string]actaNormalizadaRRV{}
	orden := 0
	for _, collectionName := range s.rrvCollections() {
		col := s.MongoClient.Database(s.DBName).Collection(collectionName)
		cur, err := col.Find(queryCtx, bson.D{}, options.Find().SetSort(bson.D{{Key: "acta_id", Value: 1}}))
		if err != nil {
			return nil, fmt.Errorf("mongo find %s: %w", collectionName, err)
		}

		for cur.Next(queryCtx) {
			var raw actaRRVMongo
			if err := cur.Decode(&raw); err != nil {
				continue
			}
			acta := normalizarRRV(raw)
			if strings.TrimSpace(acta.ActaID) == "" {
				continue
			}

			orden++
			nueva := actaNormalizadaRRV{acta: acta, fecha: fechaRRV(raw), orden: orden}
			if actual, ok := porActa[acta.ActaID]; !ok || debeReemplazarRRV(actual, nueva) {
				porActa[acta.ActaID] = nueva
			}
		}
		if err := cur.Err(); err != nil {
			_ = cur.Close(queryCtx)
			return nil, fmt.Errorf("mongo cursor %s: %w", collectionName, err)
		}
		_ = cur.Close(queryCtx)
	}

	keys := make([]string, 0, len(porActa))
	for actaID := range porActa {
		keys = append(keys, actaID)
	}
	sort.Strings(keys)

	result := make([]ActaFuente, 0, len(keys))
	for _, actaID := range keys {
		result = append(result, porActa[actaID].acta)
	}
	return result, nil
}

// FetchOficiales lee las actas de PostgreSQL y las normaliza a ActaFuente.
// Solo trae actas cuyo codigo_acta no es nulo (datos con identidad real).
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

func (s *Service) rrvCollections() []string {
	if len(s.RRVCollections) == 0 {
		return []string{coleccionActasRRVDefault}
	}
	var out []string
	seen := map[string]bool{}
	for _, collection := range s.RRVCollections {
		collection = strings.TrimSpace(collection)
		if collection == "" || seen[collection] {
			continue
		}
		seen[collection] = true
		out = append(out, collection)
	}
	if len(out) == 0 {
		return []string{coleccionActasRRVDefault}
	}
	return out
}

func fechaRRV(raw actaRRVMongo) time.Time {
	if !raw.FechaRecepcion.IsZero() {
		return raw.FechaRecepcion
	}
	return raw.FechaProcesado
}

func debeReemplazarRRV(actual, nueva actaNormalizadaRRV) bool {
	if !actual.fecha.IsZero() && !nueva.fecha.IsZero() {
		return nueva.fecha.After(actual.fecha)
	}
	if actual.fecha.IsZero() && !nueva.fecha.IsZero() {
		return true
	}
	if nueva.fecha.IsZero() {
		return nueva.orden > actual.orden
	}
	return nueva.orden > actual.orden
}

func normalizarRRV(raw actaRRVMongo) ActaFuente {
	candidatos := normalizarCandidatosRRV(raw.Candidatos)
	votosBlancos := raw.VotosBlancos
	if votosBlancos == 0 && raw.VotosBlanco != 0 {
		votosBlancos = raw.VotosBlanco
	}
	total := raw.TotalVotos
	if total == 0 {
		total = totalCandidatos(candidatos) + raw.VotosNulos + votosBlancos
	}

	fuente := raw.Fuente
	if fuente == "" {
		fuente = "RRV"
	}
	return ActaFuente{
		ActaID:       strings.TrimSpace(raw.ActaID),
		Departamento: raw.Departamento,
		Provincia:    raw.Provincia,
		Municipio:    raw.Municipio,
		Recinto:      raw.Recinto,
		Mesa:         raw.Mesa,
		Candidatos:   candidatos,
		VotosNulos:   raw.VotosNulos,
		VotosBlancos: votosBlancos,
		TotalVotos:   total,
		Estado:       raw.Estado,
		Fuente:       fuente,
	}
}

func normalizarCandidatosRRV(raw []candidatoMongo) []CandidatoVotos {
	porID := make(map[string]CandidatoVotos, len(raw)+4)
	for _, c := range raw {
		id := normalizarCandidatoIDRRV(c.CandidatoID)
		if id == "" {
			continue
		}
		nombre := strings.TrimSpace(c.Nombre)
		if nombre == "" {
			nombre = id
		}
		if prev, ok := porID[id]; ok {
			prev.Votos += c.Votos
			if prev.Nombre == "" || prev.Nombre == id {
				prev.Nombre = nombre
			}
			porID[id] = prev
			continue
		}
		porID[id] = CandidatoVotos{CandidatoID: id, Nombre: nombre, Votos: c.Votos}
	}

	defaultNames := map[string]string{
		"P1": "Partido P1",
		"P2": "Partido P2",
		"P3": "Partido P3",
		"P4": "Partido P4",
	}
	result := make([]CandidatoVotos, 0, len(porID)+4)
	for _, id := range []string{"P1", "P2", "P3", "P4"} {
		c, ok := porID[id]
		if !ok {
			c = CandidatoVotos{CandidatoID: id, Nombre: defaultNames[id], Votos: 0}
		}
		result = append(result, c)
		delete(porID, id)
	}

	var extras []string
	for id := range porID {
		extras = append(extras, id)
	}
	sort.Strings(extras)
	for _, id := range extras {
		result = append(result, porID[id])
	}
	return result
}

func normalizarCandidatoIDRRV(id string) string {
	id = strings.ToUpper(strings.TrimSpace(id))
	switch id {
	case "CAND-01", "CAND-1", "C1":
		return "P1"
	case "CAND-02", "CAND-2", "C2":
		return "P2"
	case "CAND-03", "CAND-3", "C3":
		return "P3"
	case "CAND-04", "CAND-4", "C4":
		return "P4"
	default:
		return id
	}
}

func totalCandidatos(candidatos []CandidatoVotos) int {
	total := 0
	for _, c := range candidatos {
		total += c.Votos
	}
	return total
}

func normalizarOficial(row actaOficialPG) ActaFuente {
	// P1-P4 corresponden a los 4 partidos registrados en el sistema oficial.
	// Para comparar con RRV se usa total general: validos + blancos + nulos.
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
		TotalVotos:   row.VotosValidos + row.VotosBlanco + row.VotosNulos,
		Estado:       row.Estado,
		Fuente:       "OFICIAL",
	}
}
