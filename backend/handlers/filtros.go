package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type FiltrosHandler struct {
	DB *gorm.DB
}

type filtroItem struct {
	ID     uint   `gorm:"column:id" json:"id"`
	Nombre string `gorm:"column:nombre" json:"nombre"`
}

type filtroNombre struct {
	Nombre string `gorm:"column:nombre" json:"nombre"`
}

// FiltroRecintos devuelve id y nombre de todos los recintos electorales.
func (h *FiltrosHandler) FiltroRecintos(c *gin.Context) {
	var rows []filtroItem
	h.DB.Raw(`
		SELECT recinto_id AS id, recinto AS nombre
		FROM recinto_electoral
		ORDER BY recinto
	`).Scan(&rows)
	c.JSON(http.StatusOK, rows)
}

// FiltroMesas devuelve id y codigo de todas las mesas.
func (h *FiltrosHandler) FiltroMesas(c *gin.Context) {
	var rows []filtroItem
	h.DB.Raw(`
		SELECT id, codigo AS nombre
		FROM mesa
		ORDER BY codigo
	`).Scan(&rows)
	c.JSON(http.StatusOK, rows)
}

// FiltroProvincias devuelve los nombres distintos de provincias.
func (h *FiltrosHandler) FiltroProvincias(c *gin.Context) {
	var rows []filtroNombre
	h.DB.Raw(`
		SELECT DISTINCT provincia AS nombre
		FROM distribucion_territorial
		ORDER BY provincia
	`).Scan(&rows)
	c.JSON(http.StatusOK, rows)
}

// FiltroDepartamentos devuelve los nombres distintos de departamentos.
func (h *FiltrosHandler) FiltroDepartamentos(c *gin.Context) {
	var rows []filtroNombre
	h.DB.Raw(`
		SELECT DISTINCT departamento AS nombre
		FROM distribucion_territorial
		ORDER BY departamento
	`).Scan(&rows)
	c.JSON(http.StatusOK, rows)
}
