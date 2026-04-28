package models

// DistribucionTerritorial jerarquía geográfica.
type DistribucionTerritorial struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	Departamento string `gorm:"size:255;not null" json:"departamento"`
	Municipio    string `gorm:"size:255;not null" json:"municipio"`
	Provincia    string `gorm:"size:255;not null" json:"provincia"`
}

func (DistribucionTerritorial) TableName() string {
	return "distribucion_territorial"
}
