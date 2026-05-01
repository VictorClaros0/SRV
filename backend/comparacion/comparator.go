package comparacion

import (
	"fmt"
	"math"
	"strings"
	"time"
)

// Comparer ejecuta la lógica de comparación entre fuentes oficial y RRV.
type Comparer struct{}

// Comparar recibe listas normalizadas de ambas fuentes y devuelve el resumen
// con todas las actas clasificadas. No modifica ningún dato de entrada (CQRS/Query).
func (c *Comparer) Comparar(oficiales, rrv []ActaFuente) ResumenComparacion {
	rrvMap := make(map[string]*ActaFuente, len(rrv))
	for i := range rrv {
		rrvMap[rrv[i].ActaID] = &rrv[i]
	}

	var actas []ActaComparada
	visitados := make(map[string]bool, len(oficiales))

	for i := range oficiales {
		of := &oficiales[i]
		visitados[of.ActaID] = true
		if rv, ok := rrvMap[of.ActaID]; ok {
			actas = append(actas, c.compararPar(of, rv))
		} else {
			actas = append(actas, soloOficial(of))
		}
	}

	for i := range rrv {
		if !visitados[rrv[i].ActaID] {
			actas = append(actas, soloRRV(&rrv[i]))
		}
	}

	return buildResumen(actas, len(oficiales), len(rrv))
}

// ─── Reglas de comparación ────────────────────────────────────────────────────

// compararPar aplica las 8 reglas de comparación cuando el acta existe en ambas fuentes.
func (c *Comparer) compararPar(of, rv *ActaFuente) ActaComparada {
	result := ActaComparada{
		ActaID:       of.ActaID,
		Departamento: firstNonEmpty(rv.Departamento, of.Departamento),
		Provincia:    firstNonEmpty(rv.Provincia, of.Provincia),
		Municipio:    firstNonEmpty(rv.Municipio, of.Municipio),
		Recinto:      firstNonEmpty(rv.Recinto, of.Recinto),
		Mesa:         firstNonEmpty(rv.Mesa, of.Mesa),
		DatosRRV:     rv,
		DatosOficial: of,
	}

	// Regla 8: estado conflictivo (anulada en una, validada en otra)
	if estadoConflictivo(of.Estado, rv.Estado) {
		result.Inconsistencias = append(result.Inconsistencias, Inconsistencia{
			ActaID:         of.ActaID,
			Tipo:           TipoEstadoConflictivo,
			Descripcion:    fmt.Sprintf("Estado oficial=%q vs RRV=%q", of.Estado, rv.Estado),
			Severidad:      SeveridadAlta,
			FuenteAfectada: "AMBAS",
		})
	}

	dif := &DiferenciaActa{}

	// Regla 4: diferencia total de votos
	dif.DiferenciaTotalVotos = rv.TotalVotos - of.TotalVotos
	if dif.DiferenciaTotalVotos != 0 {
		result.Inconsistencias = append(result.Inconsistencias, Inconsistencia{
			ActaID:         of.ActaID,
			Tipo:           TipoDiferenciaTotalVotos,
			Descripcion:    fmt.Sprintf("Total votos: oficial=%d RRV=%d (Δ%+d)", of.TotalVotos, rv.TotalVotos, dif.DiferenciaTotalVotos),
			Severidad:      severidadPorDelta(abs(dif.DiferenciaTotalVotos)),
			FuenteAfectada: "AMBAS",
		})
	}

	// Regla 6: diferencia en votos nulos
	dif.DiferenciaVotosNulos = rv.VotosNulos - of.VotosNulos
	if dif.DiferenciaVotosNulos != 0 {
		result.Inconsistencias = append(result.Inconsistencias, Inconsistencia{
			ActaID:         of.ActaID,
			Tipo:           TipoDiferenciaNulos,
			Descripcion:    fmt.Sprintf("Votos nulos: oficial=%d RRV=%d (Δ%+d)", of.VotosNulos, rv.VotosNulos, dif.DiferenciaVotosNulos),
			Severidad:      SeveridadBaja,
			FuenteAfectada: "AMBAS",
		})
	}

	// Regla 7: diferencia en votos blancos
	dif.DiferenciaVotosBlancos = rv.VotosBlancos - of.VotosBlancos
	if dif.DiferenciaVotosBlancos != 0 {
		result.Inconsistencias = append(result.Inconsistencias, Inconsistencia{
			ActaID:         of.ActaID,
			Tipo:           TipoDiferenciaBlancos,
			Descripcion:    fmt.Sprintf("Votos blancos: oficial=%d RRV=%d (Δ%+d)", of.VotosBlancos, rv.VotosBlancos, dif.DiferenciaVotosBlancos),
			Severidad:      SeveridadBaja,
			FuenteAfectada: "AMBAS",
		})
	}

	// Regla 5: diferencia por candidato
	ofMap := make(map[string]CandidatoVotos, len(of.Candidatos))
	for _, cand := range of.Candidatos {
		ofMap[cand.CandidatoID] = cand
	}
	for _, rc := range rv.Candidatos {
		oc := ofMap[rc.CandidatoID]
		delta := rc.Votos - oc.Votos
		nombre := firstNonEmpty(oc.Nombre, rc.Nombre)
		dif.DiferenciaPorCandidato = append(dif.DiferenciaPorCandidato, CandidatoDiff{
			CandidatoID:  rc.CandidatoID,
			Nombre:       nombre,
			VotosOficial: oc.Votos,
			VotosRRV:     rc.Votos,
			Diferencia:   delta,
		})
		if delta != 0 {
			result.Inconsistencias = append(result.Inconsistencias, Inconsistencia{
				ActaID:      of.ActaID,
				Tipo:        TipoDiferenciaCandidato,
				Descripcion: fmt.Sprintf("Candidato %s (%s): oficial=%d RRV=%d (Δ%+d)", rc.CandidatoID, nombre, oc.Votos, rc.Votos, delta),
				Severidad:   severidadPorDelta(abs(delta)),
				FuenteAfectada: "AMBAS",
			})
		}
	}

	if len(result.Inconsistencias) == 0 {
		result.EstadoComparacion = EstadoConsistente
	} else {
		result.EstadoComparacion = EstadoInconsistente
		result.Diferencias = dif
	}
	return result
}

