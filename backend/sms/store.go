package sms

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Store maneja las operaciones MongoDB del módulo SMS.
type Store struct {
	client           *mongo.Client
	dbName           string
	actasCollection  string
	eventsCollection string
	configCollection string
	defaultNumbers   []string
}

// NewStore crea un Store. actasCollection debe ser la primera colección de actas RRV.
func NewStore(client *mongo.Client, dbName, actasCollection, eventsCollection, configCollection string, defaultNumbers []string) *Store {
	if actasCollection == "" {
		actasCollection = "actas_rrv"
	}
	if eventsCollection == "" {
		eventsCollection = "rrv_eventos"
	}
	if configCollection == "" {
		configCollection = "sms_config"
	}
	return &Store{
		client:           client,
		dbName:           dbName,
		actasCollection:  actasCollection,
		eventsCollection: eventsCollection,
		configCollection: configCollection,
		defaultNumbers:   defaultNumbers,
	}
}

// ActaSMS es el documento que se inserta en actas_rrv para un SMS.
type ActaSMS struct {
	ActaID         string        `bson:"acta_id" json:"acta_id"`
	Mesa           string        `bson:"mesa" json:"mesa"`
	Candidatos     []CandidatoSMS `bson:"candidatos" json:"candidatos"`
	VotosNulos     int           `bson:"votos_nulos" json:"votos_nulos"`
	VotosBlancos   int           `bson:"votos_blancos" json:"votos_blancos"`
	TotalVotos     int           `bson:"total_votos" json:"total_votos"`
	Estado         string        `bson:"estado" json:"estado"`
	Fuente         string        `bson:"fuente" json:"fuente"`
	Remitente      string        `bson:"remitente" json:"remitente"`
	RawPayload     interface{}   `bson:"raw_payload" json:"raw_payload"`
	SMSBody        string        `bson:"sms_body" json:"sms_body"`
	DatosParsed    interface{}   `bson:"datos_parseados" json:"datos_parseados"`
	TotalDeclarado int           `bson:"total_declarado" json:"total_declarado"`
	TotalCalculado int           `bson:"total_calculado" json:"total_calculado"`
	Duplicado      bool          `bson:"duplicado" json:"duplicado"`
	Observaciones  string        `bson:"observaciones,omitempty" json:"observaciones,omitempty"`
	FechaRecepcion time.Time     `bson:"fecha_recepcion" json:"fecha_recepcion"`
	FechaProcesado time.Time     `bson:"fecha_procesado" json:"fecha_procesado"`
}

// CandidatoSMS representa los votos de un candidato en el acta SMS.
type CandidatoSMS struct {
	CandidatoID string `bson:"candidato_id"`
	Nombre      string `bson:"nombre"`
	Votos       int    `bson:"votos"`
}

type smsConfigDoc struct {
	ID                 string   `bson:"_id"`
	NumerosAutorizados []string `bson:"numeros_autorizados"`
}

// GetAuthorizedNumbers devuelve la lista de números normalizados autorizados.
// Si no existe documento en MongoDB, retorna los defaults de config.
func (s *Store) GetAuthorizedNumbers(ctx context.Context) ([]string, error) {
	col := s.client.Database(s.dbName).Collection(s.configCollection)
	var doc smsConfigDoc
	err := col.FindOne(ctx, bson.M{"_id": "config"}).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		return s.defaultNumbers, nil
	}
	if err != nil {
		return nil, fmt.Errorf("leer sms_config: %w", err)
	}
	return doc.NumerosAutorizados, nil
}

// AddAuthorizedNumber agrega un número normalizado a la lista autorizada.
// En la primera llamada inicializa el documento con los defaults + el nuevo número.
func (s *Store) AddAuthorizedNumber(ctx context.Context, numero string) error {
	col := s.client.Database(s.dbName).Collection(s.configCollection)

	var existing smsConfigDoc
	err := col.FindOne(ctx, bson.M{"_id": "config"}).Decode(&existing)
	if err == mongo.ErrNoDocuments {
		// Primera vez: arrancar desde los defaults + el número nuevo
		nums := deduplicar(append(s.defaultNumbers, numero))
		_, err = col.InsertOne(ctx, smsConfigDoc{ID: "config", NumerosAutorizados: nums})
		return err
	}
	if err != nil {
		return fmt.Errorf("leer sms_config: %w", err)
	}
	_, err = col.UpdateOne(ctx,
		bson.M{"_id": "config"},
		bson.M{"$addToSet": bson.M{"numeros_autorizados": numero}},
	)
	return err
}

