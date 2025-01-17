package user_role

import (
	"ac/bootstrap/database"
	"ac/bootstrap/logger"
	"ac/controller"
	"ac/custom/define"
	"ac/custom/input"
	"ac/custom/output"
	"ac/custom/util"
	"ac/dal"
	"ac/model"
	"fmt"
	"slices"
	"strings"

	"ac/service/subject"
	"ac/service/system"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func RegisterRoutes(g *echo.Group) {
	g.POST("/add", addItem)
	g.POST("/delete", deleteItem)
	g.GET("/query", query)
}

func addItem(ctx echo.Context) error {
	body := struct {
		SystemCode   string   `json:"system_code" validate:"required,gt=0"`
		UserCode     string   `json:"user_code" validate:"required,gt=0"`
		RoleCodeList []string `json:"role_code_list" validate:"required,gt=0,dive,required,gt=0"`
	}{}
	if err := input.BindAndValidate(ctx, &body); err != nil {
		return output.Failure(ctx, controller.ErrInvalidInput.WithMsg(err.Error()))
	}

	body.RoleCodeList = util.Deduplicate(slices.DeleteFunc(body.RoleCodeList, func(s string) bool {
		return strings.TrimSpace(s) == ""
	}))

	if ok, err := system.Validate(ctx, body.SystemCode); !ok {
		if err != nil {
			logger.Errorf(ctx, "failed to validate system, err: %v, code: %s", err, body.SystemCode)
		}
		return output.Failure(ctx, controller.ErrSystemError.WithHint("Invalid system code"))
	}
	if ok, err := subject.ValidateUser(ctx, body.SystemCode, body.UserCode); !ok {
		if err != nil {
			logger.Errorf(ctx, "failed to validate user, err: %v, system code: %s, code: %s", err, body.SystemCode, body.UserCode)
		}
		return output.Failure(ctx, controller.ErrSystemError.WithHint("Invalid user code"))
	}

	validateResult, err := subject.ValidateRoleBatch(ctx, body.SystemCode, body.RoleCodeList)
	if err != nil {
		logger.Errorf(ctx, "failed to validate role, err: %v", err)
		return output.Failure(ctx, controller.ErrSystemError)
	}

	for _, v := range body.RoleCodeList {
		if valid, ok := validateResult[v]; ok && valid {
			continue
		}
		return output.Failure(ctx, controller.ErrSystemError.WithHint("Invalid role code"))
	}

	ruleList, err := dal.NewRepo[model.CasbinRule]().QueryList(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Where(model.CasbinRule{PType: model.PTypeGroup, V0: body.UserCode}).Where("v1 IN ?", body.RoleCodeList)
	})
	if err != nil {
		logger.Errorf(ctx, "failed to query user role, err: %v", err)
		return output.Failure(ctx, controller.ErrSystemError)
	}
	if len(ruleList) > 0 {
		return output.Failure(ctx, controller.ErrSystemError.WithHint("User's roles have been updated. Please refresh and try again."))
	}
	valuesToAdd := make([]*model.CasbinRule, 0, len(body.RoleCodeList))
	for _, v := range body.RoleCodeList {
		valuesToAdd = append(valuesToAdd, &model.CasbinRule{
			PType: model.PTypeGroup,
			V0:    body.UserCode,
			V1:    v,
		})
	}
	err = database.DB.WithContext(ctx.Request().Context()).Transaction(func(tx *gorm.DB) error {
		err := dal.NewRepo[model.CasbinRule]().BatchInsert(ctx, tx, valuesToAdd, 20)
		if err != nil {
			return fmt.Errorf("failed to add casin rule, err: %v", err)
		}
		return nil
	})
	if err != nil {
		logger.Errorf(ctx, "failed to commit, err: %v", err)
		return output.Failure(ctx, controller.ErrSystemError)
	}
	return output.Success(ctx, nil)
}

