package comparacion

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

// Handler gestiona los endpoints HTTP del módulo de comparación.
// Este módulo es de solo lectura (CQRS/Query): nunca escribe en ninguna DB.
type Handler struct {
	svc  *Service
	comp *Comparer
}

func newHandler(db *gorm.DB, mongoClient *mongo.Client, dbName string) *Handler {
	return &Handler{
		svc:  &Service{DB: db, MongoClient: mongoClient, DBName: dbName},
		comp: &Comparer{},
	}
}

// ─── Filtros ─────────────────────────────────────────────────────────────────

// Filtros agrupa todos los parámetros de query aceptados por el endpoint principal.
type Filtros struct {
	Estado       string
	Departamento string
	Municipio    string
	Recinto      string
}

// parseFiltros extrae los filtros del request. Todos son case-insensitive en comparación.
func parseFiltros(c *gin.Context) Filtros {
	return Filtros{
		Estado:       strings.ToUpper(strings.TrimSpace(c.Query("estado"))),
		Departamento: strings.TrimSpace(c.Query("departamento")),
		Municipio:    strings.TrimSpace(c.Query("municipio")),
		Recinto:      strings.TrimSpace(c.Query("recinto")),
	}
}

// filtrarActas aplica los filtros al slice de actas ya comparadas.
// La comparación de estado es exacta (siempre viene en mayúsculas desde parseFiltros).
// La comparación de campos de texto es case-insensitive.
func filtrarActas(actas []ActaComparada, f Filtros) []ActaComparada {
	if f.Estado == "" && f.Departamento == "" && f.Municipio == "" && f.Recinto == "" {
		return actas
	}
	result := make([]ActaComparada, 0, len(actas))
	for _, a := range actas {
		if f.Estado != "" && string(a.EstadoComparacion) != f.Estado {
			continue
		}
		if f.Departamento != "" && !strings.EqualFold(a.Departamento, f.Departamento) {
			continue
		}
		if f.Municipio != "" && !strings.EqualFold(a.Municipio, f.Municipio) {
			continue
		}
		if f.Recinto != "" && !strings.EqualFold(a.Recinto, f.Recinto) {
			continue
		}
		result = append(result, a)
	}
	return result
}

// parsePaginacion extrae page/pagina y limit/por_pagina con compatibilidad hacia atrás.
// Orden de precedencia: pagina > page, por_pagina > limit.
func parsePaginacion(c *gin.Context) (pagina, porPagina int) {
	pagina = 1
	porPagina = 50

	if p, err := strconv.Atoi(c.Query("pagina")); err == nil && p > 0 {
		pagina = p
	} else if p, err := strconv.Atoi(c.Query("page")); err == nil && p > 0 {
		pagina = p
	}

	if pp, err := strconv.Atoi(c.Query("por_pagina")); err == nil && pp > 0 {
		if pp > 200 {
			pp = 200
		}
		porPagina = pp
	} else if pp, err := strconv.Atoi(c.Query("limit")); err == nil && pp > 0 {
		if pp > 200 {
			pp = 200
		}
		porPagina = pp
	}
	return
}

// ─── GET /api/v1/comparacion ──────────────────────────────────────────────────
//
// Query params:
//
//	estado        CONSISTENTE | INCONSISTENTE | SOLO_RRV | SOLO_OFICIAL  (case-insensitive)
//	departamento  filtra por departamento (case-insensitive)
//	municipio     filtra por municipio (case-insensitive)
//	recinto       filtra por recinto (case-insensitive)
//	pagina / page número de página (default 1)
//	por_pagina / limit registros por página (default 50, max 200)
func (h *Handler) Comparar(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	rrv, oficiales, err := h.fetchAmbas(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resumen := h.comp.Comparar(oficiales, rrv)

	filtradas := filtrarActas(resumen.Actas, parseFiltros(c))

	pagina, porPagina := parsePaginacion(c)
	total := len(filtradas)
	inicio, fin := paginaSlice(total, pagina, porPagina)

	resumenSinActas := resumen
	resumenSinActas.Actas = nil

	c.JSON(http.StatusOK, RespuestaComparacion{
		Resumen:   resumenSinActas,
		Actas:     filtradas[inicio:fin],
		Total:     total,
		Pagina:    pagina,
		PorPagina: porPagina,
	})
}

// ─── GET /api/v1/comparacion/resumen ─────────────────────────────────────────

func (h *Handler) Resumen(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	rrv, oficiales, err := h.fetchAmbas(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resumen := h.comp.Comparar(oficiales, rrv)
	resumen.Actas = nil
	c.JSON(http.StatusOK, resumen)
}

// ─── GET /api/v1/comparacion/inconsistencias ──────────────────────────────────

func (h *Handler) Inconsistencias(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	rrv, oficiales, err := h.fetchAmbas(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resumen := h.comp.Comparar(oficiales, rrv)

	var problematicas []ActaComparada
	for _, a := range resumen.Actas {
		if a.EstadoComparacion != EstadoConsistente {
			problematicas = append(problematicas, a)
		}
	}

	pagina, porPagina := parsePaginacion(c)
	total := len(problematicas)
	inicio, fin := paginaSlice(total, pagina, porPagina)

	resumenSinActas := resumen
	resumenSinActas.Actas = nil

	c.JSON(http.StatusOK, RespuestaComparacion{
		Resumen:   resumenSinActas,
		Actas:     problematicas[inicio:fin],
		Total:     total,
		Pagina:    pagina,
		PorPagina: porPagina,
	})
}

// ─── GET /api/v1/comparacion/:acta_id ────────────────────────────────────────

func (h *Handler) GetActaComparada(c *gin.Context) {
	actaID := c.Param("acta_id")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	rrv, oficiales, err := h.fetchAmbas(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resumen := h.comp.Comparar(oficiales, rrv)

	for _, a := range resumen.Actas {
		if a.ActaID == actaID {
			c.JSON(http.StatusOK, a)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "acta no encontrada en ninguna fuente: " + actaID})
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func (h *Handler) fetchAmbas(ctx context.Context) (rrv, oficiales []ActaFuente, err error) {
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

// paginaSlice calcula inicio y fin seguros para un slice de longitud total.
func paginaSlice(total, pagina, porPagina int) (inicio, fin int) {
	inicio = (pagina - 1) * porPagina
	fin = inicio + porPagina
	if inicio >= total {
		inicio = total
	}
	if fin > total {
		fin = total
	}
	return
}