// RemoveAuthorizedNumber elimina un número normalizado de la lista autorizada.
func (s *Store) RemoveAuthorizedNumber(ctx context.Context, numero string) error {
	col := s.client.Database(s.dbName).Collection(s.configCollection)

	var existing smsConfigDoc
	err := col.FindOne(ctx, bson.M{"_id": "config"}).Decode(&existing)
	if err == mongo.ErrNoDocuments {
		// No existe doc aún: crear con defaults menos el número a eliminar
		nums := deduplicar(s.defaultNumbers)
		filtered := make([]string, 0, len(nums))
		for _, n := range nums {
			if n != numero {
				filtered = append(filtered, n)
			}
		}
		_, err = col.InsertOne(ctx, smsConfigDoc{ID: "config", NumerosAutorizados: filtered})
		return err
	}
	if err != nil {
		return fmt.Errorf("leer sms_config: %w", err)
	}
	_, err = col.UpdateOne(ctx,
		bson.M{"_id": "config"},
		bson.M{"$pull": bson.M{"numeros_autorizados": numero}},
	)
	return err
}

// IsAuthorized comprueba si el número normalizado está en la lista.
func (s *Store) IsAuthorized(ctx context.Context, normalizado string) (bool, error) {
	numeros, err := s.GetAuthorizedNumbers(ctx)
	if err != nil {
		return false, err
	}
	for _, n := range numeros {
		if n == normalizado {
			return true, nil
		}
	}
	return false, nil
}

// GetExistingActa devuelve el acta previa si ya existe una para esa mesa enviada por SMS.
func (s *Store) GetExistingActa(ctx context.Context, mesa string) (*ActaSMS, error) {
	col := s.client.Database(s.dbName).Collection(s.actasCollection)
	var acta ActaSMS
	err := col.FindOne(ctx, bson.M{"mesa": mesa, "fuente": "SMS"}).Decode(&acta)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &acta, nil
}

// InsertActa guarda el acta SMS en la colección de actas RRV.
func (s *Store) InsertActa(ctx context.Context, acta ActaSMS) error {
	col := s.client.Database(s.dbName).Collection(s.actasCollection)
	_, err := col.InsertOne(ctx, acta)
	return err
}

// InsertEvent registra un evento de auditoría en rrv_eventos.
// Los errores de inserción se ignoran para no afectar el flujo principal.
func (s *Store) InsertEvent(ctx context.Context, actaID, tipo string, payload interface{}, errMsg string) {
	col := s.client.Database(s.dbName).Collection(s.eventsCollection)
	_, _ = col.InsertOne(ctx, bson.M{
		"acta_id":   actaID,
		"tipo":      tipo,
		"fuente":    "SMS",
		"payload":   payload,
		"error":     errMsg,
		"timestamp": time.Now(),
	})
}

func deduplicar(nums []string) []string {
	seen := make(map[string]bool, len(nums))
	out := make([]string, 0, len(nums))
	for _, n := range nums {
		if n != "" && !seen[n] {
			seen[n] = true
			out = append(out, n)
		}
	}
	return out
}

// GetRecentSMS obtiene las últimas 'limit' actas recibidas por SMS.
func (s *Store) GetRecentSMS(ctx context.Context, limit int) ([]ActaSMS, error) {
	col := s.client.Database(s.dbName).Collection(s.actasCollection)
	
	// Filtramos por fuente = SMS y ordenamos por fecha descendente
	filter := bson.M{"fuente": "SMS"}
	opts := options.Find().SetSort(bson.D{{Key: "fecha_recepcion", Value: -1}}).SetLimit(int64(limit))
	
	cursor, err := col.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("error consultando historial sms: %w", err)
	}
	defer cursor.Close(ctx)
	
	var actas []ActaSMS
	if err := cursor.All(ctx, &actas); err != nil {
		return nil, fmt.Errorf("error decodificando historial sms: %w", err)
	}
	
	if actas == nil {
		actas = []ActaSMS{}
	}
	return actas, nil
}
