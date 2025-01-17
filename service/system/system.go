package system

import (
	"ac/bootstrap/database"
	"ac/bootstrap/logger"
	"ac/custom/define"
	"ac/dal"
	"ac/model"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type System struct {
	ID          int64     `json:"ID"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	Description string    `json:"description"`
	ModifiedBy  string    `json:"modified_by"`
	UpdatedAt   time.Time `json:"update_at"`
}

func Validate(ctx echo.Context, code string) (bool, error) {
	if code == "" {
		return false, errors.New("code is empty")
	}
	record, err := dal.NewRepo[model.System]().Query(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Where(model.System{Code: code})
	})
	if err != nil {
		return false, fmt.Errorf("failed to query, err: %w, code: %s", err, code)
	}
	if record == nil {
		logger.Infof(ctx, "no matching system found, code: %s", code)
		return false, nil
	}
	return true, nil
}

func IsCodeAvailable(ctx echo.Context, code string) (bool, error) {
	if code == "" {
		return false, errors.New("code is empty")
	}
	if !strings.HasPrefix(code, define.PrefixSystem) {
		return false, fmt.Errorf("code must start with the prefix '%s'", define.PrefixSystem)
	}
	record, err := dal.NewRepo[model.System]().Query(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Where(model.System{Code: code})
	})
	if err != nil {
		return false, fmt.Errorf("failed to query, err: %w, code: %s", err, code)
	}
	return record == nil, nil
}

// func QueryByCode(ctx echo.Context, code string) (*System, error) {
// 	if code == "" {
// 		return nil, errors.New("code is empty")
// 	}
// 	record, err := dal.NewRepo[model.System]().Query(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
// 		return db.Where(model.System{Code: code})
// 	})
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to query, err: %w, code: %s", err, code)
// 	}
// 	return &System{
// 		ID:          record.ID,
// 		Name:        record.Name,
// 		Code:        record.Code,
// 		Description: record.Description,
// 		ModifiedBy:  record.ModifiedBy,
// 		UpdatedAt:   record.UpdatedAt,
// 	}, nil
// }
