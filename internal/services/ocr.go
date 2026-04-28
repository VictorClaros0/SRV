package services

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"srrv/internal/models"
)

// HashBytes returns the SHA-256 hex digest of data.
func HashBytes(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// SimulateOCR produces a deterministic simulated OCR result from raw file bytes.
// Different files yield different (but consistent) vote counts derived from the hash.
func SimulateOCR(fileBytes []byte) *models.RRVActa {
	hash := HashBytes(fileBytes)

	// Use first two hash bytes as a seed for vote variation (0-255 range).
	seed, _ := hex.DecodeString(hash[:2])
	base := int(seed[0]) // 0-255

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
