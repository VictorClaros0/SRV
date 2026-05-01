package dashboard

import (
	"context"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/srvof/votos-backend/comparacion"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"gorm.io/gorm"
)

// Handler expone los endpoints del dashboard listos para Chart.js.
// Es de solo lectura: no escribe en ninguna base de datos (CQRS/Query).
type Handler struct {
	svc         *comparacion.Service
	comp        *comparacion.Comparer
	db          *gorm.DB
	mongoClient *mongo.Client
}

// NewHandler construye el Handler del dashboard.
func NewHandler(db *gorm.DB, mongoClient *mongo.Client, dbName string) *Handler {
	return &Handler{
		svc:         &comparacion.Service{DB: db, MongoClient: mongoClient, DBName: dbName},
		comp:        &comparacion.Comparer{},
		db:          db,
		mongoClient: mongoClient,
	}
}

// ─── GET /api/v1/dashboard/kpis ──────────────────────────────────────────────
//
// Alias exacto de /api/v1/comparacion/resumen.
// Devuelve los KPIs globales listos para tarjetas del dashboard.

func (h *Handler) KPIs(c *gin.Context) {
	resumen, err := h.runComparacion(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	resumen.Actas = nil
	c.JSON(http.StatusOK, resumen)
}

// ─── GET /api/v1/dashboard/inconsistencias ────────────────────────────────────
//
// Alias de /api/v1/comparacion/inconsistencias.
// Devuelve solo las actas problemáticas para la tabla de alertas.

func (h *Handler) Inconsistencias(c *gin.Context) {
	resumen, err := h.runComparacion(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var problematicas []comparacion.ActaComparada
	for _, a := range resumen.Actas {
		if a.EstadoComparacion != comparacion.EstadoConsistente {
			problematicas = append(problematicas, a)
		}
	}
	if problematicas == nil {
		problematicas = []comparacion.ActaComparada{}
	}

	resumen.Actas = nil
	c.JSON(http.StatusOK, gin.H{
		"resumen":           resumen,
		"inconsistencias":   problematicas,
		"total":             len(problematicas),
	})
}

// ─── GET /api/v1/dashboard/rrv-vs-oficial ─────────────────────────────────────
//
// Devuelve datos listos para un gráfico Chart.js de dona o barras que compara
// los 4 estados de comparación entre RRV y Oficial.

func (h *Handler) RRVvsOficial(c *gin.Context) {
	resumen, err := h.runComparacion(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"labels": []string{"Consistentes", "Inconsistentes", "Solo RRV", "Solo Oficial"},
		"datasets": []gin.H{
			{
				"label":            "Actas",
				"data":             []int{resumen.ActasConsistentes, resumen.ActasInconsistentes, resumen.ActasSoloRRV, resumen.ActasSoloOficial},
				"backgroundColor":  []string{"#22c55e", "#ef4444", "#f97316", "#94a3b8"},
			},
		},
		"raw": gin.H{
			"actas_consistentes":  resumen.ActasConsistentes,
			"actas_inconsistentes": resumen.ActasInconsistentes,
			"actas_solo_rrv":      resumen.ActasSoloRRV,
			"actas_solo_oficial":  resumen.ActasSoloOficial,
			"confiabilidad_rrv":   resumen.ConfiabilidadRRV,
		},
	})
}

// ─── GET /api/v1/dashboard/votos-candidato ────────────────────────────────────
//
// Devuelve votos totales por candidato comparando RRV vs Oficial.
// Formato listo para Chart.js grouped bar chart.

func (h *Handler) VotosCandidato(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	rrv, oficiales, err := h.fetchAmbas(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	type agg struct {
		nombre   string
		votosRRV int
		votosOf  int
	}

	totales := make(map[string]*agg)

	for _, acta := range rrv {
		for _, cand := range acta.Candidatos {
			if _, ok := totales[cand.CandidatoID]; !ok {
				totales[cand.CandidatoID] = &agg{nombre: cand.Nombre}
			}
			totales[cand.CandidatoID].votosRRV += cand.Votos
			// Prefer non-generic name from RRV
			if cand.Nombre != "" {
				totales[cand.CandidatoID].nombre = cand.Nombre
			}
		}
	}

	for _, acta := range oficiales {
		for _, cand := range acta.Candidatos {
			if _, ok := totales[cand.CandidatoID]; !ok {
				totales[cand.CandidatoID] = &agg{nombre: cand.Nombre}
			}
			totales[cand.CandidatoID].votosOf += cand.Votos
		}
	}

	// Sorted deterministic order
	ids := make([]string, 0, len(totales))
	for id := range totales {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	labels := make([]string, 0, len(ids))
	rrvData := make([]int, 0, len(ids))
	ofData := make([]int, 0, len(ids))

	for _, id := range ids {
		entry := totales[id]
		labels = append(labels, entry.nombre)
		rrvData = append(rrvData, entry.votosRRV)
		ofData = append(ofData, entry.votosOf)
	}

	c.JSON(http.StatusOK, gin.H{
		"labels": labels,
		"datasets": []gin.H{
			{"label": "RRV", "data": rrvData, "backgroundColor": "#f97316"},
			{"label": "Oficial", "data": ofData, "backgroundColor": "#3b82f6"},
		},
	})
}

// ─── GET /api/v1/dashboard/participacion ──────────────────────────────────────
//
// Calcula participación por departamento.
// Como la BD oficial no tiene ciudadanos_habilitados por departamento en un join
// funcional, devuelve available=false con un mensaje claro para el frontend.
// Si en el futuro el join se repara, este endpoint se activa sin cambiar el contrato.

func (h *Handler) Participacion(c *gin.Context) {
	// Intentar cálculo de participación global desde PostgreSQL
	type resultado struct {
		TotalHabilitados int64 `gorm:"column:total_habilitados"`
		TotalVotos       int64 `gorm:"column:total_votos"`
	}
	var res resultado
	err := h.db.Raw(`
		SELECT
			COALESCE(SUM(m.cantidad_habilitada), 0) AS total_habilitados,
			COALESCE(SUM(a.votos_validos + a.votos_nulos + a.votos_blanco), 0) AS total_votos
		FROM mesa m
		LEFT JOIN acta a ON a.id_mesa = m.id
		WHERE a.fecha_eliminado IS NULL OR a.fecha_eliminado IS NOT NULL
	`).Scan(&res).Error

	if err != nil || res.TotalHabilitados == 0 {
		c.JSON(http.StatusOK, gin.H{
			"available": false,
			"message":   "No hay datos suficientes para calcular participación (id_mesa no enlazado en actas)",
			"labels":    []string{},
			"datasets":  []gin.H{},
		})
		return
	}

	pct := float64(res.TotalVotos) / float64(res.TotalHabilitados) * 100

	c.JSON(http.StatusOK, gin.H{
		"available": true,
		"labels":    []string{"Global"},
		"datasets": []gin.H{
			{
				"label": "Participación %",
				"data":  []float64{round2(pct)},
			},
		},
		"raw": gin.H{
			"total_habilitados": res.TotalHabilitados,
			"total_votos":       res.TotalVotos,
		},
	})
}

// ─── GET /api/v1/dashboard/geografico ─────────────────────────────────────────
//
// Agrupa los resultados de comparación por departamento, municipio o recinto.
//
// Query params:
//
//	group_by  departamento | municipio | recinto  (default: departamento)
func (h *Handler) Geografico(c *gin.Context) {
	groupBy := strings.ToLower(strings.TrimSpace(c.Query("group_by")))
	if groupBy == "" {
		groupBy = "departamento"
	}
	if groupBy != "departamento" && groupBy != "municipio" && groupBy != "recinto" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "group_by debe ser: departamento, municipio o recinto",
		})
		return
	}

	resumen, err := h.runComparacion(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	type GeoItem struct {
		Nombre               string `json:"nombre"`
		TotalActas           int    `json:"total_actas"`
		Consistentes         int    `json:"consistentes"`
		Inconsistentes       int    `json:"inconsistentes"`
		SoloRRV              int    `json:"solo_rrv"`
		SoloOficial          int    `json:"solo_oficial"`
		DiferenciaTotalVotos int    `json:"diferencia_total_votos"`
	}

	grupos := make(map[string]*GeoItem)

	for _, acta := range resumen.Actas {
		key := ""
		switch groupBy {
		case "departamento":
			key = acta.Departamento
		case "municipio":
			key = acta.Municipio
		case "recinto":
			key = acta.Recinto
		}
		if key == "" {
			key = "(Sin clasificar)"
		}

		if _, ok := grupos[key]; !ok {
			grupos[key] = &GeoItem{Nombre: key}
		}
		item := grupos[key]
		item.TotalActas++
		switch acta.EstadoComparacion {
		case comparacion.EstadoConsistente:
			item.Consistentes++
		case comparacion.EstadoInconsistente:
			item.Inconsistentes++
			if acta.Diferencias != nil {
				d := acta.Diferencias.DiferenciaTotalVotos
				if d < 0 {
					d = -d
				}
				item.DiferenciaTotalVotos += d
			}
		case comparacion.EstadoSoloRRV:
			item.SoloRRV++
		case comparacion.EstadoSoloOficial:
			item.SoloOficial++
		}
	}

	items := make([]GeoItem, 0, len(grupos))
	for _, item := range grupos {
		items = append(items, *item)
	}
	sort.Slice(items, func(i, j int) bool {
		// "(Sin clasificar)" always last
		if items[i].Nombre == "(Sin clasificar)" {
			return false
		}
		if items[j].Nombre == "(Sin clasificar)" {
			return true
		}
		return items[i].Nombre < items[j].Nombre
	})

	c.JSON(http.StatusOK, gin.H{
		"group_by": groupBy,
		"total":    len(items),
		"items":    items,
	})
}

// ─── GET /api/v1/dashboard/tecnico ────────────────────────────────────────────
//
// Indicadores técnicos del sistema: conectividad a bases de datos, seguridad,
// y campos de latencia/throughput marcados como not_available para esta versión.

func (h *Handler) Tecnico(c *gin.Context) {
	// Verificar MongoDB
	mongoStatus := "connected"
	pingCtx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()
	if err := h.mongoClient.Ping(pingCtx, readpref.Primary()); err != nil {
		mongoStatus = "error"
	}

	// Verificar PostgreSQL
	pgStatus := "connected"
	sqlDB, err := h.db.DB()
	if err != nil {
		pgStatus = "error"
	} else if err := sqlDB.PingContext(c.Request.Context()); err != nil {
		pgStatus = "error"
	}

	// Conteos rápidos para indicadores básicos
	var totalActasPG int64
	h.db.Table("acta").Where("fecha_eliminado IS NULL").Count(&totalActasPG)

	var totalActasRRV int64
	col := h.mongoClient.Database(h.svc.DBName).Collection("actas_rrv")
	countCtx, countCancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer countCancel()
	totalActasRRV, _ = col.CountDocuments(countCtx, map[string]interface{}{})

	c.JSON(http.StatusOK, gin.H{
		"timestamp":            time.Now().UTC(),
		"latencia_ms":          nil,
		"throughput_actas_min": nil,
		"disponibilidad":       "not_available",
		"seguridad": gin.H{
			"jwt_required":    true,
			"https_requerido": false,
		},
		"fuentes": gin.H{
			"mongodb": gin.H{
				"estado":      mongoStatus,
				"total_actas": totalActasRRV,
				"coleccion":   "actas_rrv",
			},
			"postgresql": gin.H{
				"estado":      pgStatus,
				"total_actas": totalActasPG,
				"tabla":       "acta",
			},
		},
		"modulo_comparacion": gin.H{
			"cqrs_mode":       "query_only",
			"modifica_datos":  false,
			"version":         "1.0.0",
		},
	})
}

// ─── Helpers internos ─────────────────────────────────────────────────────────

func (h *Handler) fetchAmbas(ctx context.Context) (rrv, oficiales []comparacion.ActaFuente, err error) {
	rrv, err = h.svc.FetchRRV(ctx)
	if err != nil {
		return nil, nil, err
	}
	oficiales, err = h.svc.FetchOficiales(ctx)
	if err != nil {
		return nil, nil, err
	}
	return rrv, oficiales, nil
}

func (h *Handler) runComparacion(ctx context.Context) (comparacion.ResumenComparacion, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	rrv, oficiales, err := h.fetchAmbas(timeoutCtx)
	if err != nil {
		return comparacion.ResumenComparacion{}, err
	}
	return h.comp.Comparar(oficiales, rrv), nil
}

func round2(f float64) float64 {
	return float64(int(f*100+0.5)) / 100
}
