package services

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/srvof/votos-backend/models"
)

func HashBytes(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func SimulateOCR(fileBytes []byte) *models.RRVActa {
	hash := HashBytes(fileBytes)

	seed, _ := hex.DecodeString(hash[:2])
	base := int(seed[0])

	c1 := 100 + base*2
	c2 := 80 + base
	c3 := 40 + base/2
	nulos := 10 + base/10
	blancos := 5 + base/20
	total := c1 + c2 + c3 + nulos + blancos

	return &models.RRVActa{
		ActaID:       fmt.Sprintf("ACTA-SIM-%s", hash[:8]),
		Departamento: "La Paz",
		Municipio:    "El Alto",
		Recinto:      "Recinto Simulado OCR",
		Mesa:         fmt.Sprintf("M-%04d", base%9999+1),
		Candidatos: []models.Candidato{
			{CandidatoID: "CAND-01", Nombre: "Candidato 1", Votos: c1},
			{CandidatoID: "CAND-02", Nombre: "Candidato 2", Votos: c2},
			{CandidatoID: "CAND-03", Nombre: "Candidato 3", Votos: c3},
		},
		VotosNulos:   nulos,
		VotosBlancos: blancos,
		TotalVotos:   total,
		Estado:       "PROCESADA",
		Fuente:       "RRV",
		TipoEntrada:  "IMAGEN",
		HashOrigen:   hash,
	}
}
