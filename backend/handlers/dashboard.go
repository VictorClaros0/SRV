package handlers

import (
	"math"
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DashboardHandler struct {
	DB *gorm.DB
}

// KPIs devuelve métricas globales calculadas.
func (h *DashboardHandler) KPIs(c *gin.Context) {
	var actas struct {
		TotalActas       int64 `gorm:"column:total_actas"`
		ActasTranscritas int64 `gorm:"column:actas_transcritas"`
		ActasObservadas  int64 `gorm:"column:actas_observadas"`
		ActasPendientes  int64 `gorm:"column:actas_pendientes"`
	}
	h.DB.Raw(`
		SELECT
			COUNT(*) AS total_actas,
			SUM(CASE WHEN estado = 'transcrita' THEN 1 ELSE 0 END) AS actas_transcritas,
			SUM(CASE WHEN estado = 'observada'  THEN 1 ELSE 0 END) AS actas_observadas,
			SUM(CASE WHEN estado = 'impresa'    THEN 1 ELSE 0 END) AS actas_pendientes
		FROM acta WHERE fecha_eliminado IS NULL
	`).Scan(&actas)

	var votos struct {
		P1           int64 `gorm:"column:p1"`
		P2           int64 `gorm:"column:p2"`
		P3           int64 `gorm:"column:p3"`
		P4           int64 `gorm:"column:p4"`
		VotosValidos int64 `gorm:"column:votos_validos"`
		VotosNulos   int64 `gorm:"column:votos_nulos"`
		VotosBlanco  int64 `gorm:"column:votos_blanco"`
	}
	h.DB.Raw(`
		SELECT
			COALESCE(SUM(p1),0) AS p1, COALESCE(SUM(p2),0) AS p2,
			COALESCE(SUM(p3),0) AS p3, COALESCE(SUM(p4),0) AS p4,
			COALESCE(SUM(votos_validos),0) AS votos_validos,
			COALESCE(SUM(votos_nulos),0)   AS votos_nulos,
			COALESCE(SUM(votos_blanco),0)  AS votos_blanco
		FROM acta WHERE fecha_eliminado IS NULL AND estado = 'transcrita'
	`).Scan(&votos)

	totalVotos := votos.P1 + votos.P2 + votos.P3 + votos.P4
	pct := func(v int64) float64 {
		if totalVotos == 0 {
			return 0
		}
		return math.Round(float64(v)*10000/float64(totalVotos)) / 100
	}

	type cand struct {
		Candidato  string  `json:"candidato"`
		Votos      int64   `json:"votos"`
		Porcentaje float64 `json:"porcentaje"`
	}
	candidatos := []cand{
		{"P1", votos.P1, pct(votos.P1)},
		{"P2", votos.P2, pct(votos.P2)},
		{"P3", votos.P3, pct(votos.P3)},
		{"P4", votos.P4, pct(votos.P4)},
	}
	sort.Slice(candidatos, func(i, j int) bool {
		return candidatos[i].Votos > candidatos[j].Votos
	})

	margen := 0.0
	if len(candidatos) >= 2 {
		margen = math.Round((candidatos[0].Porcentaje-candidatos[1].Porcentaje)*100) / 100
	}

	pctPublicadas := 0.0
	if actas.TotalActas > 0 {
		pctPublicadas = math.Round(float64(actas.ActasTranscritas)*10000/float64(actas.TotalActas)) / 100
	}

	c.JSON(http.StatusOK, gin.H{
		"totalActas":           actas.TotalActas,
		"actasTranscritas":     actas.ActasTranscritas,
		"actasObservadas":      actas.ActasObservadas,
		"actasPendientes":      actas.ActasPendientes,
		"porcentajePublicadas": pctPublicadas,
		"votosValidos":         votos.VotosValidos,
		"votosNulos":           votos.VotosNulos,
		"votosBlanco":          votos.VotosBlanco,
		"ganador":              candidatos[0],
		"segundo":              candidatos[1],
		"margenVictoria":       margen,
	})
}

type geoRow struct {
	Nombre       string  `gorm:"column:nombre" json:"nombre"`
	TotalActas   int64   `gorm:"column:total_actas" json:"totalActas"`
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
}

type geoOut struct {
	geoRow
	Ganador string  `json:"ganador"`
	Margen  float64 `json:"margen"`
}

func buildGeoQuery(labelExpr, groupExpr, where string) string {
	pctExpr := func(col string) string {
		return `CASE WHEN (COALESCE(SUM(a.p1),0)+COALESCE(SUM(a.p2),0)+COALESCE(SUM(a.p3),0)+COALESCE(SUM(a.p4),0)) > 0
			     THEN ROUND(` + col + `*100.0/(COALESCE(SUM(a.p1),0)+COALESCE(SUM(a.p2),0)+COALESCE(SUM(a.p3),0)+COALESCE(SUM(a.p4),0)),2)
			     ELSE 0 END`
	}
	return `SELECT ` + labelExpr + ` AS nombre,
		COUNT(*) AS total_actas,
		COALESCE(SUM(a.p1),0) AS p1, ` + pctExpr("COALESCE(SUM(a.p1),0)") + ` AS p1_pct,
		COALESCE(SUM(a.p2),0) AS p2, ` + pctExpr("COALESCE(SUM(a.p2),0)") + ` AS p2_pct,
		COALESCE(SUM(a.p3),0) AS p3, ` + pctExpr("COALESCE(SUM(a.p3),0)") + ` AS p3_pct,
		COALESCE(SUM(a.p4),0) AS p4, ` + pctExpr("COALESCE(SUM(a.p4),0)") + ` AS p4_pct,
		COALESCE(SUM(a.votos_validos),0) AS votos_validos,
		COALESCE(SUM(a.votos_nulos),0)   AS votos_nulos,
		COALESCE(SUM(a.votos_blanco),0)  AS votos_blanco
	FROM acta a
	JOIN recinto_electoral re ON LEFT(CAST(a.codigo_recinto AS TEXT),5) = LEFT(CAST(re.recinto_id AS TEXT),5)
	JOIN distribucion_territorial dt ON re.id_distribucion_territorial = dt.id
	WHERE ` + where + `
	GROUP BY ` + groupExpr + `
	ORDER BY ` + labelExpr
}

func enrichGeo(rows []geoRow) []geoOut {
	out := make([]geoOut, len(rows))
	for i, row := range rows {
		type cp struct {
			c string
			p float64
		}
		pcts := []cp{{"P1", row.P1Pct}, {"P2", row.P2Pct}, {"P3", row.P3Pct}, {"P4", row.P4Pct}}
		sort.Slice(pcts, func(a, b int) bool { return pcts[a].p > pcts[b].p })
		margen := 0.0
		ganador := ""
		if len(pcts) >= 1 {
			ganador = pcts[0].c
		}
		if len(pcts) >= 2 {
			margen = math.Round((pcts[0].p-pcts[1].p)*100) / 100
		}
		out[i] = geoOut{row, ganador, margen}
	}
	return out
}

// Geografico devuelve resultados agrupados por nivel geográfico con ganador y margen.
//
// Query params:
//
//	nivel = departamento (default) | municipio | provincia | recinto
//	departamento, municipio, provincia = filtros opcionales
func (h *DashboardHandler) Geografico(c *gin.Context) {
	nivel     := strings.TrimSpace(c.Query("nivel"))
	dept      := strings.TrimSpace(c.Query("departamento"))
	municipio := strings.TrimSpace(c.Query("municipio"))
	provincia := strings.TrimSpace(c.Query("provincia"))

	if nivel == "" {
		nivel = "departamento"
	}

	var labelExpr, groupExpr string
	switch nivel {
	case "municipio":
		labelExpr = "dt.municipio"
		groupExpr = "dt.municipio"
	case "provincia":
		labelExpr = "dt.provincia"
		groupExpr = "dt.provincia"
	case "recinto":
		labelExpr = "re.recinto"
		groupExpr = "re.recinto_id, re.recinto"
	default:
		labelExpr = "dt.departamento"
		groupExpr = "dt.departamento"
	}

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
	if provincia != "" {
		where += " AND dt.provincia = ?"
		args = append(args, provincia)
	}

	var rows []geoRow
	h.DB.Raw(buildGeoQuery(labelExpr, groupExpr, where), args...).Scan(&rows)

	c.JSON(http.StatusOK, gin.H{
		"nivel": nivel,
		"items": enrichGeo(rows),
	})
}

// Heatmap devuelve un valor escalar por región para visualización de calor.
//
// Query params:
//
//	metric = ganador (default) | inconsistencias | procesamiento
//	nivel  = departamento (default) | municipio | provincia
func (h *DashboardHandler) Heatmap(c *gin.Context) {
	metric := strings.TrimSpace(c.Query("metric"))
	nivel  := strings.TrimSpace(c.Query("nivel"))
	if nivel == "" {
		nivel = "departamento"
	}
	if metric == "" {
		metric = "ganador"
	}

	var labelExpr, groupExpr string
	switch nivel {
	case "municipio":
		labelExpr = "dt.municipio"
		groupExpr = "dt.municipio"
	case "provincia":
		labelExpr = "dt.provincia"
		groupExpr = "dt.provincia"
	default:
		labelExpr = "dt.departamento"
		groupExpr = "dt.departamento"
	}

	where := "a.fecha_eliminado IS NULL AND a.estado = 'transcrita'"
	var rows []geoRow
	h.DB.Raw(buildGeoQuery(labelExpr, groupExpr, where)).Scan(&rows)

	type heatItem struct {
		Nombre  string  `json:"nombre"`
		Valor   float64 `json:"valor"`
		Ganador string  `json:"ganador"`
		Votos   int64   `json:"votos"`
	}

	items := make([]heatItem, len(rows))
	for i, row := range rows {
		geo := enrichGeo([]geoRow{row})[0]
		var valor float64
		switch metric {
		case "inconsistencias":
			// porcentaje de actas observadas sobre total — necesita query separado
			valor = 0
		case "procesamiento":
			// actas transcritas / total (requiere datos de actas_impresa por zona — no disponible aquí)
			valor = 0
		default: // ganador
			valor = math.Max(math.Max(row.P1Pct, row.P2Pct), math.Max(row.P3Pct, row.P4Pct))
		}
		items[i] = heatItem{
			Nombre:  row.Nombre,
			Valor:   valor,
			Ganador: geo.Ganador,
			Votos:   row.VotosValidos,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"metric": metric,
		"nivel":  nivel,
		"items":  items,
	})
}

// Anomalias detecta inconsistencias aritméticas en actas transcritas.
func (h *DashboardHandler) Anomalias(c *gin.Context) {
	type anomalia struct {
		CodigoActa       int64  `gorm:"column:codigo_acta" json:"codigoActa"`
		NroMesa          int    `gorm:"column:nro_mesa" json:"nroMesa"`
		Tipo             string `json:"tipo"`
		Severidad        string `json:"severidad"`
		Detalle          string `json:"detalle"`
		ValoresEsperados int64  `json:"valoresEsperados,omitempty"`
		ValoresReales    int64  `json:"valoresReales,omitempty"`
	}

	var inconsistentes []struct {
		CodigoActa   int64 `gorm:"column:codigo_acta"`
		NroMesa      int   `gorm:"column:nro_mesa"`
		Suma         int64 `gorm:"column:suma"`
		VotosValidos int64 `gorm:"column:votos_validos"`
	}
	h.DB.Raw(`
		SELECT codigo_acta, nro_mesa,
		       (p1 + p2 + p3 + p4) AS suma,
		       votos_validos
		FROM acta
		WHERE fecha_eliminado IS NULL AND estado = 'transcrita'
		  AND (p1 + p2 + p3 + p4) <> votos_validos
		ORDER BY codigo_acta
		LIMIT 500
	`).Scan(&inconsistentes)

	var ceroTotal []struct {
		CodigoActa int64 `gorm:"column:codigo_acta"`
		NroMesa    int   `gorm:"column:nro_mesa"`
	}
	h.DB.Raw(`
		SELECT codigo_acta, nro_mesa
		FROM acta
		WHERE fecha_eliminado IS NULL AND estado = 'transcrita'
		  AND (p1 + p2 + p3 + p4) = 0
		ORDER BY codigo_acta
		LIMIT 200
	`).Scan(&ceroTotal)

	anomalias := []anomalia{}
	for _, a := range inconsistentes {
		anomalias = append(anomalias, anomalia{
			CodigoActa:       a.CodigoActa,
			NroMesa:          a.NroMesa,
			Tipo:             "votos_validos_inconsistentes",
			Severidad:        "alta",
			Detalle:          "p1+p2+p3+p4 != votos_validos",
			ValoresEsperados: a.Suma,
			ValoresReales:    a.VotosValidos,
		})
	}
	for _, a := range ceroTotal {
		anomalias = append(anomalias, anomalia{
			CodigoActa: a.CodigoActa,
			NroMesa:    a.NroMesa,
			Tipo:       "acta_sin_votos",
			Severidad:  "media",
			Detalle:    "todos los candidatos tienen 0 votos",
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"total":     len(anomalias),
		"anomalias": anomalias,
	})
}