func deleteItem(ctx echo.Context) error {
	body := struct {
		SystemCode   string   `json:"system_code" validate:"required,gt=0"`
		UserCode     string   `json:"user_code" validate:"required,gt=0"`
		RoleCodeList []string `json:"role_code_list" validate:"required,gt=0,dive,required,gt=0"`
	}{}
	if err := input.BindAndValidate(ctx, &body); err != nil {
		return output.Failure(ctx, controller.ErrInvalidInput.WithMsg(err.Error()))
	}
	body.RoleCodeList = util.Deduplicate(slices.DeleteFunc(body.RoleCodeList, func(s string) bool {
		return strings.TrimSpace(s) == ""
	}))

	if ok, err := system.Validate(ctx, body.SystemCode); !ok {
		if err != nil {
			logger.Errorf(ctx, "failed to validate system, err: %v, code: %s", err, body.SystemCode)
		}
		return output.Failure(ctx, controller.ErrSystemError.WithHint("Invalid system code"))
	}
	if ok, err := subject.ValidateUser(ctx, body.SystemCode, body.UserCode); !ok {
		if err != nil {
			logger.Errorf(ctx, "failed to validate user, err: %v, system code: %s, code: %s", err, body.SystemCode, body.UserCode)
		}
		return output.Failure(ctx, controller.ErrSystemError.WithHint("Invalid user code"))
	}

	validateResult, err := subject.ValidateRoleBatch(ctx, body.SystemCode, body.RoleCodeList)
	if err != nil {
		logger.Errorf(ctx, "failed to validate role, err: %v", err)
		return output.Failure(ctx, controller.ErrSystemError)
	}

	for _, v := range body.RoleCodeList {
		if valid, ok := validateResult[v]; ok && valid {
			continue
		}
		return output.Failure(ctx, controller.ErrSystemError.WithHint("Invalid role code"))
	}

	ruleList, err := dal.NewRepo[model.CasbinRule]().QueryList(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Where(model.CasbinRule{PType: model.PTypeGroup, V0: body.UserCode}).Where("v1 IN ?", body.RoleCodeList)
	})
	if err != nil {
		logger.Errorf(ctx, "failed to query user role, err: %v", err)
		return output.Failure(ctx, controller.ErrSystemError)
	}
	if len(ruleList) != len(body.RoleCodeList) {
		return output.Failure(ctx, controller.ErrSystemError.WithHint("User's roles have been updated. Please refresh and try again."))
	}

	err = database.DB.WithContext(ctx.Request().Context()).Transaction(func(tx *gorm.DB) error {
		err := dal.NewRepo[model.CasbinRule]().Delete(ctx, tx, func(db *gorm.DB) *gorm.DB {
			return db.Where(ruleList).Limit(len(ruleList))
		})
		if err != nil {
			return fmt.Errorf("failed to add casin rule, err: %v", err)
		}
		return nil
	})
	if err != nil {
		logger.Errorf(ctx, "failed to commit, err: %v", err)
		return output.Failure(ctx, controller.ErrSystemError)
	}
	return output.Success(ctx, nil)
}

