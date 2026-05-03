package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ResultadosHandler endpoints de consulta de resultados y auditoría.
type ResultadosHandler struct {
	DB *gorm.DB
}

type resumenGeneral struct {
	TotalActas       int64 `gorm:"column:total_actas" json:"totalActas"`
	ActasTranscritas int64 `gorm:"column:actas_transcritas" json:"actasTranscritas"`
	ActasObservadas  int64 `gorm:"column:actas_observadas" json:"actasObservadas"`
	ActasImpresa     int64 `gorm:"column:actas_impresa" json:"actasImpresa"`
}

type comparativaItem struct {
	Candidato string `json:"candidato"`
	Votos     int64  `json:"votos"`
}

type resultadoPorUbicacion struct {
	Ubicacion    string `gorm:"column:ubicacion" json:"ubicacion"`
	P1           int64  `gorm:"column:p1" json:"p1"`
	P2           int64  `gorm:"column:p2" json:"p2"`
	P3           int64  `gorm:"column:p3" json:"p3"`
	P4           int64  `gorm:"column:p4" json:"p4"`
	VotosValidos int64  `gorm:"column:votos_validos" json:"votosValidos"`
	VotosNulos   int64  `gorm:"column:votos_nulos" json:"votosNulos"`
	VotosBlanco  int64  `gorm:"column:votos_blanco" json:"votosBlanco"`
	TotalActas   int64  `gorm:"column:total_actas" json:"totalActas"`
}

// Resultados devuelve resumen general, comparativa global por candidato
// y desglose por departamento. Solo contabiliza actas con estado='transcrita'.
func (h *ResultadosHandler) Resultados(c *gin.Context) {
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

	var totales struct {
		P1 int64 `gorm:"column:p1"`
		P2 int64 `gorm:"column:p2"`
		P3 int64 `gorm:"column:p3"`
		P4 int64 `gorm:"column:p4"`
	}
	h.DB.Raw(`
		SELECT
			COALESCE(SUM(p1),0) AS p1, COALESCE(SUM(p2),0) AS p2,
			COALESCE(SUM(p3),0) AS p3, COALESCE(SUM(p4),0) AS p4
		FROM acta
		WHERE fecha_eliminado IS NULL AND estado = 'transcrita'
	`).Scan(&totales)

	comparativa := []comparativaItem{
		{Candidato: "P1", Votos: totales.P1},
		{Candidato: "P2", Votos: totales.P2},
		{Candidato: "P3", Votos: totales.P3},
		{Candidato: "P4", Votos: totales.P4},
	}

	var porDepartamento []resultadoPorUbicacion
	h.DB.Raw(`
		SELECT
			dt.departamento AS ubicacion,
			COALESCE(SUM(a.p1),0) AS p1,
			COALESCE(SUM(a.p2),0) AS p2,
			COALESCE(SUM(a.p3),0) AS p3,
			COALESCE(SUM(a.p4),0) AS p4,
			COALESCE(SUM(a.votos_validos),0) AS votos_validos,
			COALESCE(SUM(a.votos_nulos),0) AS votos_nulos,
			COALESCE(SUM(a.votos_blanco),0) AS votos_blanco,
			COUNT(*) AS total_actas
		FROM acta a
		JOIN recinto_electoral re ON a.codigo_recinto = re.recinto_id
		JOIN distribucion_territorial dt ON re.id_distribucion_territorial = dt.id
		WHERE a.fecha_eliminado IS NULL AND a.estado = 'transcrita'
		GROUP BY dt.departamento
		ORDER BY dt.departamento
	`).Scan(&porDepartamento)

	c.JSON(http.StatusOK, gin.H{
		"resumen":         resumen,
		"comparativa":     comparativa,
		"porDepartamento": porDepartamento,
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

type heatmapItem struct {
	Ubicacion        string `gorm:"column:ubicacion" json:"ubicacion"`
	CandidatoGanador string `gorm:"column:candidato_ganador" json:"candidatoGanador"`
	Votos            int64  `gorm:"column:votos" json:"votos"`
}

// ResultadosPorRecintos devuelve votos agrupados por recinto electoral.
func (h *ResultadosHandler) ResultadosPorRecintos(c *gin.Context) {
	var rows []resultadoPorUbicacion
	h.DB.Raw(`
		SELECT re.recinto AS ubicacion,
			COALESCE(SUM(a.p1),0) AS p1, COALESCE(SUM(a.p2),0) AS p2,
			COALESCE(SUM(a.p3),0) AS p3, COALESCE(SUM(a.p4),0) AS p4,
			COUNT(*) AS total_actas
		FROM acta a
		JOIN recinto_electoral re ON a.codigo_recinto = re.recinto_id
		WHERE a.fecha_eliminado IS NULL AND a.estado = 'transcrita'
		GROUP BY re.recinto_id, re.recinto
		ORDER BY re.recinto
	`).Scan(&rows)
	c.JSON(http.StatusOK, rows)
}

// ResultadosPorProvincias devuelve votos agrupados por provincia.
func (h *ResultadosHandler) ResultadosPorProvincias(c *gin.Context) {
	var rows []resultadoPorUbicacion
	h.DB.Raw(`
		SELECT dt.provincia AS ubicacion,
			COALESCE(SUM(a.p1),0) AS p1, COALESCE(SUM(a.p2),0) AS p2,
			COALESCE(SUM(a.p3),0) AS p3, COALESCE(SUM(a.p4),0) AS p4,
			COUNT(*) AS total_actas
		FROM acta a
		JOIN recinto_electoral re ON a.codigo_recinto = re.recinto_id
		JOIN distribucion_territorial dt ON re.id_distribucion_territorial = dt.id
		WHERE a.fecha_eliminado IS NULL AND a.estado = 'transcrita'
		GROUP BY dt.provincia
		ORDER BY dt.provincia
	`).Scan(&rows)
	c.JSON(http.StatusOK, rows)
}

// HeatmapDepartamentos devuelve el candidato ganador y sus votos por departamento.
func (h *ResultadosHandler) HeatmapDepartamentos(c *gin.Context) {
	var rows []heatmapItem
	h.DB.Raw(`
		SELECT dt.departamento AS ubicacion,
			CASE
				WHEN SUM(p1) >= SUM(p2) AND SUM(p1) >= SUM(p3) AND SUM(p1) >= SUM(p4) THEN 'P1'
				WHEN SUM(p2) >= SUM(p3) AND SUM(p2) >= SUM(p4) THEN 'P2'
				WHEN SUM(p3) >= SUM(p4) THEN 'P3'
				ELSE 'P4'
			END AS candidato_ganador,
			GREATEST(SUM(p1), SUM(p2), SUM(p3), SUM(p4)) AS votos
		FROM acta a
		JOIN recinto_electoral re ON a.codigo_recinto = re.recinto_id
		JOIN distribucion_territorial dt ON re.id_distribucion_territorial = dt.id
		WHERE a.fecha_eliminado IS NULL AND a.estado = 'transcrita'
		GROUP BY dt.departamento
		ORDER BY dt.departamento
	`).Scan(&rows)
	c.JSON(http.StatusOK, rows)
}

// Auditoria devuelve las actas con estado='observada': actas que fueron recibidas
// desde el Excel con observaciones y que por eso no se contabilizan en resultados.
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
