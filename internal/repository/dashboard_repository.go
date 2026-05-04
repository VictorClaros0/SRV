package repository

import (
	"context"
	"time"

	"srrv/internal/models"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type DashboardRepository struct {
	actas          *mongo.Collection
	rrvActas       *mongo.Collection
	eventos        *mongo.Collection
	mesas          *mongo.Collection
	distribuciones *mongo.Collection
	recintos       *mongo.Collection
}

func NewDashboardRepository(db *mongo.Database) *DashboardRepository {
	return &DashboardRepository{
		actas:          db.Collection("actas"),
		rrvActas:       db.Collection("rrv_actas"),
		eventos:        db.Collection("rrv_eventos"),
		mesas:          db.Collection("mesas"),
		distribuciones: db.Collection("distribuciones_territoriales"),
		recintos:       db.Collection("recintos_electorales"),
	}
}

// ── Response types ────────────────────────────────────────────────────────────

type KPIResult struct {
	TotalActasEnBD           int64   `json:"totalActasEnBD"`
	TotalActasProcesadas     int64   `json:"totalActasProcesadas"`
	TotalActasInconsistentes int64   `json:"totalActasInconsistentes"`
	TotalMesas               int64   `json:"totalMesas"`
	VotosP1                  int64   `json:"votosP1"`
	VotosP2                  int64   `json:"votosP2"`
	VotosP3                  int64   `json:"votosP3"`
	VotosP4                  int64   `json:"votosP4"`
	VotosNulos               int64   `json:"votosNulos"`
	VotosBlancos             int64   `json:"votosBlancos"`
	VotosValidos             int64   `json:"votosValidos"`
	TotalHabilitados         int64   `json:"totalHabilitados"`
	PorcentajeParticipacion  float64 `json:"porcentajeParticipacion"`
}

type CandidatoVotos struct {
	Nombre     string  `json:"nombre"`
	Campo      string  `json:"campo"`
	Votos      int64   `json:"votos"`
	Porcentaje float64 `json:"porcentaje"`
}

type ParticipacionResult struct {
	TotalHabilitados        int64   `json:"totalHabilitados"`
	TotalVotantes           int64   `json:"totalVotantes"`
	PorcentajeParticipacion float64 `json:"porcentajeParticipacion"`
}

type GeoItem struct {
	Nombre        string  `json:"nombre"`
	VotosP1       int64   `json:"votosP1"`
	VotosP2       int64   `json:"votosP2"`
	VotosP3       int64   `json:"votosP3"`
	VotosP4       int64   `json:"votosP4"`
	VotosNulos    int64   `json:"votosNulos"`
	VotosValidos  int64   `json:"votosValidos"`
	Habilitados   int64   `json:"habilitados"`
	Participacion float64 `json:"participacion"`
}

type HeatmapItem struct {
	Nombre string  `json:"nombre"`
	Valor  float64 `json:"valor"`
}

type TransparenciaItem struct {
	CodigoMesa   string `json:"codigoMesa"`
	Departamento string `json:"departamento"`
	Provincia    string `json:"provincia"`
	Municipio    string `json:"municipio"`
	Recinto      string `json:"recinto"`
	UrlImagen    string `json:"urlImagen"`
	P1           int    `json:"p1"`
	P2           int    `json:"p2"`
	P3           int    `json:"p3"`
	P4           int    `json:"p4"`
	VotosNulos   int    `json:"votosNulos"`
	VotosBlancos int    `json:"votosBlanco"`
	VotosValidos int    `json:"votosValidos"`
}

type TrazabilidadResult struct {
	Acta    *models.Acta    `json:"acta"`
	RRVActa *models.RRVActa `json:"rrvActa,omitempty"`
	Eventos []models.Evento `json:"eventos"`
}

type TecnicoResult struct {
	TotalRRVActas            int64            `json:"totalRRVActas"`
	TotalActasProcesadas     int64            `json:"totalActasProcesadas"`
	TotalActasInconsistentes int64            `json:"totalActasInconsistentes"`
	TotalEventos             int64            `json:"totalEventos"`
	EventosPorTipo           map[string]int64 `json:"eventosPorTipo"`
	ActasUltimas24h          int64            `json:"actasUltimas24h"`
}

type AnomaliasResult struct {
	Total     int64           `json:"total"`
	Anomalias []models.Evento `json:"anomalias"`
}

type RecintoFiltro struct {
	ID        string `json:"id"`
	RecintoID int    `json:"recintoId"`
	Nombre    string `json:"recinto"`
	Direccion string `json:"direccion"`
}

type MesaFiltro struct {
	ID                 string `json:"id"`
	Codigo             int    `json:"codigo"`
	Mesa               int    `json:"mesa"`
	CantidadHabilitada int    `json:"cantidadHabilitada"`
}

// ── KPIs ──────────────────────────────────────────────────────────────────────

func (r *DashboardRepository) GetKPIs(ctx context.Context) (*KPIResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	result := &KPIResult{}

	votePipeline := mongo.Pipeline{
		bson.D{{Key: "$group", Value: bson.M{
			"_id":          nil,
			"p1":           bson.M{"$sum": "$p1"},
			"p2":           bson.M{"$sum": "$p2"},
			"p3":           bson.M{"$sum": "$p3"},
			"p4":           bson.M{"$sum": "$p4"},
			"votosNulos":   bson.M{"$sum": "$votosNulos"},
			"votosBlanco":  bson.M{"$sum": "$votosBlanco"},
			"votosValidos": bson.M{"$sum": "$votosValidos"},
			"count":        bson.M{"$sum": 1},
		}}},
	}
	cursor, err := r.actas.Aggregate(ctx, votePipeline)
	if err != nil {
		return nil, err
	}
	var voteAgg []bson.M
	if err := cursor.All(ctx, &voteAgg); err != nil {
		return nil, err
	}
	if len(voteAgg) > 0 {
		v := voteAgg[0]
		result.TotalActasEnBD = toInt64(v["count"])
		result.VotosP1 = toInt64(v["p1"])
		result.VotosP2 = toInt64(v["p2"])
		result.VotosP3 = toInt64(v["p3"])
		result.VotosP4 = toInt64(v["p4"])
		result.VotosNulos = toInt64(v["votosNulos"])
		result.VotosBlancos = toInt64(v["votosBlanco"])
		result.VotosValidos = toInt64(v["votosValidos"])
	}

	estadoPipeline := mongo.Pipeline{
		bson.D{{Key: "$group", Value: bson.M{
			"_id":   "$estado",
			"count": bson.M{"$sum": 1},
		}}},
	}
	cursor2, err := r.rrvActas.Aggregate(ctx, estadoPipeline)
	if err != nil {
		return nil, err
	}
	var estadoAgg []bson.M
	if err := cursor2.All(ctx, &estadoAgg); err != nil {
		return nil, err
	}
	for _, item := range estadoAgg {
		estado, _ := item["_id"].(string)
		count := toInt64(item["count"])
		switch estado {
		case "PROCESADA":
			result.TotalActasProcesadas = count
		case "INCONSISTENTE":
			result.TotalActasInconsistentes = count
		}
	}

	mesaPipeline := mongo.Pipeline{
		bson.D{{Key: "$group", Value: bson.M{
			"_id":              nil,
			"totalHabilitados": bson.M{"$sum": "$cantidadHabilitada"},
			"totalMesas":       bson.M{"$sum": 1},
		}}},
	}
	cursor3, err := r.mesas.Aggregate(ctx, mesaPipeline)
	if err != nil {
		return nil, err
	}
	var mesaAgg []bson.M
	if err := cursor3.All(ctx, &mesaAgg); err != nil {
		return nil, err
	}
	if len(mesaAgg) > 0 {
		result.TotalHabilitados = toInt64(mesaAgg[0]["totalHabilitados"])
		result.TotalMesas = toInt64(mesaAgg[0]["totalMesas"])
	}

	if result.TotalHabilitados > 0 {
		result.PorcentajeParticipacion = float64(result.VotosValidos) / float64(result.TotalHabilitados) * 100
	}
	return result, nil
}

// ── Votos por candidato ───────────────────────────────────────────────────────

func (r *DashboardRepository) GetVotosCandidato(ctx context.Context) ([]CandidatoVotos, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	pipeline := mongo.Pipeline{
		bson.D{{Key: "$group", Value: bson.M{
			"_id":          nil,
			"p1":           bson.M{"$sum": "$p1"},
			"p2":           bson.M{"$sum": "$p2"},
			"p3":           bson.M{"$sum": "$p3"},
			"p4":           bson.M{"$sum": "$p4"},
			"votosValidos": bson.M{"$sum": "$votosValidos"},
		}}},
	}
	cursor, err := r.actas.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	var agg []bson.M
	if err := cursor.All(ctx, &agg); err != nil {
		return nil, err
	}
	if len(agg) == 0 {
		return []CandidatoVotos{}, nil
	}

	v := agg[0]
	total := toInt64(v["votosValidos"])

	defs := []struct{ nombre, campo, key string }{
		{"Daenerys Targaryen", "p1", "p1"},
		{"Sansa Stark", "p2", "p2"},
		{"Robert Baratheon", "p3", "p3"},
		{"Tyrion Lannister", "p4", "p4"},
	}

	result := make([]CandidatoVotos, 0, 4)
	for _, d := range defs {
		votos := toInt64(v[d.key])
		pct := 0.0
		if total > 0 {
			pct = float64(votos) / float64(total) * 100
		}
		result = append(result, CandidatoVotos{Nombre: d.nombre, Campo: d.campo, Votos: votos, Porcentaje: pct})
	}
	return result, nil
}

// ── Participacion ─────────────────────────────────────────────────────────────

func (r *DashboardRepository) GetParticipacion(ctx context.Context) (*ParticipacionResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cursor, err := r.actas.Aggregate(ctx, mongo.Pipeline{
		bson.D{{Key: "$group", Value: bson.M{
			"_id":      nil,
			"votantes": bson.M{"$sum": "$votosValidos"},
		}}},
	})
	if err != nil {
		return nil, err
	}
	var actasAgg []bson.M
	if err := cursor.All(ctx, &actasAgg); err != nil {
		return nil, err
	}
	var totalVotantes int64
	if len(actasAgg) > 0 {
		totalVotantes = toInt64(actasAgg[0]["votantes"])
	}

	cursor2, err := r.mesas.Aggregate(ctx, mongo.Pipeline{
		bson.D{{Key: "$group", Value: bson.M{
			"_id":         nil,
			"habilitados": bson.M{"$sum": "$cantidadHabilitada"},
		}}},
	})
	if err != nil {
		return nil, err
	}
	var mesasAgg []bson.M
	if err := cursor2.All(ctx, &mesasAgg); err != nil {
		return nil, err
	}
	var totalHabilitados int64
	if len(mesasAgg) > 0 {
		totalHabilitados = toInt64(mesasAgg[0]["habilitados"])
	}

	pct := 0.0
	if totalHabilitados > 0 {
		pct = float64(totalVotantes) / float64(totalHabilitados) * 100
	}
	return &ParticipacionResult{
		TotalHabilitados:        totalHabilitados,
		TotalVotantes:           totalVotantes,
		PorcentajeParticipacion: pct,
	}, nil
}

// ── Geografico ────────────────────────────────────────────────────────────────

func (r *DashboardRepository) GetGeografico(ctx context.Context, nivel, departamento, provincia, municipio string) ([]GeoItem, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	validNiveles := map[string]bool{"departamento": true, "provincia": true, "municipio": true, "recinto": true}
	if !validNiveles[nivel] {
		nivel = "departamento"
	}

	match := bson.M{}
	if departamento != "" {
		match["departamento"] = departamento
	}
	if provincia != "" {
		match["provincia"] = provincia
	}
	if municipio != "" {
		match["municipio"] = municipio
	}

	pipeline := mongo.Pipeline{}
	if len(match) > 0 {
		pipeline = append(pipeline, bson.D{{Key: "$match", Value: match}})
	}
	pipeline = append(pipeline,
		bson.D{{Key: "$group", Value: bson.M{
			"_id":          "$" + nivel,
			"p1":           bson.M{"$sum": "$p1"},
			"p2":           bson.M{"$sum": "$p2"},
			"p3":           bson.M{"$sum": "$p3"},
			"p4":           bson.M{"$sum": "$p4"},
			"votosNulos":   bson.M{"$sum": "$votosNulos"},
			"votosValidos": bson.M{"$sum": "$votosValidos"},
			"habilitados":  bson.M{"$sum": "$votantesHabilitados"},
		}}},
		bson.D{{Key: "$sort", Value: bson.M{"_id": 1}}},
	)

	cursor, err := r.actas.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	var raw []bson.M
	if err := cursor.All(ctx, &raw); err != nil {
		return nil, err
	}

	items := make([]GeoItem, 0, len(raw))
	for _, v := range raw {
		nombre, _ := v["_id"].(string)
		habilitados := toInt64(v["habilitados"])
		votosValidos := toInt64(v["votosValidos"])
		pct := 0.0
		if habilitados > 0 {
			pct = float64(votosValidos) / float64(habilitados) * 100
		}
		items = append(items, GeoItem{
			Nombre:        nombre,
			VotosP1:       toInt64(v["p1"]),
			VotosP2:       toInt64(v["p2"]),
			VotosP3:       toInt64(v["p3"]),
			VotosP4:       toInt64(v["p4"]),
			VotosNulos:    toInt64(v["votosNulos"]),
			VotosValidos:  votosValidos,
			Habilitados:   habilitados,
			Participacion: pct,
		})
	}
	return items, nil
}

// ── Heatmap ───────────────────────────────────────────────────────────────────

func (r *DashboardRepository) GetHeatmap(ctx context.Context, metric, nivel string) ([]HeatmapItem, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	validNiveles := map[string]bool{"departamento": true, "provincia": true, "municipio": true, "recinto": true}
	if !validNiveles[nivel] {
		nivel = "departamento"
	}
	validMetrics := map[string]bool{"participacion": true, "p1": true, "p2": true, "p3": true, "p4": true, "nulos": true}
	if !validMetrics[metric] {
		metric = "participacion"
	}

	pipeline := mongo.Pipeline{
		bson.D{{Key: "$group", Value: bson.M{
			"_id":          "$" + nivel,
			"p1":           bson.M{"$sum": "$p1"},
			"p2":           bson.M{"$sum": "$p2"},
			"p3":           bson.M{"$sum": "$p3"},
			"p4":           bson.M{"$sum": "$p4"},
			"votosNulos":   bson.M{"$sum": "$votosNulos"},
			"votosValidos": bson.M{"$sum": "$votosValidos"},
			"habilitados":  bson.M{"$sum": "$votantesHabilitados"},
		}}},
		bson.D{{Key: "$sort", Value: bson.M{"_id": 1}}},
	}

	cursor, err := r.actas.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	var raw []bson.M
	if err := cursor.All(ctx, &raw); err != nil {
		return nil, err
	}

	items := make([]HeatmapItem, 0, len(raw))
	for _, v := range raw {
		nombre, _ := v["_id"].(string)
		var valor float64
		switch metric {
		case "participacion":
			habilitados := toInt64(v["habilitados"])
			if habilitados > 0 {
				valor = float64(toInt64(v["votosValidos"])) / float64(habilitados) * 100
			}
		case "p1":
			valor = float64(toInt64(v["p1"]))
		case "p2":
			valor = float64(toInt64(v["p2"]))
		case "p3":
			valor = float64(toInt64(v["p3"]))
		case "p4":
			valor = float64(toInt64(v["p4"]))
		case "nulos":
			valor = float64(toInt64(v["votosNulos"]))
		}
		items = append(items, HeatmapItem{Nombre: nombre, Valor: valor})
	}
	return items, nil
}

// ── Transparencia ─────────────────────────────────────────────────────────────

func (r *DashboardRepository) GetTransparencia(ctx context.Context) ([]TransparenciaItem, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cursor, err := r.actas.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var actas []models.Acta
	if err := cursor.All(ctx, &actas); err != nil {
		return nil, err
	}

	items := make([]TransparenciaItem, 0, len(actas))
	for _, a := range actas {
		items = append(items, TransparenciaItem{
			CodigoMesa:   a.CodigoMesa,
			Departamento: a.Departamento,
			Provincia:    a.Provincia,
			Municipio:    a.Municipio,
			Recinto:      a.Recinto,
			UrlImagen:    "acta_" + a.CodigoMesa + ".pdf",
			P1:           a.P1,
			P2:           a.P2,
			P3:           a.P3,
			P4:           a.P4,
			VotosNulos:   a.VotosNulos,
			VotosBlancos: a.VotosBlanco,
			VotosValidos: a.VotosValidos,
		})
	}
	return items, nil
}

// ── Trazabilidad ──────────────────────────────────────────────────────────────

func (r *DashboardRepository) GetTrazabilidad(ctx context.Context, codigoActa string) (*TrazabilidadResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var acta models.Acta
	if err := r.actas.FindOne(ctx, bson.M{"codigoMesa": codigoActa}).Decode(&acta); err != nil {
		return nil, err
	}

	result := &TrazabilidadResult{Acta: &acta}

	var rrvActa models.RRVActa
	if err := r.rrvActas.FindOne(ctx, bson.M{"acta_id": codigoActa}).Decode(&rrvActa); err == nil {
		result.RRVActa = &rrvActa
	}

	cursor, err := r.eventos.Find(ctx, bson.M{"acta_id": codigoActa})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var eventos []models.Evento
	if err := cursor.All(ctx, &eventos); err != nil {
		return nil, err
	}
	if eventos == nil {
		eventos = []models.Evento{}
	}
	result.Eventos = eventos

	return result, nil
}

// ── Tecnico ───────────────────────────────────────────────────────────────────

func (r *DashboardRepository) GetTecnico(ctx context.Context) (*TecnicoResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	result := &TecnicoResult{EventosPorTipo: make(map[string]int64)}

	cursor, err := r.rrvActas.Aggregate(ctx, mongo.Pipeline{
		bson.D{{Key: "$group", Value: bson.M{
			"_id":   "$estado",
			"count": bson.M{"$sum": 1},
		}}},
	})
	if err != nil {
		return nil, err
	}
	var estadoAgg []bson.M
	if err := cursor.All(ctx, &estadoAgg); err != nil {
		return nil, err
	}
	for _, item := range estadoAgg {
		estado, _ := item["_id"].(string)
		count := toInt64(item["count"])
		result.TotalRRVActas += count
		switch estado {
		case "PROCESADA":
			result.TotalActasProcesadas = count
		case "INCONSISTENTE":
			result.TotalActasInconsistentes = count
		}
	}

	cursor2, err := r.eventos.Aggregate(ctx, mongo.Pipeline{
		bson.D{{Key: "$group", Value: bson.M{
			"_id":   "$tipo",
			"count": bson.M{"$sum": 1},
		}}},
	})
	if err != nil {
		return nil, err
	}
	var tipoAgg []bson.M
	if err := cursor2.All(ctx, &tipoAgg); err != nil {
		return nil, err
	}
	for _, item := range tipoAgg {
		tipo, _ := item["_id"].(string)
		count := toInt64(item["count"])
		result.EventosPorTipo[tipo] = count
		result.TotalEventos += count
	}

	since := time.Now().UTC().Add(-24 * time.Hour)
	count24h, err := r.rrvActas.CountDocuments(ctx, bson.M{"fecha_recepcion": bson.M{"$gte": since}})
	if err != nil {
		return nil, err
	}
	result.ActasUltimas24h = count24h

	return result, nil
}

// ── Anomalias ─────────────────────────────────────────────────────────────────

func (r *DashboardRepository) GetAnomalias(ctx context.Context) (*AnomaliasResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: -1}})
	cursor, err := r.eventos.Find(ctx, bson.M{"tipo": bson.M{"$ne": "ACTA_RECIBIDA"}}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var anomalias []models.Evento
	if err := cursor.All(ctx, &anomalias); err != nil {
		return nil, err
	}
	if anomalias == nil {
		anomalias = []models.Evento{}
	}
	return &AnomaliasResult{Total: int64(len(anomalias)), Anomalias: anomalias}, nil
}

