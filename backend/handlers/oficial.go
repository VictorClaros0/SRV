package handlers

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/srvof/votos-backend/middleware"
	"github.com/srvof/votos-backend/models"
	"gorm.io/gorm"
)

type OficialHandler struct {
	DB *gorm.DB
}

// ─── DTOs ─────────────────────────────────────────────────────────────────────

type CandidatoOficial struct {
	CandidatoID string `json:"candidato_id"`
	Nombre      string `json:"nombre"`
	Votos       int    `json:"votos"`
}

type ActaOficialDTO struct {
	ActaID       string             `json:"acta_id"`
	Departamento string             `json:"departamento"`
	Provincia    string             `json:"provincia"`
	Municipio    string             `json:"municipio"`
	Recinto      string             `json:"recinto"`
	Mesa         string             `json:"mesa"`
	Candidatos   []CandidatoOficial `json:"candidatos"`
	VotosNulos   int                `json:"votos_nulos"`
	VotosBlancos int                `json:"votos_blancos"`
	TotalVotos   int                `json:"total_votos"`
	Estado       string             `json:"estado"`
	Fuente       string             `json:"fuente"`
}

type ErrorFila struct {
	Fila   int    `json:"fila"`
	ActaID string `json:"acta_id"`
	Error  string `json:"error"`
}

// ─── Fila CSV interna ─────────────────────────────────────────────────────────

type csvRowOficial struct {
	ActaID                string
	Departamento          string
	Provincia             string
	Municipio             string
	Recinto               string
	Mesa                  string
	Candidato1            int
	Candidato2            int
	Candidato3            int
	Candidato4            int
	VotosNulos            int
	VotosBlancos          int
	TotalVotos            int
	CiudadanosHabilitados int
	Observacion           string
}

func parseCSVRowOficial(record []string, lineNum int) (*csvRowOficial, error) {
	if len(record) < 14 {
		return nil, fmt.Errorf("se esperan 14 columnas, se encontraron %d", len(record))
	}
	row := &csvRowOficial{
		ActaID:       strings.TrimSpace(record[0]),
		Departamento: strings.TrimSpace(record[1]),
		Provincia:    strings.TrimSpace(record[2]),
		Municipio:    strings.TrimSpace(record[3]),
		Recinto:      strings.TrimSpace(record[4]),
		Mesa:         strings.TrimSpace(record[5]),
	}
	if len(record) >= 15 {
		row.Observacion = strings.TrimSpace(record[14])
	}
	labels := []string{
		"candidato_1", "candidato_2", "candidato_3", "candidato_4",
		"votos_nulos", "votos_blancos", "total_votos", "ciudadanos_habilitados",
	}
	nums := make([]int, 8)
	for i, label := range labels {
		v, err := strconv.Atoi(strings.TrimSpace(record[6+i]))
		if err != nil {
			return nil, fmt.Errorf("%s debe ser un número entero", label)
		}
		nums[i] = v
	}
	row.Candidato1 = nums[0]
	row.Candidato2 = nums[1]
	row.Candidato3 = nums[2]
	row.Candidato4 = nums[3]
	row.VotosNulos = nums[4]
	row.VotosBlancos = nums[5]
	row.TotalVotos = nums[6]
	row.CiudadanosHabilitados = nums[7]
	return row, nil
}

func validateRowOficial(row *csvRowOficial) error {
	// Reglas 1-5: campos obligatorios
	if row.ActaID == "" {
		return fmt.Errorf("acta_id es obligatorio")
	}
	if row.Departamento == "" {
		return fmt.Errorf("departamento es obligatorio")
	}
	if row.Municipio == "" {
		return fmt.Errorf("municipio es obligatorio")
	}
	if row.Recinto == "" {
		return fmt.Errorf("recinto es obligatorio")
	}
	if row.Mesa == "" {
		return fmt.Errorf("mesa es obligatoria")
	}
	// Regla 6: votos no negativos
	if row.Candidato1 < 0 || row.Candidato2 < 0 || row.Candidato3 < 0 || row.Candidato4 < 0 ||
		row.VotosNulos < 0 || row.VotosBlancos < 0 || row.TotalVotos < 0 || row.CiudadanosHabilitados < 0 {
		return fmt.Errorf("los votos no pueden ser negativos")
	}
	// Reglas 7 y 8: total_votos correcto y suma coincide
	suma := row.Candidato1 + row.Candidato2 + row.Candidato3 + row.Candidato4 + row.VotosNulos + row.VotosBlancos
	if suma != row.TotalVotos {
		return fmt.Errorf("suma de votos (%d) no coincide con total_votos (%d)", suma, row.TotalVotos)
	}
	// Regla 9: total <= ciudadanos habilitados
	if row.TotalVotos > row.CiudadanosHabilitados {
		return fmt.Errorf("total_votos (%d) supera ciudadanos_habilitados (%d)", row.TotalVotos, row.CiudadanosHabilitados)
	}
	return nil
}

