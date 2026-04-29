package database

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/srvof/votos-backend/models"
	"github.com/srvof/votos-backend/services"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type rrvSeedEntry struct {
	ActaID       string             `json:"acta_id"`
	Departamento string             `json:"departamento"`
	Provincia    string             `json:"provincia"`
	Municipio    string             `json:"municipio"`
	Recinto      string             `json:"recinto"`
	Mesa         string             `json:"mesa"`
	Candidatos   []models.Candidato `json:"candidatos"`
	VotosNulos   int                `json:"votos_nulos"`
	VotosBlancos int                `json:"votos_blancos"`
	TotalVotos   int                `json:"total_votos"`
	Estado       string             `json:"estado"`
	Observacion  string             `json:"observacion"`
}

// SeedRRVActas carga automáticamente las actas desde data/rrv_seed.json a MongoDB.
// Actas observadas se registran con estado OBSERVADA sin votos (Ley N° 026).
// Actas ya existentes (por acta_id) se omiten para garantizar idempotencia.
func SeedRRVActas(db *mongo.Database, actasDir string) {
	seedPath := "data/rrv_seed.json"
	data, err := os.ReadFile(seedPath)
	if err != nil {
		log.Printf("rrv_seed: archivo %s no encontrado, omitiendo seed", seedPath)
		return
	}

	var entries []rrvSeedEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		log.Printf("rrv_seed: error leyendo JSON: %v", err)
		return
	}

	col := db.Collection("rrv_actas")
	ctx := context.Background()

	insertadas := 0
	omitidas := 0
	observadas := 0

	for _, entry := range entries {
		// Verificar si ya existe por acta_id
		count, err := col.CountDocuments(ctx, map[string]interface{}{"acta_id": entry.ActaID})
		if err != nil || count > 0 {
			omitidas++
			continue
		}

		// Leer el PDF para generar el hash
		pdfPath := actasDir + "/acta_" + entry.Mesa + ".pdf"
		var hash string
		pdfBytes, err := os.ReadFile(pdfPath)
		if err == nil {
			hash = services.HashBytes(pdfBytes)
		}

		acta := models.RRVActa{
			ActaID:         entry.ActaID,
			Departamento:   entry.Departamento,
			Municipio:      entry.Municipio,
			Recinto:        entry.Recinto,
			Mesa:           entry.Mesa,
			Candidatos:     entry.Candidatos,
			VotosNulos:     entry.VotosNulos,
			VotosBlancos:   entry.VotosBlancos,
			TotalVotos:     entry.TotalVotos,
			Estado:         entry.Estado,
			Fuente:         "RRV",
			TipoEntrada:    "IMAGEN",
			HashOrigen:     hash,
			FechaRecepcion: time.Now(),
		}

		if entry.Estado == "OBSERVADA" {
			log.Printf("[Ley N° 026] Acta %s OBSERVADA: %s — no se computa en el RRV", entry.ActaID, entry.Observacion)
			observadas++
		}

		if _, err := col.InsertOne(ctx, acta); err != nil {
			log.Printf("rrv_seed: error insertando %s: %v", entry.ActaID, err)
			continue
		}
		insertadas++
	}

	log.Printf("rrv_seed: %d actas cargadas (%d observadas por Ley N° 026), %d omitidas (ya existían)",
		insertadas, observadas, omitidas)
}
