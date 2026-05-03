package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type FiltrosHandler struct {
	DB *gorm.DB
}

type filtroItem struct {
	ID     interface{} `gorm:"column:id" json:"id"`
	Nombre string      `gorm:"column:nombre" json:"nombre"`
}

type filtroNombre struct {
	ID     string `gorm:"column:id" json:"id"`
	Nombre string `gorm:"column:nombre" json:"nombre"`
}

// FiltroDepartamentos devuelve todos los departamentos.
func (h *FiltrosHandler) FiltroDepartamentos(c *gin.Context) {
	var rows []filtroNombre
	h.DB.Raw(`
		SELECT DISTINCT departamento AS id, departamento AS nombre
		FROM distribucion_territorial
		ORDER BY departamento
	`).Scan(&rows)
	c.JSON(http.StatusOK, rows)
}

// FiltroMunicipios devuelve municipios, opcionalmente filtrados por ?departamento=X.
func (h *FiltrosHandler) FiltroMunicipios(c *gin.Context) {
	dept := strings.TrimSpace(c.Query("departamento"))
	q := "SELECT DISTINCT municipio AS id, municipio AS nombre FROM distribucion_territorial WHERE 1=1"
	var args []interface{}
	if dept != "" {
		q += " AND departamento = ?"
		args = append(args, dept)
	}
	q += " ORDER BY municipio"
	var rows []filtroNombre
	h.DB.Raw(q, args...).Scan(&rows)
	c.JSON(http.StatusOK, rows)
}

// FiltroProvincias devuelve provincias, opcionalmente filtradas por ?departamento=X&municipio=Y.
func (h *FiltrosHandler) FiltroProvincias(c *gin.Context) {
	dept := strings.TrimSpace(c.Query("departamento"))
	mun  := strings.TrimSpace(c.Query("municipio"))
	q := "SELECT DISTINCT provincia AS id, provincia AS nombre FROM distribucion_territorial WHERE 1=1"
	var args []interface{}
	if dept != "" {
		q += " AND departamento = ?"
		args = append(args, dept)
	}
	if mun != "" {
		q += " AND municipio = ?"
		args = append(args, mun)
	}
	q += " ORDER BY provincia"
	var rows []filtroNombre
	h.DB.Raw(q, args...).Scan(&rows)
	c.JSON(http.StatusOK, rows)
}

// FiltroRecintos devuelve recintos, opcionalmente filtrados por ?departamento=X&municipio=Y&provincia=Z.
func (h *FiltrosHandler) FiltroRecintos(c *gin.Context) {
	dept := strings.TrimSpace(c.Query("departamento"))
	mun  := strings.TrimSpace(c.Query("municipio"))
	prov := strings.TrimSpace(c.Query("provincia"))
	q := `
		SELECT re.recinto_id AS id, re.recinto AS nombre
		FROM recinto_electoral re
		JOIN distribucion_territorial dt ON re.id_distribucion_territorial = dt.id
		WHERE 1=1`
	var args []interface{}
	if dept != "" {
		q += " AND dt.departamento = ?"
		args = append(args, dept)
	}
	if mun != "" {
		q += " AND dt.municipio = ?"
		args = append(args, mun)
	}
	if prov != "" {
		q += " AND dt.provincia = ?"
		args = append(args, prov)
	}
	q += " ORDER BY re.recinto"
	var rows []filtroItem
	h.DB.Raw(q, args...).Scan(&rows)
	c.JSON(http.StatusOK, rows)
}

// FiltroMesas devuelve nros de mesa desde actas, opcionalmente filtradas por ?recinto=X.
func (h *FiltrosHandler) FiltroMesas(c *gin.Context) {
	recinto := strings.TrimSpace(c.Query("recinto"))
	q := `
		SELECT DISTINCT a.nro_mesa AS id, CAST(a.nro_mesa AS TEXT) AS nombre
		FROM acta a
		WHERE a.fecha_eliminado IS NULL`
	var args []interface{}
	if recinto != "" {
		q += " AND LEFT(CAST(a.codigo_recinto AS TEXT), 5) = LEFT(CAST(? AS TEXT), 5)"
		args = append(args, recinto)
	}
	q += " ORDER BY a.nro_mesa"
	var rows []filtroItem
	h.DB.Raw(q, args...).Scan(&rows)
	c.JSON(http.StatusOK, rows)
}