func (h *OficialHandler) registrarEvento(auditoriaID *uint, tipo, actaID string, payload interface{}) {
	p := ""
	if payload != nil {
		b, _ := json.Marshal(payload)
		p = string(b)
	}
	ev := models.OficialEvento{
		AuditoriaID: auditoriaID,
		Tipo:        tipo,
		ActaID:      actaID,
		Payload:     p,
		Timestamp:   time.Now(),
	}
	h.DB.Create(&ev)
}

func (h *OficialHandler) rechazar(auditoriaID uint, lineNum int, actaID, msg string, errores *[]ErrorFila) {
	*errores = append(*errores, ErrorFila{Fila: lineNum, ActaID: actaID, Error: msg})
	h.DB.Create(&models.OficialError{AuditoriaID: auditoriaID, Fila: lineNum, ActaID: actaID, Error: msg})
	h.registrarEvento(&auditoriaID, "ACTA_OFICIAL_RECHAZADA", actaID, gin.H{"fila": lineNum, "error": msg})
	h.registrarEvento(&auditoriaID, "ERROR_VALIDACION_REGISTRADO", actaID, gin.H{"fila": lineNum, "error": msg})
}

// Upload POST /api/oficial/csv/upload
func (h *OficialHandler) Upload(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "campo 'file' con el CSV es requerido"})
		return
	}
	defer file.Close()

	usuario := "sistema"
	if uid, ok := middleware.UserIDFromContext(c); ok {
		var u models.Usuario
		if h.DB.First(&u, uid).Error == nil {
			usuario = u.NombreUsuario
		}
	}

	auditoria := models.OficialAuditoria{
		NombreArchivo: header.Filename,
		Usuario:       usuario,
		Estado:        "PROCESANDO",
	}
	if err := h.DB.Create(&auditoria).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo crear auditoría"})
		return
	}

	h.registrarEvento(&auditoria.ID, "CSV_RECIBIDO", "", gin.H{
		"archivo": header.Filename,
		"usuario": usuario,
	})

	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true

	// Saltar cabecera
	if _, err := reader.Read(); err != nil {
		h.DB.Model(&auditoria).Update("estado", "FALLIDO")
		c.JSON(http.StatusBadRequest, gin.H{"error": "CSV vacío o sin cabecera"})
		return
	}

	var errores []ErrorFila
	filasProcesadas, filasValidas, filasRechazadas := 0, 0, 0
	actaIDsVistos := map[string]bool{}
	mesasVistas := map[string]bool{}
	lineNum := 1

	for {
		lineNum++
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			filasProcesadas++
			filasRechazadas++
			h.rechazar(auditoria.ID, lineNum, "", fmt.Sprintf("error leyendo fila: %v", err), &errores)
			continue
		}

		filasProcesadas++
		h.registrarEvento(&auditoria.ID, "FILA_CSV_LEIDA", "", gin.H{"fila": lineNum})

		// Parseo numérico (Regla 13)
		row, parseErr := parseCSVRowOficial(record, lineNum)
		if parseErr != nil {
			filasRechazadas++
			h.rechazar(auditoria.ID, lineNum, "", parseErr.Error(), &errores)
			continue
		}

		// Regla 14: acta observada por irregularidad formal (Ley N° 026)
		if row.Observacion != "" {
			msg := fmt.Sprintf("Acta OBSERVADA (Ley N° 026 del Régimen Electoral): %s", row.Observacion)
			filasRechazadas++
			h.rechazar(auditoria.ID, lineNum, row.ActaID, msg, &errores)
			continue
		}

		// Reglas 1-9
		if valErr := validateRowOficial(row); valErr != nil {
			filasRechazadas++
			h.rechazar(auditoria.ID, lineNum, row.ActaID, valErr.Error(), &errores)
			continue
		}

		// Regla 10: acta_id duplicado en CSV
		if actaIDsVistos[row.ActaID] {
			filasRechazadas++
			h.rechazar(auditoria.ID, lineNum, row.ActaID,
				fmt.Sprintf("acta_id '%s' duplicado en el CSV", row.ActaID), &errores)
			continue
		}
		// Regla 10: acta_id duplicado en BD
		var cnt int64
		h.DB.Model(&models.Acta{}).Where("codigo_acta = ?", row.ActaID).Count(&cnt)
		if cnt > 0 {
			filasRechazadas++
			h.rechazar(auditoria.ID, lineNum, row.ActaID,
				fmt.Sprintf("acta_id '%s' ya existe en la base de datos", row.ActaID), &errores)
			continue
		}
		actaIDsVistos[row.ActaID] = true

		// Regla 11: mesa duplicada en CSV
		if mesasVistas[row.Mesa] {
			filasRechazadas++
			h.rechazar(auditoria.ID, lineNum, row.ActaID,
				fmt.Sprintf("mesa '%s' ya tiene acta oficial en este CSV", row.Mesa), &errores)
			continue
		}

		// Regla 12: mesa debe existir en BD
		var mesa models.Mesa
		if err := h.DB.Where("codigo = ?", row.Mesa).First(&mesa).Error; err != nil {
			filasRechazadas++
			h.rechazar(auditoria.ID, lineNum, row.ActaID,
				fmt.Sprintf("mesa '%s' no existe en la base de datos", row.Mesa), &errores)
			continue
		}

		// Regla 11: mesa ya tiene acta oficial en BD
		h.DB.Model(&models.Acta{}).Where("id_mesa = ? AND fuente = 'OFICIAL'", mesa.ID).Count(&cnt)
		if cnt > 0 {
			filasRechazadas++
			h.rechazar(auditoria.ID, lineNum, row.ActaID,
				fmt.Sprintf("mesa '%s' ya tiene un acta oficial registrada", row.Mesa), &errores)
			continue
		}
		mesasVistas[row.Mesa] = true

		// Guardar acta oficial
		votosValidos := row.Candidato1 + row.Candidato2 + row.Candidato3 + row.Candidato4
		acta := models.Acta{
			CodigoActa:        row.ActaID,
			Fuente:            "OFICIAL",
			IDMesa:            mesa.ID,
			P1:                row.Candidato1,
			P2:                row.Candidato2,
			P3:                row.Candidato3,
			P4:                row.Candidato4,
			VotosNulos:        row.VotosNulos,
			VotosBlanco:       row.VotosBlancos,
			VotosValidos:      votosValidos,
			PapeletasNoUsadas: row.CiudadanosHabilitados - row.TotalVotos,
		}
		if err := h.DB.Create(&acta).Error; err != nil {
			filasRechazadas++
			h.rechazar(auditoria.ID, lineNum, row.ActaID,
				fmt.Sprintf("error guardando acta: %v", err), &errores)
			continue
		}

		filasValidas++
		h.registrarEvento(&auditoria.ID, "ACTA_OFICIAL_VALIDADA", row.ActaID, gin.H{
			"fila": lineNum, "mesa": row.Mesa,
		})
	}

	// Actualizar auditoría
	estado := "PROCESADO"
	if filasValidas == 0 && filasRechazadas > 0 {
		estado = "FALLIDO"
	} else if filasRechazadas > 0 {
		estado = "PROCESADO_CON_ERRORES"
	}
	h.DB.Model(&auditoria).Updates(map[string]interface{}{
		"filas_procesadas": filasProcesadas,
		"filas_validas":    filasValidas,
		"filas_rechazadas": filasRechazadas,
		"estado":           estado,
	})
	h.registrarEvento(&auditoria.ID, "CSV_PROCESADO", "", gin.H{
		"filas_procesadas": filasProcesadas,
		"filas_validas":    filasValidas,
		"filas_rechazadas": filasRechazadas,
	})

	if len(errores) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success":          true,
			"message":          "CSV procesado correctamente",
			"filas_procesadas": filasProcesadas,
			"filas_validas":    filasValidas,
			"filas_rechazadas": filasRechazadas,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"success":          false,
			"message":          "CSV procesado con errores",
			"filas_procesadas": filasProcesadas,
			"filas_validas":    filasValidas,
			"filas_rechazadas": filasRechazadas,
			"errores":          errores,
		})
	}
}

