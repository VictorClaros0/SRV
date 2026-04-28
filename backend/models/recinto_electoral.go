package models

// RecintoElectoral ubicación física de votación.
type RecintoElectoral struct {
	RecintoID                 uint   `gorm:"primaryKey;column:recinto_id" json:"recintoId"`
	Recinto                   string `gorm:"size:500;not null" json:"recinto"`
	Direccion                 string `gorm:"size:500" json:"direccion"`
	Mesas                     string `gorm:"size:255" json:"mesas"`
	IDDistribucionTerritorial uint   `gorm:"column:id_distribucion_territorial;not null;index" json:"idDistribucionTerritorial"`

	Distribucion *DistribucionTerritorial `gorm:"foreignKey:IDDistribucionTerritorial" json:"distribucion,omitempty"`
	MesasList    []Mesa                   `gorm:"foreignKey:IdRecintoElectoral;references:RecintoID" json:"-"`
}

func (RecintoElectoral) TableName() string {
	return "recinto_electoral"
}
