package sms

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var camposRemitente = []string{"from", "sender", "phone", "number", "source", "From"}
var camposCuerpo = []string{"body", "message", "text", "content", "sms", "Body"}

// Handler gestiona los endpoints HTTP del módulo SMS.
type Handler struct {
	store      *Store
	forwardURL string
}

// NewHandler construye un Handler para el módulo SMS.
// forwardURL es la URL del SRV-ocr2 donde se reenvían los SMS recibidos (puede ser "").
func NewHandler(store *Store, forwardURL string) *Handler {
	return &Handler{store: store, forwardURL: forwardURL}
}

// InboundSMS recibe el webhook de SMS Forwarder y procesa el mensaje.
// Siempre responde HTTP 200 para evitar reintentos de la app ante cualquier error.
func (h *Handler) InboundSMS(c *gin.Context) {
	ctx := c.Request.Context()
	timestamp := time.Now()

	// Leer cuerpo bruto para intentar JSON y después form-urlencoded
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

	remitente := ExtraerCampo(rawPayload, camposRemitente)
	smsBody := ExtraerCampo(rawPayload, camposCuerpo)

	// Manejo del formato de la app Forward SMS donde todo está en "text":
	// "Desde : 65707079\nRRV@MESA=..."
	if remitente == "" && smsBody != "" {
		importRe := false // Only to remember we could use regexp, but strings is faster
		_ = importRe
		lines := strings.Split(smsBody, "\n")
		if len(lines) > 0 {
			firstLine := strings.TrimSpace(lines[0])
			if strings.HasPrefix(strings.ToLower(firstLine), "desde") {
				// Extraer el número después de los dos puntos
				parts := strings.SplitN(firstLine, ":", 2)
				if len(parts) == 2 {
					remitente = strings.TrimSpace(parts[1])
					// Reconstruir smsBody sin la primera línea
					if len(lines) > 1 {
						smsBody = strings.Join(lines[1:], "\n")
					} else {
						smsBody = ""
					}
				}
			}
		}
	}

	normalizado := NormalizarNumero(remitente)

	log.Printf("[SMS-INBOUND] ts=%s remitente=%q norm=%q cuerpo=%q",
		timestamp.Format(time.RFC3339), remitente, normalizado, smsBody)

	// ── Verificar autorización ────────────────────────────────────────────────
	autorizado, err := h.store.IsAuthorized(ctx, normalizado)
	if err != nil {
		log.Printf("[SMS-INBOUND] error verificando autorización: %v", err)
	}
	if !autorizado {
		log.Printf("[SMS-INBOUND] ADVERTENCIA: remitente no autorizado remitente=%q norm=%q",
			remitente, normalizado)
		h.store.InsertEvent(ctx, "", "SMS_REMITENTE_NO_AUTORIZADO",
			gin.H{"remitente": remitente, "normalizado": normalizado}, "remitente no autorizado")
		c.JSON(http.StatusOK, gin.H{"ok": false, "reason": "sender_not_authorized"})
		return
	}

	// ── Parsear cuerpo del SMS ────────────────────────────────────────────────
	datos, parseErr := ParseSMS(smsBody)
	if parseErr != nil {
		actaID := fmt.Sprintf("SMS-RECHAZADA-%d", timestamp.UnixMilli())
		acta := ActaSMS{
			ActaID:         actaID,
			Fuente:         "SMS",
			Remitente:      normalizado,
			RawPayload:     rawPayload,
			SMSBody:        smsBody,
			Estado:         EstadoRechazada,
			FechaRecepcion: timestamp,
			FechaProcesado: time.Now(),
		}
		if insertErr := h.store.InsertActa(ctx, acta); insertErr != nil {
			log.Printf("[SMS-INBOUND] error guardando acta rechazada: %v", insertErr)
		}
		h.store.InsertEvent(ctx, actaID, "SMS_ACTA_RECHAZADA",
			gin.H{"remitente": normalizado, "sms_body": smsBody}, parseErr.Error())
		log.Printf("[SMS-INBOUND] acta rechazada acta_id=%s razon=%v", actaID, parseErr)
		c.JSON(http.StatusOK, gin.H{
			"ok":      true,
			"estado":  EstadoRechazada,
			"acta_id": actaID,
			"razon":   parseErr.Error(),
		})
		return
	}

	// ── Calcular total y determinar estado ────────────────────────────────────
	totalCalculado := datos.P1 + datos.P2 + datos.P3 + datos.P4 + datos.Blancos + datos.Nulos
	estado := EstadoValidada
	
	if totalCalculado != datos.Total {
		estado = EstadoPendienteRevision
		// Registrar explícitamente la inconsistencia aritmética
		h.store.InsertEvent(ctx, "", "SMS_INCONSISTENCIA_ARITMETICA",
			gin.H{"mesa": datos.Mesa, "total_declarado": datos.Total, "total_calculado": totalCalculado}, 
			"Suma de votos no coincide con total declarado")
	}

	// ── Verificar duplicado e idempotencia ────────────────────────────────────
	existing, err := h.store.GetExistingActa(ctx, datos.Mesa)
	if err != nil {
		log.Printf("[SMS-INBOUND] error buscando duplicado mesa=%s: %v", datos.Mesa, err)
	}

	duplicadoFlag := false
	if existing != nil {
		// Validar si los datos son idénticos (idempotencia)
		if existing.TotalDeclarado == datos.Total &&
			existing.VotosBlancos == datos.Blancos &&
			existing.VotosNulos == datos.Nulos {
			// Asumiremos que si los totales, blancos y nulos coinciden, es el mismo SMS (se puede extender a comparar candidatos)
			// Para simplificar, comparamos totales.
			esIdentico := true
			for _, cand := range existing.Candidatos {
				switch cand.CandidatoID {
				case "P1": if cand.Votos != datos.P1 { esIdentico = false }
				case "P2": if cand.Votos != datos.P2 { esIdentico = false }
				case "P3": if cand.Votos != datos.P3 { esIdentico = false }
				case "P4": if cand.Votos != datos.P4 { esIdentico = false }
				}
			}
			
			if esIdentico {
				log.Printf("[SMS-INBOUND] Idempotencia: SMS idéntico ignorado para mesa %s", datos.Mesa)
				// Devolver 200 OK con el acta_id original, sin insertar nada.
				c.JSON(http.StatusOK, gin.H{
					"ok":      true,
					"estado":  existing.Estado,
					"acta_id": existing.ActaID,
					"nota":    "idempotent_retry",
				})
				return
			}
		}

		// Si no es idéntico, son datos contradictorios
		duplicadoFlag = true
		estado = EstadoPendienteRevision
		log.Printf("[SMS-INBOUND] ADVERTENCIA: datos contradictorios mesa=%s", datos.Mesa)
		h.store.InsertEvent(ctx, "", "SMS_DATOS_CONTRADICTORIOS",
			gin.H{"mesa": datos.Mesa, "nuevo_total": datos.Total, "viejo_total": existing.TotalDeclarado}, 
			"Acta recibida con anterioridad y no coinciden los datos")
	}

	actaID := fmt.Sprintf("SMS-%s-%d", datos.Mesa, timestamp.UnixMilli())

	acta := ActaSMS{
		ActaID: actaID,
		Mesa:   datos.Mesa,
		Candidatos: []CandidatoSMS{
			{CandidatoID: "P1", Nombre: "Partido P1", Votos: datos.P1},
			{CandidatoID: "P2", Nombre: "Partido P2", Votos: datos.P2},
			{CandidatoID: "P3", Nombre: "Partido P3", Votos: datos.P3},
			{CandidatoID: "P4", Nombre: "Partido P4", Votos: datos.P4},
		},
		VotosNulos:     datos.Nulos,
		VotosBlancos:   datos.Blancos,
		TotalVotos:     datos.Total,
		Estado:         estado,
		Fuente:         "SMS",
		Remitente:      normalizado,
		RawPayload:     rawPayload,
		SMSBody:        smsBody,
		DatosParsed:    datos,
		TotalDeclarado: datos.Total,
		TotalCalculado: totalCalculado,
		Duplicado:      duplicadoFlag,
		Observaciones:  datos.Observaciones,
		FechaRecepcion: timestamp,
		FechaProcesado: time.Now(),
	}

	if insertErr := h.store.InsertActa(ctx, acta); insertErr != nil {
		log.Printf("[SMS-INBOUND] error guardando acta: %v", insertErr)
		c.JSON(http.StatusOK, gin.H{"ok": false, "reason": "db_error"})
		return
	}

	errMsg := ""
	if duplicadoFlag {
		errMsg = fmt.Sprintf("posible duplicado para mesa %s", datos.Mesa)
	}
	h.store.InsertEvent(ctx, actaID, "SMS_ACTA_RECIBIDA", gin.H{
		"remitente":       normalizado,
		"mesa":            datos.Mesa,
		"estado":          estado,
		"total_declarado": datos.Total,
		"total_calculado": totalCalculado,
		"duplicado":       duplicadoFlag,
	}, errMsg)

	log.Printf("[SMS-INBOUND] acta guardada acta_id=%s mesa=%s estado=%s total_dec=%d total_calc=%d dup=%v",
		actaID, datos.Mesa, estado, datos.Total, totalCalculado, duplicadoFlag)

	// Reenviar al SRV-ocr2 en background para que también lo procese
	if h.forwardURL != "" {
		go h.reenviarAOCR2(rawBody)
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":              true,
		"acta_id":         actaID,
		"estado":          estado,
		"mesa":            datos.Mesa,
		"total_declarado": datos.Total,
		"total_calculado": totalCalculado,
		"duplicado":       duplicadoFlag,
	})
}