// ─── Fila de scan para GET /actas ─────────────────────────────────────────────

type actaOficialRow struct {
	CodigoActa   string `gorm:"column:codigo_acta"`
	Departamento string `gorm:"column:departamento"`
	Provincia    string `gorm:"column:provincia"`
	Municipio    string `gorm:"column:municipio"`
	Recinto      string `gorm:"column:recinto"`
	Mesa         string `gorm:"column:mesa"`
	P1           int    `gorm:"column:p1"`
	P2           int    `gorm:"column:p2"`
	P3           int    `gorm:"column:p3"`
	P4           int    `gorm:"column:p4"`
	VotosNulos   int    `gorm:"column:votos_nulos"`
	VotosBlanco  int    `gorm:"column:votos_blanco"`
	Fuente       string `gorm:"column:fuente"`
}

const sqlActasOficiales = `
SELECT a.codigo_acta, d.departamento, d.provincia, d.municipio,
       r.recinto, m.codigo AS mesa,
       a.p1, a.p2, a.p3, a.p4,
       a.votos_nulos, a.votos_blanco, a.fuente
FROM acta a
JOIN mesa m ON m.id = a.id_mesa
JOIN recinto_electoral r ON r.recinto_id = m.id_recinto_electoral
JOIN distribucion_territorial d ON d.id = r.id_distribucion_territorial
WHERE a.fuente = 'OFICIAL' AND a.fecha_eliminado IS NULL
ORDER BY a.id
`

