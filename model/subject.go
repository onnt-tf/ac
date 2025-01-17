package model

import (
	"time"
)

const (
	SubjectTypeUser = "user"
	SubjectTypeRole = "role"
)

// Subject represents the subject table.
type Subject struct {
	ID          int64      `gorm:"column:id;type:int;primaryKey;autoIncrement;comment:'id'"`
	SystemCode  string     `gorm:"column:system_code;type:varchar(50);not null;default:'';index:idx_system_code_name_type;comment:'id'"`
	Type        string     `gorm:"column:type;type:enum('user','role');not null;default:user;index:idx_system_code_name_type;comment:'type'"`
	Name        string     `gorm:"column:name;type:varchar(50);not null;default:'';index:idx_system_code_name_type;comment:'name'"`
	Code        string     `gorm:"column:code;type:varchar(50);not null;default:'';uniqueIndex:uk_system_code_code;comment:'code'"`
	Description string     `gorm:"column:description;type:varchar(50);not null;default:'';comment:'description'"`
	ModifiedBy  string     `gorm:"column:modified_by;type:varchar(50);not null;default:'';comment:'modified_by'"`
	CreatedAt   time.Time  `gorm:"column:created_at;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:'created_at'"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;type:datetime;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:'updated_at'"`
	DeletedAt   *time.Time `gorm:"column:deleted_at;type:datetime;index;comment:'deleted_at'"`
}

func (Subject) TableName() string {
	return "subject"
}
