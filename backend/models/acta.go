package models

import (
	"time"

	"gorm.io/gorm"
)

// Acta conteo de votos por mesa.
type Acta struct {
	ID                uint           `gorm:"primaryKey" json:"id"`
	CodigoActa        int64          `gorm:"column:codigo_acta;uniqueIndex" json:"codigoActa"`
	CodigoRecinto     int64          `gorm:"column:codigo_recinto;index" json:"codigoRecinto"`
	NroMesa           int            `gorm:"column:nro_mesa" json:"nroMesa"`
	Estado            string         `gorm:"column:estado;default:'impresa'" json:"estado"`
	PapeletasAnfora   int            `gorm:"column:papeletas_anfora" json:"papeletasAnfora"`
	PapeletasNoUsadas int            `gorm:"column:papeletas_no_usadas" json:"papeletasNoUsadas"`
	P1                int            `gorm:"column:p1;not null" json:"p1"`
	P2                int            `gorm:"column:p2;not null" json:"p2"`
	P3                int            `gorm:"column:p3;not null" json:"p3"`
	P4                int            `gorm:"column:p4;not null" json:"p4"`
	VotosNulos        int            `gorm:"column:votos_nulos;not null" json:"votosNulos"`
	VotosBlanco       int            `gorm:"column:votos_blanco;not null" json:"votosBlanco"`
	VotosValidos      int            `gorm:"column:votos_validos;not null" json:"votosValidos"`
	Observaciones     string         `gorm:"column:observaciones;size:1000" json:"observaciones"`
	AperturaHora      int            `gorm:"column:apertura_hora" json:"aperturaHora"`
	AperturaMinutos   int            `gorm:"column:apertura_minutos" json:"aperturaMinutos"`
	CierreHora        int            `gorm:"column:cierre_hora" json:"cierreHora"`
	CierreMinutos     int            `gorm:"column:cierre_minutos" json:"cierreMinutos"`
	IDMesa            *uint          `gorm:"column:id_mesa;index" json:"idMesa,omitempty"`
	FechaCreacion     time.Time      `gorm:"column:fecha_creacion;autoCreateTime" json:"fechaCreacion"`
	CreadoPorID       *uint          `gorm:"column:creado_por" json:"creadoPor,omitempty"`
	FechaModificacion *time.Time     `gorm:"column:fecha_modificacion" json:"fechaModificacion,omitempty"`
	ModificadoPorID   *uint          `gorm:"column:modificado_por" json:"modificadoPor,omitempty"`
	DeletedAt         gorm.DeletedAt `gorm:"column:fecha_eliminado;index" json:"-"`
	EliminadoPorID    *uint          `gorm:"column:eliminado_por" json:"eliminadoPor,omitempty"`

	CreadoPor     *Usuario `gorm:"foreignKey:CreadoPorID" json:"-"`
	ModificadoPor *Usuario `gorm:"foreignKey:ModificadoPorID" json:"-"`
	EliminadoPor  *Usuario `gorm:"foreignKey:EliminadoPorID" json:"-"`
	Mesa          *Mesa    `gorm:"foreignKey:IDMesa" json:"mesa,omitempty"`
}

func (Acta) TableName() string {
	return "acta"
}
