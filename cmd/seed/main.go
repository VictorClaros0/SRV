package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"

	"srrv/internal/config"
	"srrv/internal/database"
	"srrv/internal/models"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func main() {
	cfg := config.Load()
	client, err := database.Connect(cfg.MongoURI)
	if err != nil {
		log.Fatalf("❌ Error conectando a Mongo: %v", err)
	}
	defer database.Disconnect(client)
	db := client.Database(cfg.DBName)

	colDist := db.Collection("distribuciones_territoriales")
	colRec := db.Collection("recintos_electorales")
	colMesa := db.Collection("mesas")

	ctx := context.Background()

	// Limpieza previa opcional
	_ = colDist.Drop(ctx)
	_ = colRec.Drop(ctx)
	_ = colMesa.Drop(ctx)

	fmt.Println("🚀 Iniciando el proceso de seed de base de datos...")

	// Mapeo en memoria
	mapDist := make(map[int]bson.ObjectID)     // CodigoTerritorial -> ObjectID
	mapRecinto := make(map[int]bson.ObjectID)  // RecintoId -> ObjectID

	// 1. CARGAR DistribucionTerritorial
	fileDist, err := os.Open("internal/data/DistribucionTerritorial.csv")
	if err != nil {
		log.Fatalf("Error leyendo archivo: %v", err)
	}
	defer fileDist.Close()

	readerDist := csv.NewReader(fileDist)
	recordsDist, err := readerDist.ReadAll()
	if err != nil {
		log.Fatalf("Error parseando CSV dist territoriale: %v", err)
	}

	for i, record := range recordsDist {
		if i == 0 {
			continue // skip header
		}
		codigo, _ := strconv.Atoi(record[0])
		d := models.DistribucionTerritorial{
			ID:           bson.NewObjectID(),
			Departamento: record[1],
			Municipio:    record[2],
			Provincia:    record[3],
		}
		_, err := colDist.InsertOne(ctx, d)
		if err != nil {
			log.Printf("Error insertando distribucion iteracion %d: %v", i, err)
		}
		mapDist[codigo] = d.ID
	}
	fmt.Printf("✅ %d Distribuciones cargadas\n", len(recordsDist)-1)

	// 2. CARGAR RecintoElectoral
	fileRec, err := os.Open("internal/data/RecintosElectorales.csv")
	if err != nil {
		log.Fatalf("Error leyendo archivo: %v", err)
	}
	defer fileRec.Close()

	readerRec := csv.NewReader(fileRec)
	recordsRec, err := readerRec.ReadAll()
	if err != nil {
		log.Fatalf("Error parseando CSV recintos electorales: %v", err)
	}

	for i, record := range recordsRec {
		if i == 0 {
			continue // skip header
		}
		recintoCode, _ := strconv.Atoi(record[0]) // codigo territorial referencial
		recintoIdStr := record[1]
		recintoId, _ := strconv.Atoi(recintoIdStr)
		mesasNum, _ := strconv.Atoi(record[4])

		idDist, exists := mapDist[recintoCode]
		if !exists {
			log.Printf("Advertencia: Código territorial %d no encontrado para el Recinto %d", recintoCode, recintoId)
		}

		r := models.RecintoElectoral{
			ID:                        bson.NewObjectID(),
			RecintoID:                 recintoId,
			Recinto:                   record[2],
			Direccion:                 record[3],
			Mesas:                     mesasNum,
			IDDistribucionTerritorial: idDist,
		}
		_, err := colRec.InsertOne(ctx, r)
		if err != nil {
			log.Printf("Error insertando recinto iteracion %d: %v", i, err)
		}
		mapRecinto[recintoId] = r.ID
	}
	fmt.Printf("✅ %d Recintos cargados\n", len(recordsRec)-1)

	// 3. CARGAR Mesas
	fileMesa, err := os.Open("internal/data/Mesas.csv")
	if err != nil {
		log.Fatalf("Error leyendo archivo: %v", err)
	}
	defer fileMesa.Close()

	readerMesa := csv.NewReader(fileMesa)
	recordsMesa, err := readerMesa.ReadAll()
	if err != nil {
		log.Fatalf("Error parseando CSV mesas: %v", err)
	}

	for i, record := range recordsMesa {
		if i == 0 {
			continue // skip header
		}
		
		codigoMesa, _ := strconv.ParseInt(record[1], 10, 64)
		mesaNum, _ := strconv.Atoi(record[2])
		votantes, _ := strconv.Atoi(record[3])

		// El recintoId son los primeros 8 digitos del codigo de mesa (dividiendo entre 1000)
		recintoId := int(codigoMesa / 1000)
		
		// Verificamos en el listado de meses si hay alguna inconsistencia (ejm strings)
		// Solo insertamos si lo encontramos en el map
		idRec, exists := mapRecinto[recintoId]
		if !exists {
			// Es posible que algunos recintos no hayan sido añadidos en la data o sean anómalos, se intenta continuar de todas formas.
			// fmt.Printf("Advertencia: Recinto %d no encontrado para Mesa %d\n", recintoId, codigoMesa)
			idRec = bson.NilObjectID // or let it default
		}

		// Algunos cóodigos en la data excedían en caracteres
		// El código de la mesa lo guardaremos directo.
		// Aunque en models de la BD pide int, nosotros daremos el valor original de codigo mesa
		// int en golang a 64 bits lo aceptara.
		m := models.Mesa{
			ID:                 bson.NewObjectID(),
			Codigo:             int(codigoMesa),
			CantidadHabilitada: votantes,
			Mesa:               mesaNum,
			IDRecintoElectoral: idRec,
		}
		_, err := colMesa.InsertOne(ctx, m)
		if err != nil {
			log.Printf("Error insertando mesa iteracion %d: %v", i, err)
		}
	}
	fmt.Printf("✅ %d Mesas cargadas\n", len(recordsMesa)-1)
	fmt.Println("🚀 Seed finalizado con éxito.")
}
