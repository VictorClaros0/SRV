package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"srrv/internal/models"
	"srrv/internal/repository"
	"srrv/internal/services"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var camposRemitente = []string{"from", "sender", "phone", "number", "source", "From"}
var camposCuerpo = []string{"body", "message", "text", "content", "sms", "Body"}

type SMSInboundHandler struct {
	rrvActaRepo *repository.RRVActaRepository
	actaRepo    *repository.ActaRepository
	eventoRepo  *repository.EventoRepository
	smsRepo     *repository.SMSConfigRepository
}

func NewSMSInboundHandler(rrvRepo *repository.RRVActaRepository, actaRepo *repository.ActaRepository, er *repository.EventoRepository, sr *repository.SMSConfigRepository) *SMSInboundHandler {
	return &SMSInboundHandler{rrvActaRepo: rrvRepo, actaRepo: actaRepo, eventoRepo: er, smsRepo: sr}
}

func (h *SMSInboundHandler) registrarEventoSMS(ctx context.Context, actaID, tipo string, payload bson.M, errMsg string) {
	ev := &models.Evento{
		ActaID:    actaID,
		Tipo:      tipo,
		Fuente:    "SMS",
		Payload:   payload,
		Error:     errMsg,
		Timestamp: time.Now(),
	}
	if err := h.eventoRepo.Create(ctx, ev); err != nil {
		log.Printf("⚠️ Error registrando evento SMS %s: %v", tipo, err)
	}
}

