package subject

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

type Subject struct {
	ID          int64     `json:"ID"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	SystemCode  string    `json:"system_code"`
	Code        string    `json:"code"`
	ModifiedBy  string    `json:"modified_by"`
	UpdatedAt   time.Time `json:"update_at"`
}

func Validate(ctx echo.Context, systemCode, code string) (bool, error) {
	return validate(ctx, systemCode, "", code)
}

func validate(ctx echo.Context, systemCode, subjectType, code string) (bool, error) {
	if systemCode == "" || code == "" {
		return false, errors.New("systemCode or code is empty")
	}

	condition := &model.Subject{SystemCode: systemCode, Code: code, Type: subjectType}

	// if subjectType == model.SubjectTypeUser && !strings.HasPrefix(code, define.PrefixUser) {
	// 	return false, fmt.Errorf("code must start with the prefix: %s", define.PrefixUser)
	// }
	// if subjectType == model.SubjectTypeRole && !strings.HasPrefix(code, define.PrefixRole) {
	// 	return false, fmt.Errorf("code must start with the prefix: %s", define.PrefixRole)
	// }

	record, err := dal.NewRepo[model.Subject]().Query(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Where(condition)
	})
	if err != nil {
		logger.Errorf(ctx, "failed to query, err: %v", err)
		return false, fmt.Errorf("failed to query, err: %w, system code: %s, code: %s", err, systemCode, code)
	}
	if record == nil {
		logger.Infof(ctx, "no matching subject found, system code: %s, code: %s", systemCode, code)
		return false, nil
	}
	return true, nil
}

func validateBatch(ctx echo.Context, systemCode string, subjectType string, codeList []string) (map[string]bool, error) {
	if systemCode == "" || len(codeList) == 0 {
		return nil, errors.New("systemCode or codeList is empty")
	}

	codeList = util.Deduplicate(slices.DeleteFunc(codeList, func(s string) bool {
		return strings.TrimSpace(s) == ""
	}))

	if len(codeList) == 0 {
		return nil, errors.New("codeList is empty")
	}

	// for _, v := range codeList {
	// 	if subjectType == model.SubjectTypeUser && !strings.HasPrefix(v, define.PrefixUser) {
	// 		return nil, fmt.Errorf("code must start with the prefix: %s", define.PrefixUser)
	// 	}
	// 	if subjectType == model.SubjectTypeRole && !strings.HasPrefix(v, define.PrefixRole) {
	// 		return nil, fmt.Errorf("code must start with the prefix: %s", define.PrefixRole)
	// 	}
	// }

	recordList, err := dal.NewRepo[model.Subject]().QueryList(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Where(&model.Subject{
			SystemCode: systemCode,
			Type:       subjectType,
		}).Where("code IN ?", codeList)
	})
	if err != nil {
		logger.Errorf(ctx, "failed to query, err: %v, system code: %s, codeList: %+v, type: %s", err, systemCode, codeList, subjectType)
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

func isCodeAvailable(ctx echo.Context, subjectType, code string) (bool, error) {
	if code == "" {
		return false, errors.New("code is empty")
	}
	if subjectType == model.SubjectTypeUser && !strings.HasPrefix(code, define.PrefixUser) {
		return false, fmt.Errorf("code must start with the prefix: %s", define.PrefixUser)
	}
	if subjectType == model.SubjectTypeRole && !strings.HasPrefix(code, define.PrefixRole) {
		return false, fmt.Errorf("code must start with the prefix: %s", define.PrefixRole)
	}
	record, err := dal.NewRepo[model.Subject]().Query(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Where(model.Subject{Code: code})
	})
	if err != nil {
		return false, fmt.Errorf("failed to query, err: %w, code: %s", err, code)
	}
	return record == nil, nil
}