// ListNumeros lista los números autorizados actuales.
func (h *Handler) ListNumeros(c *gin.Context) {
	ctx := c.Request.Context()
	numeros, err := h.store.GetAuthorizedNumbers(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"numeros": numeros})
}

// AddNumero agrega un número a la lista de autorizados.
func (h *Handler) AddNumero(c *gin.Context) {
	var body struct {
		Numero string `json:"numero"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Numero == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "campo 'numero' requerido"})
		return
	}
	normalizado := NormalizarNumero(body.Numero)
	if normalizado == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "número inválido"})
		return
	}
	ctx := c.Request.Context()
	if err := h.store.AddAuthorizedNumber(ctx, normalizado); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Printf("[SMS-CONFIG] número autorizado agregado: %s", normalizado)
	c.JSON(http.StatusOK, gin.H{"ok": true, "numero": normalizado})
}

// RemoveNumero elimina un número de la lista de autorizados.
func (h *Handler) RemoveNumero(c *gin.Context) {
	numero := c.Param("numero")
	normalizado := NormalizarNumero(numero)
	if normalizado == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "número inválido"})
		return
	}
	ctx := c.Request.Context()
	if err := h.store.RemoveAuthorizedNumber(ctx, normalizado); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Printf("[SMS-CONFIG] número eliminado de autorizados: %s", normalizado)
	c.JSON(http.StatusOK, gin.H{"ok": true, "numero": normalizado})
}

// GetHistorial obtiene el historial reciente de mensajes SMS.
func (h *Handler) GetHistorial(c *gin.Context) {
	ctx := c.Request.Context()
	actas, err := h.store.GetRecentSMS(ctx, 50)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"historial": actas})
}

// reenviarAOCR2 reenvía el payload raw al SRV-ocr2 para que lo procese independientemente.
func (h *Handler) reenviarAOCR2(rawBody []byte) {
	resp, err := http.Post(h.forwardURL, "application/json", bytes.NewReader(rawBody))
	if err != nil {
		log.Printf("[SMS-FORWARD] error reenviando a OCR2: %v", err)
		return
	}
	defer resp.Body.Close()
	log.Printf("[SMS-FORWARD] OCR2 respondió status=%d", resp.StatusCode)
}
