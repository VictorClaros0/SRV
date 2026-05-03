package data

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"srrv/internal/models"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// SeedDatabase lee los archivos CSV y puebla las colecciones de MongoDB
func SeedDatabase(db *mongo.Database, basePath string) error {
	ctx := context.Background()

	// 1. Seed Distribucion Territorial
	log.Println("Seeding Distribucion Territorial...")
	distPath := filepath.Join(basePath, "DistribucionTerritorial.csv")
	distFile, err := os.Open(distPath)
	if err != nil {
		return fmt.Errorf("error abriendo %s: %v", distPath, err)
	}
	defer distFile.Close()

	distReader := csv.NewReader(distFile)
	distRecords, err := distReader.ReadAll()
	if err != nil {
		return fmt.Errorf("error leyendo %s: %v", distPath, err)
	}

	distColl := db.Collection("distribuciones")
	_ = distColl.Drop(ctx)

	distMap := make(map[int]bson.ObjectID)
	var distDocs []interface{}

	for i, row := range distRecords {
		if i == 0 {
			continue // saltar encabezado
		}
		codigo, _ := strconv.Atoi(row[0])
		id := bson.NewObjectID()
		distMap[codigo] = id

		doc := models.DistribucionTerritorial{
			ID:           id,
			Departamento: row[1],
			Municipio:    row[2],
			Provincia:    row[3],
		}
		distDocs = append(distDocs, doc)
	}

	if len(distDocs) > 0 {
		_, err = distColl.InsertMany(ctx, distDocs)
		if err != nil {
			return fmt.Errorf("error insertando distribuciones: %v", err)
		}
	}

	// 2. Seed Recintos Electorales
	log.Println("Seeding Recintos Electorales...")
	recPath := filepath.Join(basePath, "RecintosElectorales.csv")
	recFile, err := os.Open(recPath)
	if err != nil {
		return fmt.Errorf("error abriendo %s: %v", recPath, err)
	}
	defer recFile.Close()

	recReader := csv.NewReader(recFile)
	recRecords, err := recReader.ReadAll()
	if err != nil {
		return fmt.Errorf("error leyendo %s: %v", recPath, err)
	}

	recColl := db.Collection("recintos")
	_ = recColl.Drop(ctx)

	recMap := make(map[int]bson.ObjectID)
	var recDocs []interface{}

	for i, row := range recRecords {
		if i == 0 {
			continue // saltar encabezado
		}
		codigoTerritorial, _ := strconv.Atoi(row[0])
		recintoId, _ := strconv.Atoi(row[1])
		mesas, _ := strconv.Atoi(row[4])

		id := bson.NewObjectID()
		recMap[recintoId] = id

		distID, ok := distMap[codigoTerritorial]
		if !ok {
			log.Printf("Advertencia: CodigoTerritorial %d no encontrado para Recinto %d", codigoTerritorial, recintoId)
		}

		doc := models.RecintoElectoral{
			ID:                        id,
			RecintoID:                 recintoId,
			Recinto:                   row[2],
			Direccion:                 row[3],
			Mesas:                     mesas,
			IDDistribucionTerritorial: distID,
		}
		recDocs = append(recDocs, doc)
	}

	if len(recDocs) > 0 {
		_, err = recColl.InsertMany(ctx, recDocs)
		if err != nil {
			return fmt.Errorf("error insertando recintos: %v", err)
		}
	}

	// 3. Seed Mesas
	log.Println("Seeding Mesas...")
	mesaPath := filepath.Join(basePath, "Mesas.csv")
	mesaFile, err := os.Open(mesaPath)
	if err != nil {
		return fmt.Errorf("error abriendo %s: %v", mesaPath, err)
	}
	defer mesaFile.Close()

	mesaReader := csv.NewReader(mesaFile)
	mesaRecords, err := mesaReader.ReadAll()
	if err != nil {
		return fmt.Errorf("error leyendo %s: %v", mesaPath, err)
	}

	mesaColl := db.Collection("mesas")
	_ = mesaColl.Drop(ctx)

	var mesaDocs []interface{}

	for i, row := range mesaRecords {
		if i == 0 {
			continue // saltar encabezado
		}
		
		codigoMesa, _ := strconv.Atoi(row[1])
		mesaNum, _ := strconv.Atoi(row[2])
		nroVotantes, _ := strconv.Atoi(row[3])
		
		// El ID del recinto corresponde a los primeros 8 dígitos del código de la mesa
		recintoIdStr := row[1]
		if len(recintoIdStr) > 8 {
			recintoIdStr = recintoIdStr[:8]
		}
		recintoId, _ := strconv.Atoi(recintoIdStr)

		recID, ok := recMap[recintoId]
		if !ok {
			log.Printf("Advertencia: RecintoId %d no encontrado para Mesa %d", recintoId, codigoMesa)
		}

		doc := models.Mesa{
			ID:                 bson.NewObjectID(),
			Codigo:             codigoMesa,
			CantidadHabilitada: nroVotantes,
			Mesa:               mesaNum,
			IDRecintoElectoral: recID,
		}
		mesaDocs = append(mesaDocs, doc)
	}

	if len(mesaDocs) > 0 {
		_, err = mesaColl.InsertMany(ctx, mesaDocs)
		if err != nil {
			return fmt.Errorf("error insertando mesas: %v", err)
		}
	}

	log.Println("✅ Base de datos poblada exitosamente!")
	return nil
}
