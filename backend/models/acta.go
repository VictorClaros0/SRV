package models

import (
	"time"

	"gorm.io/gorm"
)

// Acta conteo de votos por mesa.
type Acta struct {
	ID                 uint           `gorm:"primaryKey" json:"id"`
	PapeletasNoUsadas  int            `gorm:"column:papeletas_no_usadas;not null" json:"papeletasNoUsadas"`
	P1                 int            `gorm:"column:p1;not null" json:"p1"`
	P2                 int            `gorm:"column:p2;not null" json:"p2"`
	P3                 int            `gorm:"column:p3;not null" json:"p3"`
	P4                 int            `gorm:"column:p4;not null" json:"p4"`
	VotosNulos         int            `gorm:"column:votos_nulos;not null" json:"votosNulos"`
	VotosBlanco        int            `gorm:"column:votos_blanco;not null" json:"votosBlanco"`
	VotosValidos       int            `gorm:"column:votos_validos;not null" json:"votosValidos"`
	IDMesa             uint           `gorm:"column:id_mesa;not null;index" json:"idMesa"`
	FechaCreacion      time.Time      `gorm:"column:fecha_creacion;autoCreateTime" json:"fechaCreacion"`
	CreadoPorID        *uint          `gorm:"column:creado_por" json:"creadoPor,omitempty"`
	FechaModificacion  *time.Time     `gorm:"column:fecha_modificacion" json:"fechaModificacion,omitempty"`
	ModificadoPorID    *uint          `gorm:"column:modificado_por" json:"modificadoPor,omitempty"`
	DeletedAt          gorm.DeletedAt `gorm:"column:fecha_eliminado;index" json:"-"`
	EliminadoPorID     *uint          `gorm:"column:eliminado_por" json:"eliminadoPor,omitempty"`

	CreadoPor     *Usuario `gorm:"foreignKey:CreadoPorID" json:"-"`
	ModificadoPor *Usuario `gorm:"foreignKey:ModificadoPorID" json:"-"`
	EliminadoPor  *Usuario `gorm:"foreignKey:EliminadoPorID" json:"-"`
	Mesa          *Mesa    `gorm:"foreignKey:IDMesa" json:"mesa,omitempty"`
}

func (Acta) TableName() string {
	return "acta"
}