// InboundSMS recibe el webhook de SMS Forwarder y procesa el mensaje.
// Siempre responde HTTP 200 para evitar reintentos de la app.
func (h *SMSInboundHandler) InboundSMS(c *gin.Context) {
	ctx := c.Request.Context()
	timestamp := time.Now()

	rawBody, _ := io.ReadAll(c.Request.Body)
	c.Request.Body = io.NopCloser(bytes.NewBuffer(rawBody))

	var rawPayload map[string]interface{}
	if err := json.Unmarshal(rawBody, &rawPayload); err != nil {
		rawPayload = make(map[string]interface{})
		_ = c.Request.ParseForm()
		for k, v := range c.Request.Form {
			if len(v) > 0 {
				rawPayload[k] = v[0]
			}
		}
	}

	remitente := extraerCampoSMS(rawPayload, camposRemitente)
	smsBody := extraerCampoSMS(rawPayload, camposCuerpo)

	// Formato Forward SMS: "Desde : 65707079\nRRV|MESA=..."
	if remitente == "" && smsBody != "" {
		lines := strings.Split(smsBody, "\n")
		if len(lines) > 0 {
			firstLine := strings.TrimSpace(lines[0])
			if strings.HasPrefix(strings.ToLower(firstLine), "desde") {
				parts := strings.SplitN(firstLine, ":", 2)
				if len(parts) == 2 {
					remitente = strings.TrimSpace(parts[1])
					if len(lines) > 1 {
						smsBody = strings.Join(lines[1:], "\n")
					} else {
						smsBody = ""
					}
				}
			}
		}
	}

	normalizado := normalizarNumeroSMS(remitente)
	log.Printf("[SMS-INBOUND] ts=%s remitente=%q norm=%q cuerpo=%q",
		timestamp.Format(time.RFC3339), remitente, normalizado, smsBody)

	// Verificar autorización
	autorizado, err := h.smsRepo.IsAuthorized(ctx, normalizado)
	if err != nil {
		log.Printf("[SMS-INBOUND] error verificando autorización: %v", err)
	}
	if !autorizado {
		log.Printf("[SMS-INBOUND] ADVERTENCIA: remitente no autorizado remitente=%q norm=%q", remitente, normalizado)
		h.registrarEventoSMS(ctx, "", "SMS_REMITENTE_NO_AUTORIZADO",
			bson.M{"remitente": remitente, "normalizado": normalizado}, "remitente no autorizado")
		c.JSON(http.StatusOK, gin.H{"ok": false, "reason": "sender_not_authorized"})
		return
	}

	// Parsear cuerpo RRV|MESA=...|P1=...|...
	datos, parseErr := parseSMSRRV(smsBody)
	if parseErr != nil {
		actaID := fmt.Sprintf("SMS-RECHAZADA-%d", timestamp.UnixMilli())
		h.registrarEventoSMS(ctx, actaID, "SMS_ACTA_RECHAZADA",
			bson.M{"remitente": normalizado, "sms_body": smsBody}, parseErr.Error())
		log.Printf("[SMS-INBOUND] acta rechazada acta_id=%s razon=%v", actaID, parseErr)
		c.JSON(http.StatusOK, gin.H{
			"ok":      true,
			"estado":  "RECHAZADA",
			"acta_id": actaID,
			"razon":   parseErr.Error(),
		})
		return
	}

	// Calcular total y determinar estado
	totalCalculado := datos.P1 + datos.P2 + datos.P3 + datos.P4 + datos.Blancos + datos.Nulos
	estado := "PROCESADA"
	if totalCalculado != datos.Total {
		estado = "INCONSISTENTE"
		h.registrarEventoSMS(ctx, "", "SMS_INCONSISTENCIA_ARITMETICA",
			bson.M{"mesa": datos.Mesa, "total_declarado": datos.Total, "total_calculado": totalCalculado},
			"Suma de votos no coincide con total declarado")
	}

	// Verificar duplicado por mesa
	existing, err := h.rrvActaRepo.GetBySMSMesa(ctx, datos.Mesa)
	if err != nil {
		log.Printf("[SMS-INBOUND] error buscando duplicado mesa=%s: %v", datos.Mesa, err)
	}

	duplicadoFlag := false
	if existing != nil {
		// Comprobar idempotencia (mismo mensaje repetido)
		esIdentico := existing.TotalVotos == datos.Total &&
			existing.VotosBlancos == datos.Blancos &&
			existing.VotosNulos == datos.Nulos
		if esIdentico {
			for _, cand := range existing.Candidatos {
				switch cand.CandidatoID {
				case "CAND-01":
					if cand.Votos != datos.P1 {
						esIdentico = false
					}
				case "CAND-02":
					if cand.Votos != datos.P2 {
						esIdentico = false
					}
				case "CAND-03":
					if cand.Votos != datos.P3 {
						esIdentico = false
					}
				case "CAND-04":
					if cand.Votos != datos.P4 {
						esIdentico = false
					}
				}
			}
		}
		if esIdentico {
			log.Printf("[SMS-INBOUND] Idempotencia: SMS idéntico ignorado para mesa %s", datos.Mesa)
			c.JSON(http.StatusOK, gin.H{
				"ok":      true,
				"estado":  existing.Estado,
				"acta_id": existing.ActaID,
				"nota":    "idempotent_retry",
			})
			return
		}
		duplicadoFlag = true
		estado = "INCONSISTENTE"
		log.Printf("[SMS-INBOUND] ADVERTENCIA: datos contradictorios mesa=%s", datos.Mesa)
		h.registrarEventoSMS(ctx, "", "SMS_DATOS_CONTRADICTORIOS",
			bson.M{"mesa": datos.Mesa, "nuevo_total": datos.Total, "viejo_total": existing.TotalVotos},
			"Acta recibida con anterioridad y no coinciden los datos")
	}

	actaID := fmt.Sprintf("SMS-%s-%d", datos.Mesa, timestamp.UnixMilli())

	acta := &models.RRVActa{
		ActaID: actaID,
		Mesa:   datos.Mesa,
		Candidatos: []models.Candidato{
			{CandidatoID: "CAND-01", Nombre: "Candidato P1", Votos: datos.P1},
			{CandidatoID: "CAND-02", Nombre: "Candidato P2", Votos: datos.P2},
			{CandidatoID: "CAND-03", Nombre: "Candidato P3", Votos: datos.P3},
			{CandidatoID: "CAND-04", Nombre: "Candidato P4", Votos: datos.P4},
		},
		VotosNulos:     datos.Nulos,
		VotosBlancos:   datos.Blancos,
		TotalVotos:     datos.Total,
		Estado:         estado,
		Fuente:         "SMS",
		TipoEntrada:    "SMS_INBOUND",
		HashOrigen:     services.HashBytes([]byte(smsBody)),
		FechaRecepcion: timestamp,
		Remitente:      normalizado,
		SMSBody:        smsBody,
		Duplicado:      duplicadoFlag,
		Observaciones:  datos.Observaciones,
	}

	saved, insertErr := h.rrvActaRepo.Create(ctx, acta)
	if insertErr != nil {
		log.Printf("[SMS-INBOUND] error guardando acta: %v", insertErr)
		c.JSON(http.StatusOK, gin.H{"ok": false, "reason": "db_error"})
		return
	}

	ocrGuardado := false
	if estado == "PROCESADA" && !duplicadoFlag {
		existsOCR, existsErr := h.actaRepo.ExistsByCodigoMesa(ctx, datos.Mesa)
		if existsErr != nil {
			log.Printf("[SMS-INBOUND] error verificando acta OCR mesa=%s: %v", datos.Mesa, existsErr)
		} else if existsOCR {
			log.Printf("[SMS-INBOUND] OCR omitido: ya existe codigoMesa=%s", datos.Mesa)
		} else {
			mesaNumero, _ := strconv.Atoi(datos.Mesa)
			ocrActa := &models.Acta{
				P1:              datos.P1,
				P2:              datos.P2,
				P3:              datos.P3,
				P4:              datos.P4,
				VotosNulos:      datos.Nulos,
				VotosBlanco:     datos.Blancos,
				VotosValidos:    datos.P1 + datos.P2 + datos.P3 + datos.P4,
				CodigoMesa:      datos.Mesa,
				Mesa:            mesaNumero,
				PapeletasAnfora: datos.Total,
				TipoCliente:     "SMS",
			}
			if _, createErr := h.actaRepo.Create(ctx, ocrActa); createErr != nil {
				log.Printf("[SMS-INBOUND] error guardando copia OCR mesa=%s: %v", datos.Mesa, createErr)
			} else {
				ocrGuardado = true
				log.Printf("[SMS-INBOUND] copia OCR guardada codigoMesa=%s", datos.Mesa)
			}
		}
	}

	errMsg := ""
	if duplicadoFlag {
		errMsg = fmt.Sprintf("posible duplicado para mesa %s", datos.Mesa)
	}
	h.registrarEventoSMS(ctx, actaID, "SMS_ACTA_RECIBIDA", bson.M{
		"remitente":       normalizado,
		"mesa":            datos.Mesa,
		"estado":          estado,
		"total_declarado": datos.Total,
		"total_calculado": totalCalculado,
		"duplicado":       duplicadoFlag,
		"ocr_guardado":    ocrGuardado,
	}, errMsg)

	log.Printf("[SMS-INBOUND] acta guardada acta_id=%s mesa=%s estado=%s dup=%v ocr=%v", actaID, datos.Mesa, estado, duplicadoFlag, ocrGuardado)

	c.JSON(http.StatusOK, gin.H{
		"ok":              true,
		"acta_id":         saved.ActaID,
		"estado":          estado,
		"mesa":            datos.Mesa,
		"total_declarado": datos.Total,
		"total_calculado": totalCalculado,
		"duplicado":       duplicadoFlag,
		"ocr_guardado":    ocrGuardado,
	})
}

