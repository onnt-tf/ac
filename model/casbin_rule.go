package model

const PTypePolicy = "p"
const PTypeGroup = "g"

// CasbinRule represents the casbin_rule table.
type CasbinRule struct {
	ID    int64  `gorm:"column:id;type:int;primaryKey;autoIncrement;comment:'id'"`
	PType string `gorm:"column:ptype;type:varchar(255);not null;default:'';uniqueIndex:uk_ptype_v0_v1;index:idx_ptype;comment:'ptype'"`
	V0    string `gorm:"column:v0;type:varchar(255);not null;default:'';uniqueIndex:uk_ptype_v0_v1;index:idx_v0;comment:'v0'"`
	V1    string `gorm:"column:v1;type:varchar(255);not null;default:'';uniqueIndex:uk_ptype_v0_v1;index:idx_v1;comment:'v1'"`
	V2    string `gorm:"column:v2;type:varchar(255);not null;default:'';comment:'v2'"`
	V3    string `gorm:"column:v3;type:varchar(255);not null;default:'';comment:'v3'"`
	V4    string `gorm:"column:v4;type:varchar(255);not null;default:'';comment:'v4'"`
	V5    string `gorm:"column:v5;type:varchar(255);not null;default:'';comment:'v5'"`
}

func (CasbinRule) TableName() string {
	return "casbin_rule"
}
