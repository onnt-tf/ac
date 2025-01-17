package model

import (
	"time"
)

// Resource represents the resource table.
type Resource struct {
	ID          int64      `gorm:"column:id;type:int;primaryKey;autoIncrement;comment:'id'"`
	SystemCode  string     `gorm:"column:system_code;type:varchar(50);not null;default:'';index:idx_system_code_name,uniqueIndex:uk_system_code_code;comment:'system_code'"`
	Name        string     `gorm:"column:name;type:varchar(50);not null;default:'';index:idx_system_code_name;comment:'name'"`
	Code        string     `gorm:"column:code;type:varchar(50);not null;default:'';uniqueIndex:uk_system_code_code;comment:'code'"`
	ParentCode  string     `gorm:"column:parent_code;type:varchar(50);not null;default:'';index:idx_parent_code;comment:'parent_code'"`
	Description string     `gorm:"column:description;type:varchar(50);not null;default:'';comment:'description'"`
	ModifiedBy  string     `gorm:"column:modified_by;type:varchar(50);not null;default:'';comment:'modified_by'"`
	CreatedAt   time.Time  `gorm:"column:created_at;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:'created_at'"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;type:datetime;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:'updated_at'"`
	DeletedAt   *time.Time `gorm:"column:deleted_at;type:datetime;index;comment:'deleted_at'"`
}

func (Resource) TableName() string {
	return "resource"
}
