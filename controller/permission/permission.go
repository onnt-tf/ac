package permission

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
	"ac/service/casbin"
	"ac/service/resource"
	"ac/service/rule"
	"ac/service/subject"
	"ac/service/system"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func RegisterRoutes(g *echo.Group) {
	g.POST("/add", addItem)
	g.POST("/delete", deleteItem)
	g.GET("/query", query)
}

type Permission struct {
	ResourceIndex string `json:"resource_index" validate:"required,gt=0"`
	Action        string `json:"action" validate:"required,gt=0"`
	BeginTime     int64  `json:"begin_time" validate:"required,gt=0"`
	EndTime       int64  `json:"end_time" validate:"required,gt=0"`
}

func addItem(ctx echo.Context) error {
	body := struct {
		SystemCode     string       `json:"system_code" validate:"required,gt=0"`
		SubjectCode    string       `json:"subject_code" validate:"required,gt=0"`
		PermissionList []Permission `json:"permission_list" validate:"required,gt=0,dive,required"`
		Inherit        bool         `json:"inherit"`
	}{}
	if err := input.BindAndValidate(ctx, &body); err != nil {
		return output.Failure(ctx, controller.ErrInvalidInput.WithMsg(err.Error()))
	}

	if ok, err := validateSystemAndSubject(ctx, body.SystemCode, body.SubjectCode); !ok {
		if err != nil {
			logger.Errorf(ctx, "failed to validate, err: %v, system code: %s, user code: %s", err, body.SystemCode, body.SubjectCode)
		}
		return output.Failure(ctx, controller.ErrSystemError)
	}

	tmpPermissionList, err := validatePermissionList(ctx, body.SystemCode, body.PermissionList)
	if err != nil {
		logger.Errorf(ctx, "failed to validate, err: %v", err)
		return output.Failure(ctx, controller.ErrSystemError)
	}

	ruleToAdd := make([]rule.Rule, 0, len(tmpPermissionList))

	for _, v := range tmpPermissionList {
		bt := time.Unix(v.BeginTime, 0).UTC()
		et := time.Unix(v.EndTime, 0).UTC()
		ruleToAdd = append(ruleToAdd, rule.Rule{
			V0: body.SubjectCode,
			V1: body.SystemCode + "/" + v.ResourceIndex,
			V2: v.Action,
			V3: bt,
			V4: et,
		})
		if body.Inherit {
			ruleToAdd = append(ruleToAdd, rule.Rule{
				V0: body.SubjectCode,
				V1: body.SystemCode + "/" + v.ResourceIndex + "/*",
				V2: v.Action,
				V3: bt,
				V4: et,
			})
		}
	}

	err = rule.Add(ctx, ruleToAdd)
	if err != nil {
		logger.Errorf(ctx, "failed to add permission, err: %v", err)
		if errors.Is(err, rule.ErrDuplicateRule) {
			return output.Failure(ctx, controller.ErrSystemError.WithHint("User's permissions have been updated. Please refresh and try again."))
		}
		return output.Failure(ctx, controller.ErrSystemError)
	}
	return output.Success(ctx, nil)
}

