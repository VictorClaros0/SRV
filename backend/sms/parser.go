package sms

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	EstadoValidada          = "VALIDADA"
	EstadoPendienteRevision = "PENDIENTE_REVISION"
	EstadoRechazada         = "RECHAZADA"
)

var camposObligatorios = []string{"MESA", "P1", "P2", "P3", "P4", "BLANCOS", "NULOS", "TOTAL"}

// DatosActa contiene los datos parseados del cuerpo del SMS.
type DatosActa struct {
	Mesa         string `bson:"mesa"          json:"mesa"`
	P1           int    `bson:"p1"            json:"p1"`
	P2           int    `bson:"p2"            json:"p2"`
	P3           int    `bson:"p3"            json:"p3"`
	P4           int    `bson:"p4"            json:"p4"`
	Blancos      int    `bson:"blancos"       json:"blancos"`
	Nulos        int    `bson:"nulos"         json:"nulos"`
	Total        int    `bson:"total"         json:"total"`
	Observaciones string `bson:"observaciones" json:"observaciones,omitempty"`
}

// ParseSMS parsea el cuerpo del SMS con formato: RRV|MESA=...|P1=...|... o con @ en lugar de |
// Retorna error si el formato es inválido o faltan campos obligatorios.
func ParseSMS(body string) (*DatosActa, error) {
	body = strings.TrimSpace(body)
	// Soportar @ como separador en lugar de |
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
		k := strings.TrimSpace(strings.ToUpper(kv[0]))
		v := strings.TrimSpace(kv[1])
		campos[k] = v
	}

	for _, campo := range camposObligatorios {
		if _, ok := campos[campo]; !ok {
			return nil, fmt.Errorf("campo obligatorio faltante: %s", campo)
		}
	}

	p1, err := parseEnteroPositivo(campos["P1"], "P1")
	if err != nil {
		return nil, err
	}
	p2, err := parseEnteroPositivo(campos["P2"], "P2")
	if err != nil {
		return nil, err
	}
	p3, err := parseEnteroPositivo(campos["P3"], "P3")
	if err != nil {
		return nil, err
	}
	p4, err := parseEnteroPositivo(campos["P4"], "P4")
	if err != nil {
		return nil, err
	}
	blancos, err := parseEnteroPositivo(campos["BLANCOS"], "BLANCOS")
	if err != nil {
		return nil, err
	}
	nulos, err := parseEnteroPositivo(campos["NULOS"], "NULOS")
	if err != nil {
		return nil, err
	}
	total, err := parseEnteroPositivo(campos["TOTAL"], "TOTAL")
	if err != nil {
		return nil, err
	}

	return &DatosActa{
		Mesa:          strings.TrimSpace(campos["MESA"]),
		P1:            p1,
		P2:            p2,
		P3:            p3,
		P4:            p4,
		Blancos:       blancos,
		Nulos:         nulos,
		Total:         total,
		Observaciones: strings.TrimSpace(campos["OBS"]),
	}, nil
}

func parseEnteroPositivo(s, campo string) (int, error) {
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("%s: valor no entero: %q", campo, s)
	}
	if v < 0 {
		return 0, fmt.Errorf("%s: valor negativo: %d", campo, v)
	}
	return v, nil
}

// NormalizarNumero elimina el prefijo boliviano +591 o 591 para normalizar.
// +59165707079 → 65707079
// 59165707079  → 65707079
// 65707079     → 65707079
func NormalizarNumero(numero string) string {
	numero = strings.TrimSpace(numero)
	if strings.HasPrefix(numero, "+591") {
		return numero[4:]
	}
	if strings.HasPrefix(numero, "591") && len(numero) > 3 {
		return numero[3:]
	}
	return numero
}

// ExtraerCampo busca el primer campo no vacío del payload según lista de prioridad.
func ExtraerCampo(payload map[string]interface{}, campos []string) string {
	for _, campo := range campos {
		if v, ok := payload[campo]; ok {
			if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
				return strings.TrimSpace(s)
			}
		}
	}
	return ""
}
