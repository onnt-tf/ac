package model

import "time"

// CasbinRuleDeleted represents the casbin_rule_deleted table.
type CasbinRuleDeleted struct {
	ID        int64     `gorm:"column:id;type:int;primaryKey;autoIncrement;comment:'id'"`
	LogID     int64     `gorm:"column:log_id;type:int;not null;default:0;index:idx_log_id;comment:'casbin_rule_log ID''"`
	PType     string    `gorm:"column:ptype;type:varchar(255);not null;default:'';index:idx_ptype;comment:'ptype'"`
	V0        string    `gorm:"column:v0;type:varchar(255);not null;default:'';index:idx_v0;comment:'v0'"`
	V1        string    `gorm:"column:v1;type:varchar(255);not null;default:'';index:idx_v1;comment:'v1'"`
	V2        string    `gorm:"column:v2;type:varchar(255);not null;default:'';comment:'v2'"`
	V3        string    `gorm:"column:v3;type:varchar(255);not null;default:'';comment:'v3'"`
	V4        string    `gorm:"column:v4;type:varchar(255);not null;default:'';comment:'v4'"`
	V5        string    `gorm:"column:v5;type:varchar(255);not null;default:'';comment:'v5'"`
	CreatedAt time.Time `gorm:"column:created_at;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:'created_at'"`
}

func (CasbinRuleDeleted) TableName() string {
	return "casbin_rule_deleted"
}