func deleteItem(ctx echo.Context) error {
	body := struct {
		SystemCode     string       `json:"system_code" validate:"required,gt=0"`
		SubjectCode    string       `json:"subject_code" validate:"required,gt=0"`
		PermissionList []Permission `json:"permission_list" validate:"required,gt=0,dive,required"`
	}{}
	if err := input.BindAndValidate(ctx, &body); err != nil {
		return output.Failure(ctx, controller.ErrInvalidInput.WithMsg(err.Error()))
	}

	if ok, err := validateSystemAndSubject(ctx, body.SystemCode, body.SubjectCode); !ok {
		if err != nil {
			logger.Errorf(ctx, "failed to validate, err: %v, system code: %s, user code: %s", err, body.SystemCode, body.SubjectCode)
		}
		return output.Failure(ctx, controller.ErrSystemError)
	}

	tmpPermissionList, err := validatePermissionList(ctx, body.SystemCode, body.PermissionList)
	if err != nil {
		logger.Errorf(ctx, "failed to validate, err: %v", err)
		return output.Failure(ctx, controller.ErrSystemError)
	}

	ruleToDelete := make([]rule.Rule, 0, len(tmpPermissionList))

	for _, v := range tmpPermissionList {
		bt := time.Unix(v.BeginTime, 0).UTC()
		et := time.Unix(v.EndTime, 0).UTC()
		ruleToDelete = append(ruleToDelete, rule.Rule{
			V0: body.SubjectCode,
			V1: body.SystemCode + "/" + v.ResourceIndex,
			V2: v.Action,
			V3: bt,
			V4: et,
		})
	}

	err = rule.Delete(ctx, ruleToDelete)
	if err != nil {
		logger.Errorf(ctx, "failed to delete permission, err: %v", err)
		if errors.Is(err, rule.ErrRuleNotFound) {
			return output.Failure(ctx, controller.ErrSystemError.WithHint("User's permissions have been updated. Please refresh and try again."))
		}
		return output.Failure(ctx, controller.ErrSystemError)
	}
	return output.Success(ctx, nil)
}

func query(ctx echo.Context) error {
	body := struct {
		Page        int    `json:"page" validate:"required,gt=0"`
		PageSize    int    `json:"page_size" validate:"required,gt=0"`
		SystemCode  string `json:"system_code" validate:"required,gt=0"`
		SubjectCode string `json:"subject_code" validate:"required,gt=0"`
	}{}
	if err := input.BindAndValidate(ctx, &body); err != nil {
		return output.Failure(ctx, controller.ErrInvalidInput.WithMsg(err.Error()))
	}

	if ok, err := validateSystemAndSubject(ctx, body.SystemCode, body.SubjectCode); !ok {
		if err != nil {
			logger.Errorf(ctx, "failed to validate, err: %v, system code: %s, user code: %s", err, body.SystemCode, body.SubjectCode)
		}
		return output.Failure(ctx, controller.ErrSystemError)
	}

	enforcer, err := casbin.NewEnforcer(database.DB)
	if err != nil {
		logger.Errorf(ctx, "failed to create enforcer, err: %v", err)
		return output.Failure(ctx, controller.ErrSystemError)
	}
	ruleList, err := enforcer.GetImplicitPermissionsForUser(body.SubjectCode)
	if err != nil {
		logger.Errorf(ctx, "failed to get rule list, err: %v", err)
		return output.Failure(ctx, controller.ErrSystemError)
	}
	type Permission struct {
		FromCode      string `json:"from_code"`
		FromName      string `json:"from_name"`
		ResourceIndex string `json:"resource_index"`
		ResourceName  string `json:"resource_name"`
		Action        string `json:"action"`
		BeiginTime    string `json:"begin_time"`
		EndTime       string `json:"end_time"`
	}
	list := make([]Permission, 0, len(ruleList))
	systemCodeList := make([]string, 0, len(ruleList))
	recourceCodeList := make([]string, 0, len(ruleList))
	subjectCodeList := make([]string, 0, len(ruleList))
	for _, rule := range ruleList {
		for _, part := range strings.Split(rule[1], "/") {
			if strings.HasPrefix(part, define.PrefixSystem) {
				systemCodeList = append(systemCodeList, part)
			}
			if strings.HasPrefix(part, define.PrefixResource) {
				recourceCodeList = append(recourceCodeList, part)
			}
		}
		subjectCodeList = append(subjectCodeList, rule[0])
		list = append(list, Permission{
			FromCode:      rule[0],
			ResourceIndex: rule[1],
			Action:        rule[2],
			BeiginTime:    rule[3],
			EndTime:       rule[4],
		})
	}
	systemCodeList = util.Deduplicate(systemCodeList)
	recourceCodeList = util.Deduplicate(recourceCodeList)
	subjectCodeList = util.Deduplicate(subjectCodeList)
	systemList, err := dal.NewRepo[model.System]().QueryList(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Where("code IN ?", systemCodeList)
	})
	if err != nil {
		logger.Errorf(ctx, "failed to get system list, err: %v", err)
		return output.Failure(ctx, controller.ErrSystemError)
	}
	systemCode2System := util.ToMap(systemList, func(obj model.System) string {
		return obj.Code
	})
	resouceCode2Resouce, err := resource.QueryResourceByCode(ctx, body.SystemCode, recourceCodeList)
	if err != nil {
		logger.Errorf(ctx, "failed to get resource, err: %v", err)
		return output.Failure(ctx, controller.ErrSystemError)
	}
	subjectList, err := dal.NewRepo[model.Subject]().QueryList(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Where(model.Subject{SystemCode: body.SystemCode}).Where("code IN ?", subjectCodeList)
	})
	if err != nil {
		logger.Errorf(ctx, "failed to get subject list, err: %v", err)
		return output.Failure(ctx, controller.ErrSystemError)
	}
	subjectCode2Subject := util.ToMap(subjectList, func(obj model.Subject) string {
		return obj.Code
	})
	for i, v := range list {
		if subject, ok := subjectCode2Subject[v.FromCode]; ok {
			v.FromName = subject.Name
		}
		partList := strings.Split(v.ResourceIndex, "/")
		pathNameList := []string{}
		for i := 1; i < len(partList); i++ {
			part := partList[i]
			if system, ok := systemCode2System[part]; ok {
				pathNameList = append(pathNameList, system.Name)
			}
			if resource, ok := resouceCode2Resouce[part]; ok {
				pathNameList = append(pathNameList, resource.Name)
			}
			if part == "*" {
				pathNameList = append(pathNameList, "全部子集")
			}
		}
		v.ResourceName = strings.Join(pathNameList, "/")
		list[i] = v
	}
	return output.Success(ctx, map[string]interface{}{
		"list": list,
	})
}

