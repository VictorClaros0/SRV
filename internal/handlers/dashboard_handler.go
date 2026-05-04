package handlers

import (
	"bufio"
	"net/http"
	"os"
	"strings"

	"srrv/internal/repository"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

const logPath = "/app/data/logs/unreadable_records.log"

type LogEntry struct {
	CodigoMesa      string   `json:"codigoMesa"`
	Archivo         string   `json:"archivo"`
	CamposIlegibles []string `json:"camposIlegibles"`
}

type DashboardHandler struct {
	repo *repository.DashboardRepository
}

func NewDashboardHandler(repo *repository.DashboardRepository) *DashboardHandler {
	return &DashboardHandler{repo: repo}
}

func (h *DashboardHandler) GetKPIs(c *gin.Context) {
	result, err := h.repo.GetKPIs(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *DashboardHandler) GetVotosCandidato(c *gin.Context) {
	result, err := h.repo.GetVotosCandidato(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"candidatos": result})
}

func (h *DashboardHandler) GetParticipacion(c *gin.Context) {
	result, err := h.repo.GetParticipacion(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *DashboardHandler) GetGeografico(c *gin.Context) {
	nivel := c.DefaultQuery("nivel", "departamento")
	departamento := c.Query("departamento")
	provincia := c.Query("provincia")
	municipio := c.Query("municipio")

	result, err := h.repo.GetGeografico(c.Request.Context(), nivel, departamento, provincia, municipio)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"nivel": nivel, "datos": result})
}

func (h *DashboardHandler) GetHeatmap(c *gin.Context) {
	metric := c.DefaultQuery("metric", "participacion")
	nivel := c.DefaultQuery("nivel", "departamento")

	result, err := h.repo.GetHeatmap(c.Request.Context(), metric, nivel)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"nivel": nivel, "metric": metric, "datos": result})
}

func (h *DashboardHandler) GetTransparencia(c *gin.Context) {
	result, err := h.repo.GetTransparencia(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"actas": result})
}

func (h *DashboardHandler) GetTrazabilidad(c *gin.Context) {
	codigoActa := c.Param("codigoActa")

	result, err := h.repo.GetTrazabilidad(c.Request.Context(), codigoActa)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Acta no encontrada"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *DashboardHandler) GetTecnico(c *gin.Context) {
	result, err := h.repo.GetTecnico(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *DashboardHandler) GetAnomalias(c *gin.Context) {
	result, err := h.repo.GetAnomalias(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *DashboardHandler) GetDepartamentos(c *gin.Context) {
	result, err := h.repo.GetDepartamentos(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"departamentos": result})
}

func (h *DashboardHandler) GetProvincias(c *gin.Context) {
	departamento := c.Query("departamento")
	result, err := h.repo.GetProvincias(c.Request.Context(), departamento)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"provincias": result})
}

func (h *DashboardHandler) GetMunicipios(c *gin.Context) {
	provincia := c.Query("provincia")
	result, err := h.repo.GetMunicipios(c.Request.Context(), provincia)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"municipios": result})
}

func (h *DashboardHandler) GetRecintos(c *gin.Context) {
	municipio := c.Query("municipio")
	result, err := h.repo.GetRecintosByMunicipio(c.Request.Context(), municipio)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"recintos": result})
}

func (h *DashboardHandler) GetLogInconsistencias(c *gin.Context) {
	f, err := os.Open(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusOK, gin.H{"total": 0, "registros": []LogEntry{}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer f.Close()

	var registros []LogEntry
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		entry := parsearLineaLog(line)
		if entry != nil {
			registros = append(registros, *entry)
		}
	}
	if registros == nil {
		registros = []LogEntry{}
	}
	c.JSON(http.StatusOK, gin.H{"total": len(registros), "registros": registros})
}

func parsearLineaLog(line string) *LogEntry {
	// Formato: Mesa: 10001 | Archivo: acta_10001.pdf | Ilegibles: p1, p2
	partes := strings.Split(line, " | ")
	if len(partes) < 3 {
		return nil
	}
	entry := &LogEntry{}
	for _, p := range partes {
		p = strings.TrimSpace(p)
		switch {
		case strings.HasPrefix(p, "Mesa: "):
			entry.CodigoMesa = strings.TrimPrefix(p, "Mesa: ")
		case strings.HasPrefix(p, "Archivo: "):
			entry.Archivo = strings.TrimPrefix(p, "Archivo: ")
		case strings.HasPrefix(p, "Ilegibles: "):
			raw := strings.TrimPrefix(p, "Ilegibles: ")
			campos := strings.Split(raw, ", ")
			for _, c := range campos {
				c = strings.TrimSpace(c)
				if c != "" {
					entry.CamposIlegibles = append(entry.CamposIlegibles, c)
				}
			}
		}
	}
	if entry.CodigoMesa == "" {
		return nil
	}
	return entry
}

func (h *DashboardHandler) GetMesas(c *gin.Context) {
	recintoID := c.Query("recinto_id")
	if recintoID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "recinto_id es requerido"})
		return
	}
	result, err := h.repo.GetMesasByRecinto(c.Request.Context(), recintoID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"mesas": result})
}
