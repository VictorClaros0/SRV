package models

import "time"

// OficialAuditoria registra cada carga de CSV oficial.
type OficialAuditoria struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	NombreArchivo   string    `gorm:"size:500;not null" json:"nombre_archivo"`
	Usuario         string    `gorm:"size:255;not null" json:"usuario"`
	FechaCarga      time.Time `gorm:"autoCreateTime" json:"fecha_carga"`
	FilasProcesadas int       `json:"filas_procesadas"`
	FilasValidas    int       `json:"filas_validas"`
	FilasRechazadas int       `json:"filas_rechazadas"`
	Estado          string    `gorm:"size:50;not null" json:"estado"`
}

func (OficialAuditoria) TableName() string { return "oficial_auditoria" }

// OficialError almacena errores de validación por fila.
type OficialError struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	AuditoriaID uint      `gorm:"not null;index" json:"auditoria_id"`
	Fila        int       `gorm:"not null" json:"fila"`
	ActaID      string    `gorm:"size:100" json:"acta_id"`
	Error       string    `gorm:"size:1000;not null" json:"error"`
	CreatedAt   time.Time `json:"created_at"`
}

func (OficialError) TableName() string { return "oficial_error" }

// OficialEvento registra los eventos del flujo de carga oficial.
type OficialEvento struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	AuditoriaID *uint     `gorm:"index" json:"auditoria_id,omitempty"`
	Tipo        string    `gorm:"size:100;not null" json:"tipo"`
	ActaID      string    `gorm:"size:100" json:"acta_id,omitempty"`
	Payload     string    `gorm:"type:jsonb" json:"payload,omitempty"`
	Timestamp   time.Time `gorm:"autoCreateTime" json:"timestamp"`
}

func (OficialEvento) TableName() string { return "oficial_evento" }
