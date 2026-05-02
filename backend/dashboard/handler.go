package dashboard

import (
	"context"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/srvof/votos-backend/comparacion"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	eventsCollection string
}

// NewHandler construye el Handler del dashboard.
func NewHandler(db *gorm.DB, mongoClient *mongo.Client, dbName string, rrvCollections []string, eventsCollection string) *Handler {
	if eventsCollection == "" {
		eventsCollection = "rrv_eventos"
	}
	return &Handler{
		svc:              &comparacion.Service{DB: db, MongoClient: mongoClient, DBName: dbName, RRVCollections: rrvCollections},
		comp:             &comparacion.Comparer{},
		db:               db,
		mongoClient:      mongoClient,
		eventsCollection: eventsCollection,
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
// Calcula participación desde MongoDB (RRV) agrupada por departamento.
// Deriva el departamento de los primeros 2 dígitos del código de mesa (PNRE Bolivia).

func (h *Handler) Participacion(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	type docPart struct {
		Mesa         string `bson:"mesa"`
		Habilitados  int    `bson:"habilitados"`
		TotalVotos   int    `bson:"total_votos"`
		VotosNulos   int    `bson:"votos_nulos"`
		VotosBlancos int    `bson:"votos_blancos"`
		VotosBlanco  int    `bson:"votos_blanco"`
	}

	type deptData struct{ hab, vot int }
	byDept := map[string]*deptData{}
	var totalHab, totalVot int64

	for _, colName := range h.svc.RRVCollections {
		colName = strings.TrimSpace(colName)
		if colName == "" {
			continue
		}
		col := h.mongoClient.Database(h.svc.DBName).Collection(colName)
		cur, err := col.Find(ctx, bson.D{}, options.Find().SetProjection(bson.D{
			{Key: "mesa", Value: 1}, {Key: "habilitados", Value: 1},
			{Key: "total_votos", Value: 1}, {Key: "votos_nulos", Value: 1},
			{Key: "votos_blancos", Value: 1}, {Key: "votos_blanco", Value: 1},
		}))
		if err != nil {
			continue
		}
		for cur.Next(ctx) {
			var doc docPart
			if cur.Decode(&doc) != nil {
				continue
			}
			if doc.Habilitados == 0 {
				continue
			}
			blancos := doc.VotosBlancos
			if blancos == 0 {
				blancos = doc.VotosBlanco
			}
			votos := doc.TotalVotos
			if votos == 0 {
				votos = doc.VotosNulos + blancos
			}
			dept := departamentoFromMesa(doc.Mesa)
			if _, ok := byDept[dept]; !ok {
				byDept[dept] = &deptData{}
			}
			byDept[dept].hab += doc.Habilitados
			byDept[dept].vot += votos
			totalHab += int64(doc.Habilitados)
			totalVot += int64(votos)
		}
		_ = cur.Close(ctx)
	}

	if totalHab == 0 {
		// Fallback: PostgreSQL
		type pgRes struct {
			TotalHabilitados int64 `gorm:"column:total_habilitados"`
			TotalVotos       int64 `gorm:"column:total_votos"`
		}
		var pgr pgRes
		h.db.Raw(`SELECT COALESCE(SUM(m.cantidad_habilitada),0) AS total_habilitados,
			COALESCE(SUM(a.votos_validos+a.votos_nulos+a.votos_blanco),0) AS total_votos
			FROM mesa m LEFT JOIN acta a ON a.id_mesa=m.id`).Scan(&pgr)
		if pgr.TotalHabilitados > 0 {
			pct := round2(float64(pgr.TotalVotos) / float64(pgr.TotalHabilitados) * 100)
			c.JSON(http.StatusOK, gin.H{
				"available": true, "labels": []string{"Global"},
				"datasets": []gin.H{{"label": "Participación %", "data": []float64{pct}, "backgroundColor": "#6c5ce7"}},
				"raw":      gin.H{"total_habilitados": pgr.TotalHabilitados, "total_votos": pgr.TotalVotos},
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"available": false,
			"message":   "Sin datos de participación (escanea actas con electores habilitados para ver esta métrica)",
			"labels": []string{}, "datasets": []gin.H{},
		})
		return
	}

	depts := make([]string, 0, len(byDept))
	for d := range byDept {
		depts = append(depts, d)
	}
	sort.Strings(depts)

	labels := make([]string, 0, len(depts))
	data := make([]float64, 0, len(depts))
	for _, d := range depts {
		dd := byDept[d]
		if dd.hab == 0 {
			continue
		}
		labels = append(labels, d)
		data = append(data, round2(float64(dd.vot)/float64(dd.hab)*100))
	}

	c.JSON(http.StatusOK, gin.H{
		"available": true, "labels": labels,
		"datasets": []gin.H{{"label": "Participación %", "data": data, "backgroundColor": "#6c5ce7"}},
		"raw":      gin.H{"total_habilitados": totalHab, "total_votos": totalVot},
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
			if key == "" {
				key = departamentoFromMesa(acta.Mesa)
			}
		case "municipio":
			key = acta.Municipio
			if key == "" {
				key = departamentoFromMesa(acta.Mesa) + " (mesa " + acta.Mesa + ")"
			}
		case "recinto":
			key = acta.Recinto
			if key == "" && acta.Mesa != "" {
				key = "Mesa " + acta.Mesa
			}
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

	totalActasRRV := h.countRRVActas(c.Request.Context())

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
				"estado":       mongoStatus,
				"total_actas":  totalActasRRV,
				"colecciones":  h.svc.RRVCollections,
				"eventos_rrv":  h.eventsCollection,
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

// EventosRRV expone el audit log que produce SRRV2 en MongoDB.
func (h *Handler) EventosRRV(c *gin.Context) {
	type eventoMongo struct {
		ActaID    string                 `bson:"acta_id"`
		Tipo      string                 `bson:"tipo"`
		Fuente    string                 `bson:"fuente"`
		Payload   map[string]interface{} `bson:"payload"`
		Error     string                 `bson:"error"`
		Timestamp time.Time              `bson:"timestamp"`
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	col := h.mongoClient.Database(h.svc.DBName).Collection(h.eventsCollection)
	total, _ := col.CountDocuments(ctx, bson.D{})
	cur, err := col.Find(ctx, bson.D{}, options.Find().
		SetSort(bson.D{{Key: "timestamp", Value: -1}}).
		SetLimit(50),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cur.Close(ctx)

	eventos := []gin.H{}
	for cur.Next(ctx) {
		var raw eventoMongo
		if err := cur.Decode(&raw); err != nil {
			continue
		}
		descripcion := raw.Tipo
		if raw.Error != "" {
			descripcion = raw.Error
		}
		eventos = append(eventos, gin.H{
			"tipo_evento":  raw.Tipo,
			"acta_id":      raw.ActaID,
			"fuente":       raw.Fuente,
			"descripcion":  descripcion,
			"fecha_evento": raw.Timestamp,
			"payload":      raw.Payload,
		})
	}
	if err := cur.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"coleccion":     h.eventsCollection,
		"total_eventos": total,
		"eventos":       eventos,
	})
}

// ─── GET /api/v1/dashboard/mobile-scans ──────────────────────────────────────
//
// Lista los últimos scans recibidos desde la app móvil / scanner web (MongoDB).
// Devuelve id, mesa, estado, fuente, imagen_url, votos, fecha.

func (h *Handler) MobileScans(c *gin.Context) {
	type scanDoc struct {
		ActaID         string    `bson:"acta_id"          json:"acta_id"`
		Mesa           string    `bson:"mesa"             json:"mesa"`
		Estado         string    `bson:"estado"           json:"estado"`
		FuenteScan     string    `bson:"fuente_scan"      json:"fuente_scan"`
		ImagenURL      string    `bson:"imagen_url"       json:"imagen_url"`
		FechaRecepcion time.Time `bson:"fecha_recepcion"  json:"fecha_recepcion"`
		TotalVotos     int       `bson:"total_votos"      json:"total_votos"`
		VotosNulos     int       `bson:"votos_nulos"      json:"votos_nulos"`
		VotosBlancos   int       `bson:"votos_blancos"    json:"votos_blancos"`
		Habilitados    int       `bson:"habilitados"      json:"habilitados"`
		Candidatos     []struct {
			CandidatoID string `bson:"candidato_id" json:"candidato_id"`
			Nombre      string `bson:"nombre"       json:"nombre"`
			Votos       int    `bson:"votos"        json:"votos"`
		} `bson:"candidatos" json:"candidatos"`
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var all []scanDoc
	seen := map[string]bool{}
	collections := h.svc.RRVCollections
	if len(collections) == 0 {
		collections = []string{"actas_rrv"}
	}
	for _, colName := range collections {
		colName = strings.TrimSpace(colName)
		if colName == "" || seen[colName] {
			continue
		}
		seen[colName] = true
		col := h.mongoClient.Database(h.svc.DBName).Collection(colName)
		cur, err := col.Find(ctx, bson.D{},
			options.Find().SetSort(bson.D{{Key: "fecha_recepcion", Value: -1}}).SetLimit(50))
		if err != nil {
			continue
		}
		for cur.Next(ctx) {
			var doc scanDoc
			if cur.Decode(&doc) == nil {
				all = append(all, doc)
			}
		}
		_ = cur.Close(ctx)
	}

	// Ordenar por fecha desc y tomar los 30 más recientes
	sort.Slice(all, func(i, j int) bool {
		return all[i].FechaRecepcion.After(all[j].FechaRecepcion)
	})
	if len(all) > 30 {
		all = all[:30]
	}
	if all == nil {
		all = []scanDoc{}
	}

	// Agregar totales de votos por candidato para el dashboard
	type candTotales struct {
		Nombre string `json:"nombre"`
		Votos  int    `json:"votos"`
	}
	totalesCand := map[string]*candTotales{}
	for _, scan := range all {
		if scan.Estado != "validada" && scan.Estado != "pendiente_revision" {
			continue
		}
		for _, cand := range scan.Candidatos {
			if _, ok := totalesCand[cand.CandidatoID]; !ok {
				totalesCand[cand.CandidatoID] = &candTotales{Nombre: cand.Nombre}
			}
			totalesCand[cand.CandidatoID].Votos += cand.Votos
			if cand.Nombre != "" {
				totalesCand[cand.CandidatoID].Nombre = cand.Nombre
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"total":         len(all),
		"scans":         all,
		"votos_totales": totalesCand,
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

func (h *Handler) countRRVActas(ctx context.Context) int64 {
	countCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var total int64
	seen := map[string]bool{}
	collections := h.svc.RRVCollections
	if len(collections) == 0 {
		collections = []string{"actas_rrv"}
	}
	for _, collection := range collections {
		collection = strings.TrimSpace(collection)
		if collection == "" || seen[collection] {
			continue
		}
		seen[collection] = true
		col := h.mongoClient.Database(h.svc.DBName).Collection(collection)
		count, err := col.CountDocuments(countCtx, bson.D{})
		if err == nil {
			total += count
		}
	}
	return total
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

// departamentoFromMesa deriva el departamento de Bolivia a partir del código de mesa PNRE.
// Los primeros 2 dígitos del código de mesa identifican el departamento.
func departamentoFromMesa(mesa string) string {
	mesa = strings.TrimSpace(mesa)
	if len(mesa) < 2 {
		return "(Sin clasificar)"
	}
	prefix := mesa[:2]
	dept := map[string]string{
		"10": "Chuquisaca", "20": "La Paz", "30": "Cochabamba",
		"40": "Oruro", "50": "Potosí", "60": "Tarija",
		"70": "Santa Cruz", "80": "Beni", "90": "Pando",
	}
	if name, ok := dept[prefix]; ok {
		return name
	}
	return "(Sin clasificar)"
}
