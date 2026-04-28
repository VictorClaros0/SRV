package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"srrv/internal/models"
	"srrv/internal/repository"
	"srrv/internal/services"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// validPINs is the set of accepted SMS security PINs (pre-distributed to mesa officials).
// Additional PINs can be injected via env var VALID_PINS=pin1,pin2,...
var validPINs = map[string]bool{
	"1234": true,
	"5678": true,
	"9012": true,
	"2025": true,
	"4321": true,
}

type RRVHandler struct {
	actaRepo   *repository.RRVActaRepository
	eventoRepo *repository.EventoRepository
}

func NewRRVHandler(ar *repository.RRVActaRepository, er *repository.EventoRepository) *RRVHandler {
	if extra := os.Getenv("VALID_PINS"); extra != "" {
		for _, pin := range strings.Split(extra, ",") {
			if p := strings.TrimSpace(pin); p != "" {
				validPINs[p] = true
			}
		}
	}
	return &RRVHandler{actaRepo: ar, eventoRepo: er}
}

// registrarEvento persists an audit event. Errors are only logged, never propagated.
func (h *RRVHandler) registrarEvento(actaID, tipo, fuente string, payload bson.M, errMsg string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ev := &models.Evento{
		ActaID:    actaID,
		Tipo:      tipo,
		Fuente:    fuente,
		Payload:   payload,
		Error:     errMsg,
		Timestamp: time.Now(),
	}
	if err := h.eventoRepo.Create(ctx, ev); err != nil {
		log.Printf("⚠️  Error registrando evento %s para acta %s: %v", tipo, actaID, err)
	}
}

// validarAritmetica comprueba que la suma de votos coincide con total_votos.
// Devuelve "" si es consistente, o un mensaje de error si no.
func validarAritmetica(acta *models.RRVActa) string {
	suma := acta.VotosNulos + acta.VotosBlancos
	for _, c := range acta.Candidatos {
		if c.Votos < 0 {
			return fmt.Sprintf("votos negativos en candidato %s", c.CandidatoID)
		}
		suma += c.Votos
	}
	if acta.VotosNulos < 0 {
		return "votos_nulos no puede ser negativo"
	}
	if acta.VotosBlancos < 0 {
		return "votos_blancos no puede ser negativo"
	}
	if suma != acta.TotalVotos {
		return fmt.Sprintf("inconsistencia aritmética: suma=%d, total_votos=%d", suma, acta.TotalVotos)
	}
	return ""
}

// ──────────────────────────────────────────────────────────────────────────────
// POST /api/rrv/actas/upload
// ──────────────────────────────────────────────────────────────────────────────
// Acepta multipart/form-data con:
//   - file       : imagen o PDF del acta (obligatorio)
//   - acta_data  : JSON string con datos del acta (opcional; si ausente se simula OCR)
func (h *RRVHandler) Upload(c *gin.Context) {
	// 1. Leer el archivo
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "campo 'file' requerido"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo abrir el archivo"})
		return
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error leyendo el archivo"})
		return
	}

	// 2. Calcular hash para idempotencia
	hash := services.HashBytes(fileBytes)

	// 3. Verificar duplicado por hash (mismo archivo enviado dos veces)
	dupHash, err := h.actaRepo.ExistsByHash(c.Request.Context(), hash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if dupHash {
		h.registrarEvento("desconocido", "DUPLICADO_HASH", "IMAGEN",
			bson.M{"hash": hash, "filename": fileHeader.Filename},
			"archivo ya procesado anteriormente")
		c.JSON(http.StatusConflict, gin.H{"error": "este archivo ya fue procesado", "hash": hash})
		return
	}

	// 4. Obtener datos del acta (OCR real simulado o datos provistos)
	var acta *models.RRVActa
	if raw := c.PostForm("acta_data"); raw != "" {
		// Datos provistos manualmente (simula resultado de OCR externo)
		acta = &models.RRVActa{}
		if err := json.Unmarshal([]byte(raw), acta); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "acta_data JSON inválido: " + err.Error()})
			return
		}
		acta.TipoEntrada = "IMAGEN"
		acta.Fuente = "RRV"
	} else {
		// OCR simulado
		acta = services.SimulateOCR(fileBytes)
	}
	acta.HashOrigen = hash
	acta.FechaRecepcion = time.Now()

	// 5. Verificar duplicado por acta_id
	dupID, err := h.actaRepo.ExistsByActaID(c.Request.Context(), acta.ActaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if dupID {
		h.registrarEvento(acta.ActaID, "DUPLICADO_ACTA_ID", "IMAGEN",
			bson.M{"acta_id": acta.ActaID, "hash": hash},
			"acta_id ya existe en el sistema")
		c.JSON(http.StatusConflict, gin.H{"error": "acta ya registrada", "acta_id": acta.ActaID})
		return
	}

	// 6. Validar aritmética
	if msg := validarAritmetica(acta); msg != "" {
		acta.Estado = "INCONSISTENTE"
		h.registrarEvento(acta.ActaID, "INCONSISTENCIA_ARITMETICA", "IMAGEN",
			bson.M{"acta_id": acta.ActaID, "detalle": msg}, msg)
		log.Printf("⚠️  INCONSISTENCIA_ARITMETICA acta=%s: %s", acta.ActaID, msg)
	} else {
		acta.Estado = "PROCESADA"
	}

	// 7. Guardar en MongoDB
	saved, err := h.actaRepo.Create(c.Request.Context(), acta)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.registrarEvento(saved.ActaID, "ACTA_RECIBIDA", "IMAGEN",
		bson.M{"acta_id": saved.ActaID, "estado": saved.Estado, "hash": hash}, "")

	c.JSON(http.StatusCreated, saved)
}

