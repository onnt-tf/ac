package model

import "time"

const OperateAdd = "add"
const OperateDelete = "delete"
const OperateSet = "set"

// CasbinRuleLog represents the casbin_rule_log table.
type CasbinRuleLog struct {
	ID         int64     `gorm:"column:id;type:int;primaryKey;autoIncrement;comment:'id'"`
	Operate    string    `gorm:"column:operate;type:enum('add','delete','set');not null;default:add;comment:'operate'"`
	Content    string    `gorm:"column:content;type:text;not null;comment:'content'"`
	ModifiedBy string    `gorm:"column:modified_by;type:varchar(50);not null;default:'';comment:'modified_by'"`
	CreatedAt  time.Time `gorm:"column:created_at;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:'created_at'"`
}

func (CasbinRuleLog) TableName() string {
	return "casbin_rule_log"
}
