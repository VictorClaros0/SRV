package comparacion

import "time"

// ─── Fuentes crudas ──────────────────────────────────────────────────────────

// CandidatoVotos representa los votos de un candidato dentro de un acta.
type CandidatoVotos struct {
	CandidatoID string `json:"candidato_id"`
	Nombre      string `json:"nombre"`
	Votos       int    `json:"votos"`
}

// ActaFuente es la representación normalizada de un acta, independientemente
// de si viene de MongoDB (RRV) o de PostgreSQL (Oficial).
type ActaFuente struct {
	ActaID       string           `json:"acta_id"`
	Departamento string           `json:"departamento"`
	Provincia    string           `json:"provincia"`
	Municipio    string           `json:"municipio"`
	Recinto      string           `json:"recinto"`
	Mesa         string           `json:"mesa"`
	Candidatos   []CandidatoVotos `json:"candidatos"`
	VotosNulos   int              `json:"votos_nulos"`
	VotosBlancos int              `json:"votos_blancos"`
	TotalVotos   int              `json:"total_votos"`
	Estado       string           `json:"estado"`
	Fuente       string           `json:"fuente"` // "RRV" | "OFICIAL"
}

// ─── Resultados de comparación ───────────────────────────────────────────────

// DiferenciaActa captura la diferencia numérica en votos entre ambas fuentes.
type DiferenciaActa struct {
	DiferenciaTotalVotos   int              `json:"diferencia_total_votos"`
	DiferenciaVotosNulos   int              `json:"diferencia_votos_nulos"`
	DiferenciaVotosBlancos int              `json:"diferencia_votos_blancos"`
	DiferenciaPorCandidato []CandidatoDiff  `json:"diferencia_por_candidato,omitempty"`
}

// CandidatoDiff muestra la diferencia de votos de un candidato específico.
type CandidatoDiff struct {
	CandidatoID  string `json:"candidato_id"`
	Nombre       string `json:"nombre"`
	VotosOficial int    `json:"votos_oficial"`
	VotosRRV     int    `json:"votos_rrv"`
	Diferencia   int    `json:"diferencia"` // RRV - Oficial (negativo = RRV tiene menos)
}

// TipoInconsistencia describe el tipo de discrepancia detectada.
type TipoInconsistencia string

const (
	TipoActaNoExisteEnOficial TipoInconsistencia = "ACTA_NO_EXISTE_EN_OFICIAL"
	TipoActaNoExisteEnRRV     TipoInconsistencia = "ACTA_NO_EXISTE_EN_RRV"
	TipoDiferenciaTotalVotos  TipoInconsistencia = "DIFERENCIA_TOTAL_VOTOS"
	TipoDiferenciaCandidato   TipoInconsistencia = "DIFERENCIA_CANDIDATO"
	TipoDiferenciaNulos       TipoInconsistencia = "DIFERENCIA_NULOS"
	TipoDiferenciaBlancos     TipoInconsistencia = "DIFERENCIA_BLANCOS"
	TipoEstadoConflictivo     TipoInconsistencia = "ESTADO_ACTA_CONFLICTIVO"
)

// Severidad indica el nivel de gravedad de una inconsistencia.
type Severidad string

const (
	SeveridadAlta  Severidad = "ALTA"
	SeveridadMedia Severidad = "MEDIA"
	SeveridadBaja  Severidad = "BAJA"
)

// Inconsistencia describe una discrepancia concreta entre las dos fuentes.
type Inconsistencia struct {
	ActaID        string             `json:"acta_id"`
	Tipo          TipoInconsistencia `json:"tipo_inconsistencia"`
	Descripcion   string             `json:"descripcion"`
	Severidad     Severidad          `json:"severidad"`
	FuenteAfectada string            `json:"fuente_afectada"` // "RRV" | "OFICIAL" | "AMBAS"
}

// EstadoComparacion clasifica el resultado de comparar un par de actas.
type EstadoComparacion string

const (
	EstadoConsistente  EstadoComparacion = "CONSISTENTE"
	EstadoInconsistente EstadoComparacion = "INCONSISTENTE"
	EstadoSoloRRV      EstadoComparacion = "SOLO_RRV"
	EstadoSoloOficial  EstadoComparacion = "SOLO_OFICIAL"
)

// ActaComparada es el resultado completo de comparar un acta entre ambas fuentes.
type ActaComparada struct {
	ActaID             string            `json:"acta_id"`
	Departamento       string            `json:"departamento"`
	Provincia          string            `json:"provincia"`
	Municipio          string            `json:"municipio"`
	Recinto            string            `json:"recinto"`
	Mesa               string            `json:"mesa"`
	EstadoComparacion  EstadoComparacion `json:"estado_comparacion"`
	DatosRRV           *ActaFuente       `json:"datos_rrv,omitempty"`
	DatosOficial       *ActaFuente       `json:"datos_oficial,omitempty"`
	Diferencias        *DiferenciaActa   `json:"diferencias,omitempty"`
	Inconsistencias    []Inconsistencia  `json:"inconsistencias,omitempty"`
}

// ─── Resumen global ───────────────────────────────────────────────────────────

// ResumenComparacion agrega los KPIs globales de una corrida de comparación.
type ResumenComparacion struct {
	FechaComparacion      time.Time       `json:"fecha_comparacion"`
	TotalActasRRV         int             `json:"total_actas_rrv"`
	TotalActasOficial     int             `json:"total_actas_oficial"`
	TotalActasComparadas  int             `json:"total_actas_comparadas"`
	ActasConsistentes     int             `json:"actas_consistentes"`
	ActasInconsistentes   int             `json:"actas_inconsistentes"`
	ActasSoloRRV          int             `json:"actas_solo_rrv"`
	ActasSoloOficial      int             `json:"actas_solo_oficial"`
	ConfiabilidadRRV      float64         `json:"confiabilidad_rrv"`       // 0–100
	DiferenciaTotalVotos  int             `json:"diferencia_total_votos"`  // suma absoluta de diferencias
	Actas                 []ActaComparada `json:"actas,omitempty"`         // omitido en /resumen
}

// ─── Respuesta paginada ───────────────────────────────────────────────────────

// RespuestaComparacion es el envelope de la respuesta del endpoint principal.
type RespuestaComparacion struct {
	Resumen  ResumenComparacion `json:"resumen"`
	Actas    []ActaComparada    `json:"actas"`
	Total    int                `json:"total"`
	Pagina   int                `json:"pagina"`
	PorPagina int               `json:"por_pagina"`
}