// soloOficial genera un ActaComparada para actas que no existen en RRV.
func soloOficial(of *ActaFuente) ActaComparada {
	return ActaComparada{
		ActaID:            of.ActaID,
		Departamento:      of.Departamento,
		Provincia:         of.Provincia,
		Municipio:         of.Municipio,
		Recinto:           of.Recinto,
		Mesa:              of.Mesa,
		EstadoComparacion: EstadoSoloOficial,
		DatosOficial:      of,
		Inconsistencias: []Inconsistencia{{
			ActaID:         of.ActaID,
			Tipo:           TipoActaNoExisteEnRRV,
			Descripcion:    "El acta existe en el sistema oficial pero no fue reportada por RRV/TREP",
			Severidad:      SeveridadMedia,
			FuenteAfectada: "RRV",
		}},
	}
}

// soloRRV genera un ActaComparada para actas que no existen en el sistema oficial.
func soloRRV(rv *ActaFuente) ActaComparada {
	return ActaComparada{
		ActaID:            rv.ActaID,
		Departamento:      rv.Departamento,
		Provincia:         rv.Provincia,
		Municipio:         rv.Municipio,
		Recinto:           rv.Recinto,
		Mesa:              rv.Mesa,
		EstadoComparacion: EstadoSoloRRV,
		DatosRRV:          rv,
		Inconsistencias: []Inconsistencia{{
			ActaID:         rv.ActaID,
			Tipo:           TipoActaNoExisteEnOficial,
			Descripcion:    "El acta fue reportada por RRV/TREP pero no existe en el sistema oficial",
			Severidad:      SeveridadAlta,
			FuenteAfectada: "OFICIAL",
		}},
	}
}

