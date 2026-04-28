package models

// Mesa mesa de votación.
type Mesa struct {
	ID                 uint   `gorm:"primaryKey;column:id" json:"id"`
	Codigo             string `gorm:"size:100;not null;index" json:"codigo"`
	CantidadHabilitada int    `gorm:"column:cantidad_habilitada;not null" json:"cantidadHabilitada"`
	IdRecintoElectoral uint   `gorm:"column:id_recinto_electoral;not null;index" json:"idRecintoElectoral"`

	Recinto *RecintoElectoral `gorm:"foreignKey:IdRecintoElectoral;references:RecintoID" json:"recinto,omitempty"`
	Actas   []Acta            `gorm:"foreignKey:IDMesa" json:"-"`
}

func (Mesa) TableName() string {
	return "mesa"
}
