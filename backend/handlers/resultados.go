package handlers

import (
	"math"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ResultadosHandler struct {
	DB *gorm.DB
}

type resumenGeneral struct {
	TotalActas       int64   `gorm:"column:total_actas" json:"totalActas"`
	ActasTranscritas int64   `gorm:"column:actas_transcritas" json:"actasTranscritas"`
	ActasObservadas  int64   `gorm:"column:actas_observadas" json:"actasObservadas"`
	ActasImpresa     int64   `gorm:"column:actas_impresa" json:"actasImpresa"`
	PctProcesadas    float64 `json:"pctProcesadas"`
}

type comparativaItem struct {
	Candidato string  `json:"candidato"`
	Votos     int64   `json:"votos"`
	Pct       float64 `json:"pct"`
}

type resultadoPorUbicacion struct {
	Ubicacion    string  `gorm:"column:ubicacion" json:"ubicacion"`
	P1           int64   `gorm:"column:p1" json:"p1"`
	P1Pct        float64 `gorm:"column:p1_pct" json:"p1_pct"`
	P2           int64   `gorm:"column:p2" json:"p2"`
	P2Pct        float64 `gorm:"column:p2_pct" json:"p2_pct"`
	P3           int64   `gorm:"column:p3" json:"p3"`
	P3Pct        float64 `gorm:"column:p3_pct" json:"p3_pct"`
	P4           int64   `gorm:"column:p4" json:"p4"`
	P4Pct        float64 `gorm:"column:p4_pct" json:"p4_pct"`
	VotosValidos int64   `gorm:"column:votos_validos" json:"votosValidos"`
	VotosNulos   int64   `gorm:"column:votos_nulos" json:"votosNulos"`
	VotosBlanco  int64   `gorm:"column:votos_blanco" json:"votosBlanco"`
	TotalActas   int64   `gorm:"column:total_actas" json:"totalActas"`
}

type velocidadStats struct {
	ActasUltimaHora int64 `gorm:"column:actas_ultima_hora" json:"actasUltimaHora"`
	ActasUltimas24h int64 `gorm:"column:actas_ultimas_24h" json:"actasUltimas24h"`
}

// Resultados devuelve resumen global + comparativa con % + desglose por nivel geográfico + velocidad.
func (h *ResultadosHandler) Resultados(c *gin.Context) {
	dept      := strings.TrimSpace(c.Query("departamento"))
	municipio := strings.TrimSpace(c.Query("municipio"))
	prov      := strings.TrimSpace(c.Query("provincia"))
	recinto   := strings.TrimSpace(c.Query("recinto"))
	mesa      := strings.TrimSpace(c.Query("mesa"))

	// Resumen global
	var resumen resumenGeneral
	h.DB.Raw(`
		SELECT
			COUNT(*) AS total_actas,
			SUM(CASE WHEN estado = 'transcrita' THEN 1 ELSE 0 END) AS actas_transcritas,
			SUM(CASE WHEN estado = 'observada'  THEN 1 ELSE 0 END) AS actas_observadas,
			SUM(CASE WHEN estado = 'impresa'    THEN 1 ELSE 0 END) AS actas_impresa
		FROM acta
		WHERE fecha_eliminado IS NULL
	`).Scan(&resumen)
	if resumen.TotalActas > 0 {
		resumen.PctProcesadas = math.Round(float64(resumen.ActasTranscritas)*10000/float64(resumen.TotalActas)) / 100
	}

	// WHERE dinámico
	where := "a.fecha_eliminado IS NULL AND a.estado = 'transcrita'"
	var args []interface{}
	if dept != "" {
		where += " AND dt.departamento = ?"
		args = append(args, dept)
	}
	if municipio != "" {
		where += " AND dt.municipio = ?"
		args = append(args, municipio)
	}
	if prov != "" {
		where += " AND dt.provincia = ?"
		args = append(args, prov)
	}
	if recinto != "" {
		where += " AND CAST(re.recinto_id AS TEXT) = ?"
		args = append(args, recinto)
	}
	if mesa != "" {
		where += " AND CAST(a.nro_mesa AS TEXT) = ?"
		args = append(args, mesa)
	}

	baseJoin := `
		FROM acta a
		JOIN recinto_electoral re
			ON LEFT(CAST(a.codigo_recinto AS TEXT), 5) = LEFT(CAST(re.recinto_id AS TEXT), 5)
		JOIN distribucion_territorial dt ON re.id_distribucion_territorial = dt.id
		WHERE ` + where

	// Totales filtrados para comparativa
	var totales struct {
		P1 int64 `gorm:"column:p1"`
		P2 int64 `gorm:"column:p2"`
		P3 int64 `gorm:"column:p3"`
		P4 int64 `gorm:"column:p4"`
	}
	h.DB.Raw(`
		SELECT COALESCE(SUM(a.p1),0) AS p1, COALESCE(SUM(a.p2),0) AS p2,
		       COALESCE(SUM(a.p3),0) AS p3, COALESCE(SUM(a.p4),0) AS p4
		`+baseJoin, args...).Scan(&totales)

	// Porcentajes por candidato
	totalVotos := totales.P1 + totales.P2 + totales.P3 + totales.P4
	pct := func(v int64) float64 {
		if totalVotos == 0 {
			return 0
		}
		return math.Round(float64(v)*10000/float64(totalVotos)) / 100
	}
	p1pct := pct(totales.P1)
	p2pct := pct(totales.P2)
	p3pct := pct(totales.P3)
	p4pct := pct(totales.P4)

	// Margen de victoria: diferencia entre los dos porcentajes más altos
	pcts := []float64{p1pct, p2pct, p3pct, p4pct}
	sort.Float64s(pcts)
	margen := math.Round((pcts[3]-pcts[2])*100) / 100

	comparativa := []comparativaItem{
		{"P1", totales.P1, p1pct},
		{"P2", totales.P2, p2pct},
		{"P3", totales.P3, p3pct},
		{"P4", totales.P4, p4pct},
	}

	// Nivel de desglose según la cascada de filtros
	var groupExpr, labelExpr string
	switch {
	case recinto != "" || mesa != "":
		groupExpr = "a.nro_mesa"
		labelExpr = "CAST(a.nro_mesa AS TEXT)"
	case prov != "":
		groupExpr = "re.recinto_id, re.recinto"
		labelExpr = "re.recinto"
	case municipio != "":
		groupExpr = "dt.provincia"
		labelExpr = "dt.provincia"
	case dept != "":
		groupExpr = "dt.municipio"
		labelExpr = "dt.municipio"
	default:
		groupExpr = "dt.departamento"
		labelExpr = "dt.departamento"
	}

	var desglose []resultadoPorUbicacion
	h.DB.Raw(`
		SELECT `+labelExpr+` AS ubicacion,
			COALESCE(SUM(a.p1),0) AS p1,
			CASE WHEN (COALESCE(SUM(a.p1),0)+COALESCE(SUM(a.p2),0)+COALESCE(SUM(a.p3),0)+COALESCE(SUM(a.p4),0)) > 0
			     THEN ROUND(COALESCE(SUM(a.p1),0)*100.0 / (COALESCE(SUM(a.p1),0)+COALESCE(SUM(a.p2),0)+COALESCE(SUM(a.p3),0)+COALESCE(SUM(a.p4),0)), 2)
			     ELSE 0 END AS p1_pct,
			COALESCE(SUM(a.p2),0) AS p2,
			CASE WHEN (COALESCE(SUM(a.p1),0)+COALESCE(SUM(a.p2),0)+COALESCE(SUM(a.p3),0)+COALESCE(SUM(a.p4),0)) > 0
			     THEN ROUND(COALESCE(SUM(a.p2),0)*100.0 / (COALESCE(SUM(a.p1),0)+COALESCE(SUM(a.p2),0)+COALESCE(SUM(a.p3),0)+COALESCE(SUM(a.p4),0)), 2)
			     ELSE 0 END AS p2_pct,
			COALESCE(SUM(a.p3),0) AS p3,
			CASE WHEN (COALESCE(SUM(a.p1),0)+COALESCE(SUM(a.p2),0)+COALESCE(SUM(a.p3),0)+COALESCE(SUM(a.p4),0)) > 0
			     THEN ROUND(COALESCE(SUM(a.p3),0)*100.0 / (COALESCE(SUM(a.p1),0)+COALESCE(SUM(a.p2),0)+COALESCE(SUM(a.p3),0)+COALESCE(SUM(a.p4),0)), 2)
			     ELSE 0 END AS p3_pct,
			COALESCE(SUM(a.p4),0) AS p4,
			CASE WHEN (COALESCE(SUM(a.p1),0)+COALESCE(SUM(a.p2),0)+COALESCE(SUM(a.p3),0)+COALESCE(SUM(a.p4),0)) > 0
			     THEN ROUND(COALESCE(SUM(a.p4),0)*100.0 / (COALESCE(SUM(a.p1),0)+COALESCE(SUM(a.p2),0)+COALESCE(SUM(a.p3),0)+COALESCE(SUM(a.p4),0)), 2)
			     ELSE 0 END AS p4_pct,
			COALESCE(SUM(a.votos_validos),0) AS votos_validos,
			COALESCE(SUM(a.votos_nulos),0)   AS votos_nulos,
			COALESCE(SUM(a.votos_blanco),0)  AS votos_blanco,
			COUNT(*) AS total_actas
		`+baseJoin+`
		GROUP BY `+groupExpr+`
		ORDER BY `+labelExpr, args...).Scan(&desglose)

	// Velocidad de procesamiento (global, no filtrada)
	var vel velocidadStats
	h.DB.Raw(`
		SELECT
			COALESCE(SUM(CASE WHEN fecha_modificacion >= NOW() - INTERVAL '1 hour'  THEN 1 ELSE 0 END), 0) AS actas_ultima_hora,
			COALESCE(SUM(CASE WHEN fecha_modificacion >= NOW() - INTERVAL '24 hours' THEN 1 ELSE 0 END), 0) AS actas_ultimas_24h
		FROM acta
		WHERE fecha_eliminado IS NULL AND estado = 'transcrita'
	`).Scan(&vel)

	c.JSON(http.StatusOK, gin.H{
		"resumen":         resumen,
		"comparativa":     comparativa,
		"margen_victoria": margen,
		"desglose":        desglose,
		"velocidad":       vel,
	})
}

type actaAuditoria struct {
	ID            uint      `gorm:"column:id" json:"id"`
	CodigoActa    int64     `gorm:"column:codigo_acta" json:"codigoActa"`
	CodigoRecinto int64     `gorm:"column:codigo_recinto" json:"codigoRecinto"`
	NroMesa       int       `gorm:"column:nro_mesa" json:"nroMesa"`
	Estado        string    `gorm:"column:estado" json:"estado"`
	Observaciones string    `gorm:"column:observaciones" json:"observaciones"`
	FechaCreacion time.Time `gorm:"column:fecha_creacion" json:"fechaCreacion"`
}

// Auditoria devuelve las actas con estado='observada'.
func (h *ResultadosHandler) Auditoria(c *gin.Context) {
	var actas []actaAuditoria
	h.DB.Raw(`
		SELECT id, codigo_acta, codigo_recinto, nro_mesa, estado, observaciones, fecha_creacion
		FROM acta
		WHERE fecha_eliminado IS NULL
		  AND estado = 'observada'
		ORDER BY codigo_acta
	`).Scan(&actas)

	c.JSON(http.StatusOK, gin.H{
		"total": len(actas),
		"actas": actas,
	})
}