// ── Filtros cascada ───────────────────────────────────────────────────────────

func (r *DashboardRepository) GetDepartamentos(ctx context.Context) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return r.distinctStrings(ctx, r.distribuciones, "departamento", bson.M{})
}

func (r *DashboardRepository) GetProvincias(ctx context.Context, departamento string) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	filter := bson.M{}
	if departamento != "" {
		filter["departamento"] = departamento
	}
	return r.distinctStrings(ctx, r.distribuciones, "provincia", filter)
}

func (r *DashboardRepository) GetMunicipios(ctx context.Context, provincia string) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	filter := bson.M{}
	if provincia != "" {
		filter["provincia"] = provincia
	}
	return r.distinctStrings(ctx, r.distribuciones, "municipio", filter)
}

func (r *DashboardRepository) GetRecintosByMunicipio(ctx context.Context, municipio string) ([]RecintoFiltro, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.M{}
	if municipio != "" {
		filter["municipio"] = municipio
	}
	cursor, err := r.distribuciones.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	var distribs []models.DistribucionTerritorial
	if err := cursor.All(ctx, &distribs); err != nil {
		return nil, err
	}
	if len(distribs) == 0 {
		return []RecintoFiltro{}, nil
	}

	ids := make([]bson.ObjectID, 0, len(distribs))
	for _, d := range distribs {
		ids = append(ids, d.ID)
	}

	recCursor, err := r.recintos.Find(ctx, bson.M{"idDistribucionTerritorial": bson.M{"$in": ids}})
	if err != nil {
		return nil, err
	}
	var recs []models.RecintoElectoral
	if err := recCursor.All(ctx, &recs); err != nil {
		return nil, err
	}

	result := make([]RecintoFiltro, 0, len(recs))
	for _, rec := range recs {
		result = append(result, RecintoFiltro{
			ID:        rec.ID.Hex(),
			RecintoID: rec.RecintoID,
			Nombre:    rec.Recinto,
			Direccion: rec.Direccion,
		})
	}
	return result, nil
}