func query(ctx echo.Context) error {
	body := struct {
		Page         int      `json:"page" validate:"required,gt=0"`
		PageSize     int      `json:"page_size" validate:"required,gt=0"`
		SystemCode   string   `json:"system_code" validate:"required,gt=0"`
		UserCodeList []string `json:"user_code_list"`
		RoleCodeList []string `json:"role_code_list"`
	}{}
	if err := input.BindAndValidate(ctx, &body); err != nil {
		return output.Failure(ctx, controller.ErrInvalidInput.WithMsg(err.Error()))
	}
	if len(body.UserCodeList) > 0 && len(body.RoleCodeList) > 0 {
		return output.Failure(ctx, controller.ErrInvalidInput.WithMsg("both 'UserCodeList' and 'RoleCodeList' cannot be provided simultaneously"))
	}
	body.UserCodeList = util.Deduplicate(slices.DeleteFunc(body.UserCodeList, func(s string) bool {
		return strings.TrimSpace(s) == ""
	}))
	body.RoleCodeList = util.Deduplicate(slices.DeleteFunc(body.RoleCodeList, func(s string) bool {
		return strings.TrimSpace(s) == ""
	}))
	if ok, err := system.Validate(ctx, body.SystemCode); !ok {
		if err != nil {
			logger.Errorf(ctx, "failed to validate system, err: %v, code: %s", err, body.SystemCode)
		}
		return output.Failure(ctx, controller.ErrSystemError.WithHint("Invalid system code"))
	}

	if len(body.UserCodeList) > 0 {
		validateResult, err := subject.ValidateUserBatch(ctx, body.SystemCode, body.UserCodeList)
		if err != nil {
			logger.Errorf(ctx, "failed to validate user, err: %v", err)
			return output.Failure(ctx, controller.ErrSystemError)
		}

		for _, v := range body.UserCodeList {
			if valid, ok := validateResult[v]; ok && valid {
				continue
			}
			return output.Failure(ctx, controller.ErrSystemError.WithHint("Invalid user code"))
		}
	}

	if len(body.RoleCodeList) > 0 {
		validateResult, err := subject.ValidateRoleBatch(ctx, body.SystemCode, body.RoleCodeList)
		if err != nil {
			logger.Errorf(ctx, "failed to validate role, err: %v", err)
			return output.Failure(ctx, controller.ErrSystemError)
		}

		for _, v := range body.RoleCodeList {
			if valid, ok := validateResult[v]; ok && valid {
				continue
			}
			return output.Failure(ctx, controller.ErrSystemError.WithHint("Invalid role code"))
		}
	}

	ruleList, err := dal.NewRepo[model.CasbinRule]().QueryList(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		if len(body.UserCodeList) > 0 {
			db.Where(model.CasbinRule{PType: model.PTypeGroup}).Where("v0 IN ?", body.UserCodeList)
		}
		if len(body.RoleCodeList) > 0 {
			db.Where(model.CasbinRule{PType: model.PTypeGroup}).Where("v1 IN ?", body.RoleCodeList)
		}
		db.Order("id desc")
		return db
	}, dal.Paginate(body.Page, body.PageSize))
	if err != nil {
		logger.Errorf(ctx, "failed to query, err: %v", err)
		return output.Failure(ctx, controller.ErrSystemError)
	}

	count, err := dal.NewRepo[model.CasbinRule]().Count(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		if len(body.UserCodeList) > 0 {
			db.Where(model.CasbinRule{PType: model.PTypeGroup}).Where("v0 IN ?", body.UserCodeList)
		}
		if len(body.RoleCodeList) > 0 {
			db.Or(db.Where(model.CasbinRule{PType: model.PTypeGroup}).Where("v1 IN ?", body.RoleCodeList))
		}
		return db
	}, dal.Paginate(body.Page, body.PageSize))
	if err != nil {
		logger.Errorf(ctx, "failed to query, err: %v", err)
		return output.Failure(ctx, controller.ErrSystemError)
	}

	fmt.Println(111)
	tmpCodeList := make([]string, 0, len(ruleList))
	for _, v := range ruleList {
		if strings.HasPrefix(v.V0, define.PrefixUser) || strings.HasPrefix(v.V0, define.PrefixRole) {
			tmpCodeList = append(tmpCodeList, v.V0)
		}
		if strings.HasPrefix(v.V1, define.PrefixUser) || strings.HasPrefix(v.V1, define.PrefixRole) {
			tmpCodeList = append(tmpCodeList, v.V1)
		}
	}
	recordList, err := dal.NewRepo[model.Subject]().QueryList(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Where(model.Subject{SystemCode: body.SystemCode}).Where("code IN ?", tmpCodeList)
	})
	if err != nil {
		logger.Errorf(ctx, "failed to query, err: %v", err)
		return output.Failure(ctx, controller.ErrSystemError)
	}
	subjectCode2Name := make(map[string]string)
	for _, v := range recordList {
		subjectCode2Name[v.Code] = v.Name
	}

	type Rule struct {
		SystemCode string `json:"system_code"`
		UserCode   string `json:"user_code"`
		UserName   string `json:"user_name"`
		RoleCode   string `json:"role_code"`
		RoleName   string `json:"role_name"`
	}
	list := make([]Rule, 0, len(ruleList))
	for _, v := range ruleList {
		rule := Rule{
			SystemCode: body.SystemCode,
			UserCode:   v.V0,
			RoleCode:   v.V1,
		}
		if name, ok := subjectCode2Name[v.V0]; ok {
			rule.UserName = name
		}
		if name, ok := subjectCode2Name[v.V1]; ok {
			rule.RoleName = name
		}
		list = append(list, rule)
	}
	return output.Success(ctx, map[string]interface{}{
		"total": count,
		"list":  list,
	})
}
