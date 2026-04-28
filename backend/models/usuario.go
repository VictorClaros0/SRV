package models

import (
	"time"

	"gorm.io/gorm"
)

// Usuario refleja la entidad de la imagen con auditoría y autoreferencias.
type Usuario struct {
	ID                uint           `gorm:"primaryKey" json:"id"`
	Nombre            string         `gorm:"size:255;not null" json:"nombre"`
	Apellido          string         `gorm:"size:255;not null" json:"apellido"`
	SegundoApellido   string         `gorm:"size:255" json:"segundoApellido"`
	NombreUsuario     string         `gorm:"column:usuario;uniqueIndex;size:100;not null" json:"usuario"`
	Contrasena        string         `gorm:"column:contrasena;size:255;not null" json:"-"`
	EsAdmin           bool           `gorm:"column:es_admin;default:false" json:"esAdmin"`
	FechaCreacion     time.Time      `gorm:"column:fecha_creacion;autoCreateTime" json:"fechaCreacion"`
	CreadoPorID       *uint          `gorm:"column:creado_por" json:"creadoPor,omitempty"`
	FechaModificacion *time.Time     `gorm:"column:fecha_modificacion" json:"fechaModificacion,omitempty"`
	ModificadoPorID   *uint          `gorm:"column:modificado_por" json:"modificadoPor,omitempty"`
	DeletedAt         gorm.DeletedAt `gorm:"column:fecha_eliminado;index" json:"-"`
	EliminadoPorID    *uint          `gorm:"column:eliminado_por" json:"eliminadoPor,omitempty"`

	CreadoPor     *Usuario `gorm:"foreignKey:CreadoPorID" json:"-"`
	ModificadoPor *Usuario `gorm:"foreignKey:ModificadoPorID" json:"-"`
	EliminadoPor  *Usuario `gorm:"foreignKey:EliminadoPorID" json:"-"`
}

func (Usuario) TableName() string {
	return "usuario"
}