func validatePermissionList(ctx echo.Context, systemCode string, permissionList []Permission) ([]Permission, error) {
	seen := make(map[string]struct{})
	filtered := make([]Permission, 0, len(permissionList))

	tmpResourceCodeList := make([]string, 0, len(permissionList))

	for _, v := range permissionList {
		trimmedResourceIndex := strings.TrimSpace(strings.Trim(strings.TrimSpace(v.ResourceIndex), "/"))
		v.ResourceIndex = trimmedResourceIndex

		key := fmt.Sprintf("%s:%s:%d:%d", v.ResourceIndex, v.Action, v.BeginTime, v.EndTime)

		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}

		filtered = append(filtered, v)

		for _, part := range strings.Split(trimmedResourceIndex, "/") {
			if strings.HasPrefix(part, define.PrefixResource) {
				tmpResourceCodeList = append(tmpResourceCodeList, part)
			}
		}
	}

	tmpResourceCodeList = util.Deduplicate(tmpResourceCodeList)

	resouceCode2Resouce, err := resource.QueryResourceByCode(ctx, systemCode, tmpResourceCodeList)
	if err != nil {
		return nil, fmt.Errorf("failed to query resource by code: %w", err)
	}

	for _, v := range filtered {
		partList := strings.Split(v.ResourceIndex, "/")

		allExist := true
		for _, path := range partList {
			if strings.HasPrefix(path, define.PrefixResource) {
				if _, exists := resouceCode2Resouce[path]; !exists {
					allExist = false
					break
				}
			}
		}

		if !allExist {
			return nil, fmt.Errorf("invalid resource in ResourceIndex: %s", v.ResourceIndex)
		}
	}

	return filtered, nil
}

func validateSystemAndSubject(ctx echo.Context, systemCode, subjectCode string) (bool, error) {
	if ok, err := system.Validate(ctx, systemCode); !ok {
		if err != nil {
			return false, fmt.Errorf("failed to validate system, err: %w", err)
		}
		return false, nil
	}

	if ok, err := subject.Validate(ctx, systemCode, subjectCode); !ok {
		if err != nil {
			return false, fmt.Errorf("failed to validate subject, err: %w", err)
		}
		return false, nil
	}

	return true, nil
}