// ListNumeros lista los números autorizados.
func (h *SMSInboundHandler) ListNumeros(c *gin.Context) {
	numeros, err := h.smsRepo.GetAuthorizedNumbers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"numeros": numeros})
}

// AddNumero agrega un número a la lista de autorizados.
func (h *SMSInboundHandler) AddNumero(c *gin.Context) {
	var body struct {
		Numero string `json:"numero"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Numero == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "campo 'numero' requerido"})
		return
	}
	normalizado := normalizarNumeroSMS(body.Numero)
	if normalizado == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "número inválido"})
		return
	}
	if err := h.smsRepo.AddAuthorizedNumber(c.Request.Context(), normalizado); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "numero": normalizado})
}

// RemoveNumero elimina un número de la lista de autorizados.
func (h *SMSInboundHandler) RemoveNumero(c *gin.Context) {
	numero := c.Param("numero")
	normalizado := normalizarNumeroSMS(numero)
	if normalizado == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "número inválido"})
		return
	}
	if err := h.smsRepo.RemoveAuthorizedNumber(c.Request.Context(), normalizado); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "numero": normalizado})
}

// GetHistorial obtiene los últimos mensajes SMS recibidos.
func (h *SMSInboundHandler) GetHistorial(c *gin.Context) {
	actas, err := h.rrvActaRepo.GetRecentSMS(c.Request.Context(), 100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"historial": actas})
}

// ─── SMS helpers ──────────────────────────────────────────────────────────────

type datosActaSMS struct {
	Mesa                  string
	P1, P2, P3, P4        int
	Blancos, Nulos, Total int
	Observaciones         string
}

var camposObligatoriosSMS = []string{"MESA", "P1", "P2", "P3", "P4", "BLANCOS", "NULOS", "TOTAL"}

func parseSMSRRV(body string) (*datosActaSMS, error) {
	body = strings.TrimSpace(body)
	body = strings.ReplaceAll(body, "@", "|")

	parts := strings.Split(body, "|")
	if len(parts) == 0 || strings.TrimSpace(parts[0]) != "RRV" {
		return nil, fmt.Errorf("el SMS no comienza con RRV")
	}

	campos := make(map[string]string, len(parts))
	for _, part := range parts[1:] {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		campos[strings.TrimSpace(strings.ToUpper(kv[0]))] = strings.TrimSpace(kv[1])
	}

	for _, campo := range camposObligatoriosSMS {
		if _, ok := campos[campo]; !ok {
			return nil, fmt.Errorf("campo obligatorio faltante: %s", campo)
		}
	}

	parseInt := func(s, campo string) (int, error) {
		var v int
		if _, err := fmt.Sscanf(s, "%d", &v); err != nil {
			return 0, fmt.Errorf("%s: valor no entero: %q", campo, s)
		}
		if v < 0 {
			return 0, fmt.Errorf("%s: valor negativo: %d", campo, v)
		}
		return v, nil
	}

	p1, err := parseInt(campos["P1"], "P1")
	if err != nil {
		return nil, err
	}
	p2, err := parseInt(campos["P2"], "P2")
	if err != nil {
		return nil, err
	}
	p3, err := parseInt(campos["P3"], "P3")
	if err != nil {
		return nil, err
	}
	p4, err := parseInt(campos["P4"], "P4")
	if err != nil {
		return nil, err
	}
	blancos, err := parseInt(campos["BLANCOS"], "BLANCOS")
	if err != nil {
		return nil, err
	}
	nulos, err := parseInt(campos["NULOS"], "NULOS")
	if err != nil {
		return nil, err
	}
	total, err := parseInt(campos["TOTAL"], "TOTAL")
	if err != nil {
		return nil, err
	}

	return &datosActaSMS{
		Mesa: strings.TrimSpace(campos["MESA"]),
		P1:   p1, P2: p2, P3: p3, P4: p4,
		Blancos: blancos, Nulos: nulos, Total: total,
		Observaciones: strings.TrimSpace(campos["OBS"]),
	}, nil
}

func normalizarNumeroSMS(numero string) string {
	numero = strings.TrimSpace(numero)
	if strings.HasPrefix(numero, "+591") {
		return numero[4:]
	}
	if strings.HasPrefix(numero, "591") && len(numero) > 3 {
		return numero[3:]
	}
	return numero
}

func extraerCampoSMS(payload map[string]interface{}, campos []string) string {
	for _, campo := range campos {
		if v, ok := payload[campo]; ok {
			if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
				return strings.TrimSpace(s)
			}
		}
	}
	return ""
}
