package model

import (
	"time"
)

// System represents the system table.
type System struct {
	ID          int64      `gorm:"column:id;type:int;primaryKey;autoIncrement;comment:'id'"`
	Name        string     `gorm:"column:name;type:varchar(50);not null;default:'';index:idx_name;comment:'name'"`
	Code        string     `gorm:"column:code;type:varchar(50);not null;default:'';uniqueIndex:uk_code;comment:'code'"`
	Description string     `gorm:"column:description;type:varchar(50);not null;default:'';comment:'description'"`
	ModifiedBy  string     `gorm:"column:modified_by;type:varchar(50);not null;default:'';comment:'modified_by'"`
	CreatedAt   time.Time  `gorm:"column:created_at;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:'created_at'"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;type:datetime;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:'updated_at'"`
	DeletedAt   *time.Time `gorm:"column:deleted_at;type:datetime;index;comment:'deleted_at'"`
}

func (System) TableName() string {
	return "system"
}