// ─── Resumen y KPIs ───────────────────────────────────────────────────────────

func buildResumen(actas []ActaComparada, totalOf, totalRRV int) ResumenComparacion {
	r := ResumenComparacion{
		FechaComparacion:     time.Now().UTC(),
		TotalActasRRV:        totalRRV,
		TotalActasOficial:    totalOf,
		TotalActasComparadas: len(actas),
		Actas:                actas,
	}

	var sumaDifAbs int
	for _, a := range actas {
		switch a.EstadoComparacion {
		case EstadoConsistente:
			r.ActasConsistentes++
		case EstadoInconsistente:
			r.ActasInconsistentes++
			if a.Diferencias != nil {
				sumaDifAbs += abs(a.Diferencias.DiferenciaTotalVotos)
			}
		case EstadoSoloOficial:
			r.ActasSoloOficial++
		case EstadoSoloRRV:
			r.ActasSoloRRV++
		}
	}
	r.DiferenciaTotalVotos = sumaDifAbs
	r.ConfiabilidadRRV = calcularConfiabilidad(r, sumaDifAbs)
	return r
}

// calcularConfiabilidad devuelve un score 0-100 según la fórmula:
//
//	confiabilidad = 100 - (pen_inconsistentes + pen_faltantes + pen_extras + pen_diferencia)
//
// Pesos: inconsistentes=50%, solo_oficial=30%, solo_rrv=20%. Bonus penalización
// por magnitud promedio de diferencia de votos (hasta 10 pts adicionales).
func calcularConfiabilidad(r ResumenComparacion, sumaDifAbs int) float64 {
	if r.TotalActasComparadas == 0 {
		return 100.0
	}
	total := float64(r.TotalActasComparadas)

	penInconsistentes := float64(r.ActasInconsistentes) / total * 50.0

	penSoloOficial := 0.0
	if r.TotalActasOficial > 0 {
		penSoloOficial = float64(r.ActasSoloOficial) / float64(r.TotalActasOficial) * 30.0
	}

	penSoloRRV := 0.0
	if r.TotalActasRRV > 0 {
		penSoloRRV = float64(r.ActasSoloRRV) / float64(r.TotalActasRRV) * 20.0
	}

	// Penalización por magnitud de diferencia (cap 10 pts a 50 votos promedio)
	penDiferencia := 0.0
	if r.ActasInconsistentes > 0 {
		promDif := float64(sumaDifAbs) / float64(r.ActasInconsistentes)
		penDiferencia = math.Min(promDif/50.0*10.0, 10.0)
	}

	penTotal := penInconsistentes + penSoloOficial + penSoloRRV + penDiferencia
	result := 100.0 - penTotal
	if result < 0 {
		result = 0
	}
	return math.Round(result*100) / 100
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

// estadoConflictivo detecta el caso en que un acta está anulada/observada en
// una fuente y validada/transcrita en la otra.
func estadoConflictivo(estadoOf, estadoRRV string) bool {
	anulados := map[string]bool{"anulada": true, "observada": true, "rechazada": true}
	validados := map[string]bool{"transcrita": true, "validada": true, "procesada": true}
	ofAnulada := anulados[strings.ToLower(estadoOf)]
	ofValida := validados[strings.ToLower(estadoOf)]
	rvAnulada := anulados[strings.ToLower(estadoRRV)]
	rvValida := validados[strings.ToLower(estadoRRV)]
	return (ofAnulada && rvValida) || (ofValida && rvAnulada)
}

func severidadPorDelta(delta int) Severidad {
	switch {
	case delta >= 20:
		return SeveridadAlta
	case delta >= 5:
		return SeveridadMedia
	default:
		return SeveridadBaja
	}
}