// ──────────────────────────────────────────────────────────────────────────────
// POST /api/rrv/sms
// ──────────────────────────────────────────────────────────────────────────────
// Body JSON: { "mensaje": "ACTA:ACTA-001|DEP:La Paz|...|PIN:1234" }
func (h *RRVHandler) SMS(c *gin.Context) {
	var body struct {
		Mensaje string `json:"mensaje" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 1. Hash del mensaje completo para detectar mensajes idénticos
	msgHash := services.HashBytes([]byte(body.Mensaje))
	dupMsg, err := h.actaRepo.ExistsByHash(c.Request.Context(), msgHash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if dupMsg {
		h.registrarEvento("desconocido", "DUPLICADO_SMS", "SMS",
			bson.M{"hash": msgHash}, "mensaje SMS ya recibido anteriormente")
		c.JSON(http.StatusConflict, gin.H{"error": "mensaje SMS duplicado"})
		return
	}

	// 2. Parsear el mensaje
	acta, parseErr := parseSMS(body.Mensaje)
	if parseErr != "" {
		h.registrarEvento("desconocido", "SMS_INVALIDO", "SMS",
			bson.M{"mensaje": body.Mensaje, "detalle": parseErr}, parseErr)
		c.JSON(http.StatusBadRequest, gin.H{"error": parseErr})
		return
	}
	acta.HashOrigen = msgHash
	acta.FechaRecepcion = time.Now()

	// 3. Verificar PIN
	pin := extractField(body.Mensaje, "PIN")
	if !validPINs[pin] {
		h.registrarEvento(acta.ActaID, "PIN_INVALIDO", "SMS",
			bson.M{"acta_id": acta.ActaID}, "PIN de seguridad inválido")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "PIN inválido o no autorizado"})
		return
	}

	// 4. Verificar duplicado por acta_id
	dupID, err := h.actaRepo.ExistsByActaID(c.Request.Context(), acta.ActaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if dupID {
		h.registrarEvento(acta.ActaID, "DUPLICADO_ACTA_ID", "SMS",
			bson.M{"acta_id": acta.ActaID}, "acta_id ya existe en el sistema")
		c.JSON(http.StatusConflict, gin.H{"error": "acta ya registrada", "acta_id": acta.ActaID})
		return
	}

	// 5. Validar aritmética
	if msg := validarAritmetica(acta); msg != "" {
		acta.Estado = "INCONSISTENTE"
		h.registrarEvento(acta.ActaID, "INCONSISTENCIA_ARITMETICA", "SMS",
			bson.M{"acta_id": acta.ActaID, "detalle": msg}, msg)
		log.Printf("⚠️  INCONSISTENCIA_ARITMETICA acta=%s (SMS): %s", acta.ActaID, msg)
	} else {
		acta.Estado = "PROCESADA"
	}

	// 6. Guardar
	saved, err := h.actaRepo.Create(c.Request.Context(), acta)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.registrarEvento(saved.ActaID, "ACTA_RECIBIDA", "SMS",
		bson.M{"acta_id": saved.ActaID, "estado": saved.Estado}, "")

	c.JSON(http.StatusCreated, saved)
}

// ──────────────────────────────────────────────────────────────────────────────
// GET /api/rrv/actas
// ──────────────────────────────────────────────────────────────────────────────
func (h *RRVHandler) GetAll(c *gin.Context) {
	actas, err := h.actaRepo.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, actas)
}

// ──────────────────────────────────────────────────────────────────────────────
// GET /api/rrv/actas/:acta_id
// ──────────────────────────────────────────────────────────────────────────────
func (h *RRVHandler) GetByID(c *gin.Context) {
	actaID := c.Param("acta_id")
	acta, err := h.actaRepo.GetByActaID(c.Request.Context(), actaID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "acta no encontrada"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, acta)
}

// ──────────────────────────────────────────────────────────────────────────────
// GET /api/rrv/eventos
// ──────────────────────────────────────────────────────────────────────────────
func (h *RRVHandler) GetEventos(c *gin.Context) {
	eventos, err := h.eventoRepo.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, eventos)
}

// ─── Helpers de SMS ───────────────────────────────────────────────────────────

// extractField devuelve el valor de un campo clave:valor en el mensaje pipe-separado.
func extractField(msg, key string) string {
	for _, part := range strings.Split(msg, "|") {
		kv := strings.SplitN(part, ":", 2)
		if len(kv) == 2 && strings.EqualFold(strings.TrimSpace(kv[0]), key) {
			return strings.TrimSpace(kv[1])
		}
	}
	return ""
}

// parseSMS parsea el formato pipe-separado y devuelve un RRVActa o un mensaje de error.
// Formato: ACTA:id|DEP:dpto|MUN:mun|REC:recinto|MESA:mesa|C1:v|C2:v|...|NULOS:n|BLANCOS:b|TOTAL:t|PIN:pin
func parseSMS(msg string) (*models.RRVActa, string) {
	get := func(key string) string { return extractField(msg, key) }

	actaID := get("ACTA")
	dep := get("DEP")
	mun := get("MUN")
	rec := get("REC")
	mesa := get("MESA")
	nulosStr := get("NULOS")
	blancosStr := get("BLANCOS")
	totalStr := get("TOTAL")

	// Campos obligatorios de ubicación
	switch "" {
	case actaID:
		return nil, "campo ACTA requerido"
	case dep:
		return nil, "campo DEP requerido"
	case mun:
		return nil, "campo MUN requerido"
	case rec:
		return nil, "campo REC requerido"
	case mesa:
		return nil, "campo MESA requerido"
	case nulosStr:
		return nil, "campo NULOS requerido"
	case blancosStr:
		return nil, "campo BLANCOS requerido"
	case totalStr:
		return nil, "campo TOTAL requerido"
	}

	nulos, err := strconv.Atoi(nulosStr)
	if err != nil || nulos < 0 {
		return nil, "NULOS debe ser un entero no negativo"
	}
	blancos, err := strconv.Atoi(blancosStr)
	if err != nil || blancos < 0 {
		return nil, "BLANCOS debe ser un entero no negativo"
	}
	total, err := strconv.Atoi(totalStr)
	if err != nil || total < 0 {
		return nil, "TOTAL debe ser un entero no negativo"
	}

	// Extraer candidatos C1, C2, ... de forma dinámica
	var candidatos []models.Candidato
	for i := 1; ; i++ {
		vStr := get(fmt.Sprintf("C%d", i))
		if vStr == "" {
			break
		}
		v, err := strconv.Atoi(vStr)
		if err != nil || v < 0 {
			return nil, fmt.Sprintf("C%d debe ser un entero no negativo", i)
		}
		candidatos = append(candidatos, models.Candidato{
			CandidatoID: fmt.Sprintf("CAND-%02d", i),
			Nombre:      fmt.Sprintf("Candidato %d", i),
			Votos:       v,
		})
	}
	if len(candidatos) == 0 {
		return nil, "se requiere al menos un candidato (C1)"
	}

	acta := &models.RRVActa{
		ActaID:       actaID,
		Departamento: dep,
		Municipio:    mun,
		Recinto:      rec,
		Mesa:         mesa,
		Candidatos:   candidatos,
		VotosNulos:   nulos,
		VotosBlancos: blancos,
		TotalVotos:   total,
		Fuente:       "RRV",
		TipoEntrada:  "SMS",
	}
	return acta, ""
}