func (r *DashboardRepository) GetMesasByRecinto(ctx context.Context, recintoIDHex string) ([]MesaFiltro, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	recintoID, err := bson.ObjectIDFromHex(recintoIDHex)
	if err != nil {
		return nil, err
	}

	cursor, err := r.mesas.Find(ctx, bson.M{"idRecintoElectoral": recintoID})
	if err != nil {
		return nil, err
	}
	var mesas []models.Mesa
	if err := cursor.All(ctx, &mesas); err != nil {
		return nil, err
	}

	result := make([]MesaFiltro, 0, len(mesas))
	for _, m := range mesas {
		result = append(result, MesaFiltro{
			ID:                 m.ID.Hex(),
			Codigo:             m.Codigo,
			Mesa:               m.Mesa,
			CantidadHabilitada: m.CantidadHabilitada,
		})
	}
	return result, nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func (r *DashboardRepository) distinctStrings(ctx context.Context, col *mongo.Collection, field string, filter bson.M) ([]string, error) {
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: filter}},
		bson.D{{Key: "$group", Value: bson.M{"_id": "$" + field}}},
		bson.D{{Key: "$sort", Value: bson.M{"_id": 1}}},
	}
	cursor, err := col.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	var raw []bson.M
	if err := cursor.All(ctx, &raw); err != nil {
		return nil, err
	}
	result := make([]string, 0, len(raw))
	for _, v := range raw {
		if s, ok := v["_id"].(string); ok {
			result = append(result, s)
		}
	}
	return result, nil
}

func toInt64(v interface{}) int64 {
	switch val := v.(type) {
	case int32:
		return int64(val)
	case int64:
		return val
	case float64:
		return int64(val)
	case int:
		return int64(val)
	default:
		return 0
	}
}
