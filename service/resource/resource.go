package resource

import (
	"ac/bootstrap/database"
	"ac/bootstrap/logger"
	"ac/custom/define"
	"ac/custom/util"
	"ac/dal"
	"ac/model"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type Resource struct {
	ID          int64     `json:"ID"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	SystemCode  string    `json:"system_code"`
	Code        string    `json:"code"`
	ModifiedBy  string    `json:"modified_by"`
	UpdatedAt   time.Time `json:"update_at"`
}

func Validate(ctx echo.Context, systemCode, code string) (bool, error) {
	if systemCode == "" || code == "" {
		return false, errors.New("systemCode or code is empty")
	}
	record, err := dal.NewRepo[model.Resource]().Query(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Where(model.Resource{SystemCode: systemCode, Code: code})
	})
	if err != nil {
		logger.Errorf(ctx, "failed to query, err: %v", err)
		return false, fmt.Errorf("failed to query, err: %w, system code: %s, code: %s", err, systemCode, code)
	}
	if record == nil {
		logger.Infof(ctx, "no matching user found, system code: %s, code: %s", systemCode, code)
		return false, nil
	}
	return true, nil
}

func ValidateBatch(ctx echo.Context, systemCode string, codeList []string) (map[string]bool, error) {
	if systemCode == "" || len(codeList) == 0 {
		return nil, errors.New("systemCode or code is empty")
	}
	codeList = slices.DeleteFunc(codeList, func(s string) bool {
		return strings.TrimSpace(s) == ""
	})
	codeList = util.Deduplicate(codeList)

	if len(codeList) == 0 {
		return nil, errors.New("codeList is empty")
	}

	for _, v := range codeList {
		if !strings.HasPrefix(v, define.PrefixResource) {
			return nil, fmt.Errorf("code must start with the prefix '%s'", define.PrefixUser)
		}
	}

	recordList, err := dal.NewRepo[model.Resource]().QueryList(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Where(&model.Resource{
			SystemCode: systemCode,
		}).Where("code IN ?", codeList)
	})
	if err != nil {
		logger.Errorf(ctx, "failed to query, err: %v, system code: %s, codeList: %+v", err, systemCode, codeList)
		return nil, fmt.Errorf("failed to query, err: %w", err)
	}

	result := make(map[string]bool, len(codeList))
	for _, code := range codeList {
		result[code] = false
	}
	for _, record := range recordList {
		result[record.Code] = true
	}
	return result, nil
}

func IsCodeAvailable(ctx echo.Context, code string) (bool, error) {
	if code == "" {
		return false, errors.New("code is empty")
	}
	if !strings.HasPrefix(code, define.PrefixResource) {
		return false, fmt.Errorf("code must start with the prefix '%s'", define.PrefixResource)
	}
	record, err := dal.NewRepo[model.Resource]().Query(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Where(model.Resource{Code: code})
	})
	if err != nil {
		return false, fmt.Errorf("failed to query, err: %w, code: %s", err, code)
	}
	return record == nil, nil
}

func QueryResourceByCode(ctx echo.Context, systemCode string, codeList []string) (map[string]Resource, error) {
	codeList = util.Deduplicate(slices.DeleteFunc(codeList, func(s string) bool {
		return strings.TrimSpace(s) == ""
	}))
	resourceList, err := dal.NewRepo[model.Resource]().QueryList(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Where(model.Resource{SystemCode: systemCode}).Where("code IN ?", codeList)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query resources: %w", err)
	}
	resourceCodeMap := make(map[string]Resource)
	for _, v := range resourceList {
		resourceCodeMap[v.Code] = Resource{
			ID:          v.ID,
			Name:        v.Name,
			Description: v.Description,
			SystemCode:  v.SystemCode,
			Code:        v.Code,
			ModifiedBy:  v.ModifiedBy,
			UpdatedAt:   v.UpdatedAt,
		}
	}
	return resourceCodeMap, nil
}
