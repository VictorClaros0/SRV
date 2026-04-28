package services

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

// TwilioClient envía SMS vía la API REST de Twilio.
type TwilioClient struct {
	AccountSID         string
	AuthToken          string
	MessagingServiceSID string
	// VerifiedNumber es el número al que se enviarán confirmaciones en cuentas trial
	// (Twilio trial solo permite enviar a números verificados).
	VerifiedNumber string
}

// NewTwilioClient construye un cliente Twilio. Devuelve nil si las credenciales están vacías.
func NewTwilioClient(accountSID, authToken, messagingSID, verifiedNumber string) *TwilioClient {
	if accountSID == "" || authToken == "" || messagingSID == "" {
		log.Println("⚠️  Twilio no configurado: TWILIO_ACCOUNT_SID / TWILIO_AUTH_TOKEN / TWILIO_MESSAGING_SERVICE_SID ausentes")
		return nil
	}
	return &TwilioClient{
		AccountSID:          accountSID,
		AuthToken:           authToken,
		MessagingServiceSID: messagingSID,
		VerifiedNumber:      verifiedNumber,
	}
}

// SendSMS envía un SMS al número `to` con el texto `body`.
// En cuentas Twilio trial, `to` debe ser un número verificado.
func (t *TwilioClient) SendSMS(to, body string) error {
	apiURL := fmt.Sprintf(
		"https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json",
		t.AccountSID,
	)

	data := url.Values{}
	data.Set("To", to)
	data.Set("MessagingServiceSid", t.MessagingServiceSID)
	data.Set("Body", body)

	req, err := http.NewRequest(http.MethodPost, apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("twilio: error creando request: %w", err)
	}
	req.SetBasicAuth(t.AccountSID, t.AuthToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("twilio: error enviando SMS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("twilio: error HTTP %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// SendConfirmation envía un SMS de confirmación. En cuentas trial usa VerifiedNumber
// como destino ya que Twilio no permite enviar a números no verificados.
func (t *TwilioClient) SendConfirmation(toNumber, mensaje string) {
	dest := toNumber
	// Cuenta trial: solo se puede enviar a números verificados
	if t.VerifiedNumber != "" {
		dest = t.VerifiedNumber
	}
	if dest == "" {
		log.Printf("⚠️  Twilio: no hay número destino para confirmación")
		return
	}
	if err := t.SendSMS(dest, mensaje); err != nil {
		log.Printf("⚠️  Twilio: error enviando confirmación a %s: %v", dest, err)
	} else {
		log.Printf("✅ Twilio: confirmación enviada a %s", dest)
	}
}