func toDTO(row actaOficialRow) ActaOficialDTO {
	total := row.P1 + row.P2 + row.P3 + row.P4 + row.VotosNulos + row.VotosBlanco
	return ActaOficialDTO{
		ActaID:       row.CodigoActa,
		Departamento: row.Departamento,
		Provincia:    row.Provincia,
		Municipio:    row.Municipio,
		Recinto:      row.Recinto,
		Mesa:         row.Mesa,
		Candidatos: []CandidatoOficial{
			{CandidatoID: "CAND-01", Nombre: "Candidato 1", Votos: row.P1},
			{CandidatoID: "CAND-02", Nombre: "Candidato 2", Votos: row.P2},
			{CandidatoID: "CAND-03", Nombre: "Candidato 3", Votos: row.P3},
			{CandidatoID: "CAND-04", Nombre: "Candidato 4", Votos: row.P4},
		},
		VotosNulos:   row.VotosNulos,
		VotosBlancos: row.VotosBlanco,
		TotalVotos:   total,
		Estado:       "VALIDADA",
		Fuente:       row.Fuente,
	}
}

// GetActas GET /api/oficial/actas
func (h *OficialHandler) GetActas(c *gin.Context) {
	var rows []actaOficialRow
	if err := h.DB.Raw(sqlActasOficiales).Scan(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	dtos := make([]ActaOficialDTO, len(rows))
	for i, r := range rows {
		dtos[i] = toDTO(r)
	}
	c.JSON(http.StatusOK, dtos)
}

const sqlActaOficialByID = `
SELECT a.codigo_acta, d.departamento, d.provincia, d.municipio,
       r.recinto, m.codigo AS mesa,
       a.p1, a.p2, a.p3, a.p4,
       a.votos_nulos, a.votos_blanco, a.fuente
FROM acta a
JOIN mesa m ON m.id = a.id_mesa
JOIN recinto_electoral r ON r.recinto_id = m.id_recinto_electoral
JOIN distribucion_territorial d ON d.id = r.id_distribucion_territorial
WHERE a.fuente = 'OFICIAL' AND a.fecha_eliminado IS NULL AND a.codigo_acta = ?
LIMIT 1
`

// GetActaByID GET /api/oficial/actas/:acta_id
func (h *OficialHandler) GetActaByID(c *gin.Context) {
	actaID := c.Param("acta_id")
	var row actaOficialRow
	if err := h.DB.Raw(sqlActaOficialByID, actaID).Scan(&row).Error; err != nil || row.CodigoActa == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "acta no encontrada"})
		return
	}
	c.JSON(http.StatusOK, toDTO(row))
}

// GetAuditoria GET /api/oficial/auditoria
func (h *OficialHandler) GetAuditoria(c *gin.Context) {
	var list []models.OficialAuditoria
	if err := h.DB.Order("id desc").Find(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

// GetErrores GET /api/oficial/errores
func (h *OficialHandler) GetErrores(c *gin.Context) {
	var list []models.OficialError
	q := h.DB.Order("id")
	if aid := c.Query("auditoria_id"); aid != "" {
		q = q.Where("auditoria_id = ?", aid)
	}
	if err := q.Find(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

// GetEventos GET /api/oficial/eventos
func (h *OficialHandler) GetEventos(c *gin.Context) {
	var list []models.OficialEvento
	q := h.DB.Order("id")
	if aid := c.Query("auditoria_id"); aid != "" {
		q = q.Where("auditoria_id = ?", aid)
	}
	if err := q.Find(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}
