package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/srvof/votos-backend/models"
	"gorm.io/gorm"
)


func seedCatalogData(db *gorm.DB) {
	if err := seedDistribuciones(db); err != nil {
		log.Printf("seed distribuciones: %v", err)
	}
	if err := seedRecintos(db); err != nil {
		log.Printf("seed recintos: %v", err)
	}
	if err := seedMesas(db); err != nil {
		log.Printf("seed mesas: %v", err)
	}
	if err := seedActas(db); err != nil {
		log.Printf("seed actas: %v", err)
	}
}

// seedActas lee ActasImpresas.csv y crea una acta en estado "impresa" por cada fila.
// Los votos se dejan en cero: el workflow de n8n es el encargado de transcribirlos
// llamando a POST /webhook/n8n/transcripcion con los datos de Transcripciones.csv.
func seedActas(db *gorm.DB) error {
	var count int64
	if err := db.Model(&models.Acta{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		log.Printf("seed actas: omitido (tabla ya contiene %d registros)", count)
		return nil
	}

	// Obtener ID del admin para campo de auditoría creado_por.
	var adminID *uint
	var admin models.Usuario
	if err := db.Where("es_admin = true").First(&admin).Error; err == nil {
		id := admin.ID
		adminID = &id
	}

	actasImpresasRows, err := readCSV(filepath.Join("data", "ActasImpresas.csv"))
	if err != nil {
		return fmt.Errorf("ActasImpresas.csv no encontrado: %w", err)
	}

	inserted := 0
	for _, row := range actasImpresasRows {
		if len(row) < 3 {
			continue
		}
		codigoActaCSV, err := strconv.ParseInt(strings.TrimSpace(row[1]), 10, 64)
		if err != nil {
			continue
		}
		nroMesa, err := strconv.Atoi(strings.TrimSpace(row[2]))
		if err != nil {
			continue
		}

		var codigoActa, codigoRecinto int64
		if codigoActaCSV%1000 == int64(nroMesa) {
			// Patrón normal: CodigoActa = CodigoRecinto(10 dígitos) + NroMesa(3 dígitos)
			codigoActa = codigoActaCSV
			codigoRecinto, err = strconv.ParseInt(strings.TrimSpace(row[0]), 10, 64)
			if err != nil {
				continue
			}
		} else {
			// Patrón especial: row[1] contiene el ID real del recinto (10 dígitos)
			// y row[0] es solo el código territorial. Se construye CodigoActa único.
			codigoRecinto = codigoActaCSV
			codigoActa = codigoActaCSV*1000 + int64(nroMesa)
		}

		acta := models.Acta{
			CodigoActa:    codigoActa,
			CodigoRecinto: codigoRecinto,
			NroMesa:       nroMesa,
			Estado:        "impresa",
			CreadoPorID:   adminID,
		}
		if err := db.Create(&acta).Error; err != nil {
			log.Printf("seed actas: error codigoActa=%d: %v", codigoActa, err)
			continue
		}
		inserted++
	}
	log.Printf("seed actas: %d actas insertadas en estado 'impresa' (pendientes de transcripción por n8n)", inserted)
	return nil
}

func seedDistribuciones(db *gorm.DB) error {
	var count int64
	if err := db.Model(&models.DistribucionTerritorial{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		log.Printf("seed distribuciones: omitido (tabla ya contiene %d registros)", count)
		return nil
	}

	rows, err := readCSV(filepath.Join("data", "DistribucionTerritorial.csv"))
	if err != nil {
		return err
	}
	for _, row := range rows {
		if len(row) < 4 {
			continue
		}
		item := models.DistribucionTerritorial{
			Departamento: strings.TrimSpace(row[1]),
			Municipio:    strings.TrimSpace(row[2]),
			Provincia:    strings.TrimSpace(row[3]),
		}
		if err := db.Create(&item).Error; err != nil {
			return err
		}
	}
	log.Printf("seed distribuciones: %d registros insertados", len(rows))
	return nil
}

func seedRecintos(db *gorm.DB) error {
	var count int64
	if err := db.Model(&models.RecintoElectoral{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		log.Printf("seed recintos: omitido (tabla ya contiene %d registros)", count)
		return nil
	}

	distribByCode, err := buildDistribucionCodeMap(db)
	if err != nil {
		return err
	}

	rows, err := readCSV(filepath.Join("data", "RecintosElectorales.csv"))
	if err != nil {
		return err
	}
	inserted := 0
	for _, row := range rows {
		if len(row) < 5 {
			continue
		}
		code := strings.TrimSpace(row[0])
		distID, ok := distribByCode[code]
		if !ok {
			continue
		}
		recintoID, err := strconv.ParseUint(strings.TrimSpace(row[1]), 10, 64)
		if err != nil {
			continue
		}
		recinto := models.RecintoElectoral{
			RecintoID:                 uint(recintoID),
			Recinto:                   strings.TrimSpace(row[2]),
			Direccion:                 strings.TrimSpace(row[3]),
			Mesas:                     strings.TrimSpace(row[4]),
			IDDistribucionTerritorial: distID,
		}
		if err := db.Create(&recinto).Error; err != nil {
			return err
		}
		inserted++
	}
	log.Printf("seed recintos: %d registros insertados", inserted)
	return nil
}

func seedMesas(db *gorm.DB) error {
	var count int64
	if err := db.Model(&models.Mesa{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		log.Printf("seed mesas: omitido (tabla ya contiene %d registros)", count)
		return nil
	}

	rows, err := readCSV(filepath.Join("data", "Mesas.csv"))
	if err != nil {
		return err
	}
	inserted := 0
	for _, row := range rows {
		if len(row) < 4 {
			continue
		}
		codigoMesa := strings.TrimSpace(row[1])
		if len(codigoMesa) < 4 {
			continue
		}
		recintoRaw := codigoMesa[:len(codigoMesa)-3]
		recintoID, err := strconv.ParseUint(recintoRaw, 10, 64)
		if err != nil {
			continue
		}
		votantes, err := strconv.Atoi(strings.TrimSpace(row[3]))
		if err != nil {
			continue
		}
		item := models.Mesa{
			Codigo:             codigoMesa,
			CantidadHabilitada: votantes,
			IdRecintoElectoral: uint(recintoID),
		}
		if err := db.Create(&item).Error; err != nil {
			return err
		}
		inserted++
	}
	log.Printf("seed mesas: %d registros insertados", inserted)
	return nil
}

func buildDistribucionCodeMap(db *gorm.DB) (map[string]uint, error) {
	rows, err := readCSV(filepath.Join("data", "DistribucionTerritorial.csv"))
	if err != nil {
		return nil, err
	}
	var items []models.DistribucionTerritorial
	if err := db.Order("id asc").Find(&items).Error; err != nil {
		return nil, err
	}
	if len(items) != len(rows) {
		return nil, fmt.Errorf("distribuciones en DB (%d) no coincide con CSV (%d)", len(items), len(rows))
	}
	out := make(map[string]uint, len(rows))
	for i, row := range rows {
		if len(row) < 1 {
			continue
		}
		out[strings.TrimSpace(row[0])] = items[i].ID
	}
	return out, nil
}

func readCSV(path string) ([][]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.FieldsPerRecord = -1

	_, err = r.Read() // header
	if err != nil {
		return nil, err
	}
	var rows [][]string
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		rows = append(rows, record)
	}
	return rows, nil
}
