package resource

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
	"ac/service/resource"

	"ac/service/system"
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

func RegisterRoutes(g *echo.Group) {
	g.POST("/add", addItem)
	g.POST("/update", updateItem)
	g.POST("/delete", deleteItem)
	g.GET("/query", query)
	g.GET("/get", GetItem)
}

func addItem(ctx echo.Context) error {
	body := struct {
		SystemCode  string `json:"system_code" validate:"required,gt=0"`
		Name        string `json:"name" validate:"required,gt=0"`
		Description string `json:"description"`
		ParentCode  string `json:"parent_code"`
	}{}
	if err := input.BindAndValidate(ctx, &body); err != nil {
		return output.Failure(ctx, controller.ErrInvalidInput.WithMsg(err.Error()))
	}

	if ok, err := system.Validate(ctx, body.SystemCode); !ok {
		if err != nil {
			logger.Errorf(ctx, "failed to validate system, err: %v, code: %s", err, body.SystemCode)
		}
		return output.Failure(ctx, controller.ErrSystemError.WithHint("Invalid system code"))
	}

	if body.ParentCode != "" {
		if ok, err := resource.Validate(ctx, body.SystemCode, body.ParentCode); !ok {
			if err != nil {
				logger.Errorf(ctx, "failed to validate resource, err: %v, system code: %s, code: %s", err, body.SystemCode, body.ParentCode)
			}
			return output.Failure(ctx, controller.ErrSystemError.WithHint("Invalid resource code"))
		}
	}

	var code string
	for i := 0; i < 3; i++ {
		tmpCode := util.GenerateCode(define.PrefixResource)

		ok, err := resource.IsCodeAvailable(ctx, tmpCode)
		if err != nil {
			logger.Errorf(ctx, "failed to check code availability, err: %v, code: %s", err, code)
			return output.Failure(ctx, controller.ErrSystemError)
		}
		if ok {
			code = tmpCode
			break
		}
	}

	if code == "" {
		logger.Errorf(ctx, "failed to generate unique code after 3 attempts")
		return output.Failure(ctx, controller.ErrSystemError.WithHint("Unable to generate the code. Please try again later"))
	}

	now := util.UTCNow()
	newValue := &model.Resource{
		SystemCode:  body.SystemCode,
		Name:        body.Name,
		Code:        code,
		ParentCode:  body.ParentCode,
		Description: body.Description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := dal.NewRepo[model.Resource]().Insert(ctx, database.DB, newValue); err != nil {
		logger.Errorf(ctx, "failed to insert record, err: %v", err)
		return output.Failure(ctx, controller.ErrSystemError)
	}

	return output.Success(ctx, Resource{
		ID:          newValue.ID,
		SystemCode:  newValue.SystemCode,
		Code:        newValue.Code,
		Name:        newValue.Name,
		Description: newValue.Description,
		ModifiedBy:  newValue.ModifiedBy,
		UpdatedAt:   newValue.UpdatedAt,
	})
}

func updateItem(ctx echo.Context) error {
	body := struct {
		SystemCode  string `json:"system_code" validate:"required,gt=0"`
		Code        string `json:"code" validate:"required,gt=0"`
		Name        string `json:"name" validate:"required,gt=0"`
		Description string `json:"description"`
		ParentCode  string `json:"parent_code"`
	}{}
	if err := input.BindAndValidate(ctx, &body); err != nil {
		return output.Failure(ctx, controller.ErrInvalidInput.WithMsg(err.Error()))
	}

	if ok, err := system.Validate(ctx, body.SystemCode); !ok {
		if err != nil {
			logger.Errorf(ctx, "failed to validate system, err: %v, code: %s", err, body.SystemCode)
		}
		return output.Failure(ctx, controller.ErrSystemError.WithHint("Invalid system code"))
	}

	if ok, err := resource.Validate(ctx, body.SystemCode, body.Code); !ok {
		if err != nil {
			logger.Errorf(ctx, "failed to validate resource, err: %v, system code: %s, code: %s", err, body.SystemCode, body.ParentCode)
		}
		return output.Failure(ctx, controller.ErrSystemError.WithHint("Invalid resource code"))
	}

	if body.ParentCode != "" {
		if ok, err := resource.Validate(ctx, body.SystemCode, body.ParentCode); !ok {
			if err != nil {
				logger.Errorf(ctx, "failed to validate resource, err: %v, system code: %s, code: %s", err, body.SystemCode, body.ParentCode)
			}
			return output.Failure(ctx, controller.ErrSystemError.WithHint("Invalid resource code"))
		}
	}

	now := util.UTCNow()
	newValue := &model.Resource{
		Name:        body.Name,
		Description: body.Description,
		ParentCode:  body.Code,
		UpdatedAt:   now,
	}
	if err := dal.NewRepo[model.Resource]().Update(ctx, database.DB, newValue, func(db *gorm.DB) *gorm.DB {
		return db.Where(model.Resource{SystemCode: body.SystemCode, Code: body.Code}).Limit(1)
	}); err != nil {
		logger.Errorf(ctx, "failed to update record, err: %v", err)
		return output.Failure(ctx, controller.ErrSystemError)
	}
	return output.Success(ctx, nil)
}

func deleteItem(ctx echo.Context) error {
	body := struct {
		SystemCode string `json:"system_code" validate:"required,gt=0"`
		Code       string `json:"code" validate:"required,gt=0"`
	}{}
	if err := input.BindAndValidate(ctx, &body); err != nil {
		return output.Failure(ctx, controller.ErrInvalidInput.WithMsg(err.Error()))
	}

	if ok, err := system.Validate(ctx, body.SystemCode); !ok {
		if err != nil {
			logger.Errorf(ctx, "failed to validate system, err: %v, code: %s", err, body.SystemCode)
		}
		return output.Failure(ctx, controller.ErrSystemError.WithHint("Invalid system code"))
	}

	if ok, err := resource.Validate(ctx, body.SystemCode, body.Code); !ok {
		if err != nil {
			logger.Errorf(ctx, "failed to validate resource, err: %v, system code: %s, code: %s", err, body.SystemCode, body.Code)
		}
		return output.Failure(ctx, controller.ErrSystemError.WithHint("Invalid resource code"))
	}

	now := util.UTCNow()
	newValue := &model.Resource{
		DeletedAt: &now,
	}
	if err := dal.NewRepo[model.Resource]().Update(ctx, database.DB, newValue, func(db *gorm.DB) *gorm.DB {
		return db.Where(model.Resource{SystemCode: body.SystemCode, Code: body.Code}).Limit(1)
	}); err != nil {
		logger.Errorf(ctx, "failed to update record, err: %v", err)
		return output.Failure(ctx, controller.ErrSystemError)
	}
	return output.Success(ctx, nil)
}

func query(ctx echo.Context) error {
	body := struct {
		Page     int `json:"page"`
		PageSize int `json:"page_size"`
	}{}
	if err := input.BindAndValidate(ctx, &body); err != nil {
		return output.Failure(ctx, controller.ErrInvalidInput.WithMsg(err.Error()))
	}

	recordList, err := dal.NewRepo[model.Resource]().QueryList(ctx, database.DB, dal.Paginate(body.Page, body.PageSize))
	if err != nil {
		logger.Errorf(ctx, "failed to query, err: %v", err)
		return output.Failure(ctx, controller.ErrSystemError)
	}

	count, err := dal.NewRepo[model.Resource]().Count(ctx, database.DB)
	if err != nil {
		logger.Errorf(ctx, "failed to count, err: %v", err)
		return output.Failure(ctx, controller.ErrSystemError)
	}
	list := make([]Resource, 0, len(recordList))
	for _, v := range recordList {
		list = append(list, Resource{
			ID:         v.ID,
			Name:       v.Name,
			SystemCode: v.SystemCode,
			Code:       v.Code,
			ModifiedBy: v.ModifiedBy,
			UpdatedAt:  v.UpdatedAt,
		})
	}

	return output.Success(ctx, map[string]interface{}{
		"total": count,
		"list":  list,
	})
}

func GetItem(ctx echo.Context) error {
	body := struct {
		SystemCode string `json:"system_code" validate:"required,gt=0"`
		Code       string `json:"code" validate:"required,gt=0"`
	}{}
	if err := input.BindAndValidate(ctx, &body); err != nil {
		return output.Failure(ctx, controller.ErrInvalidInput.WithMsg(err.Error()))
	}

	record, err := dal.NewRepo[model.Resource]().Query(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Where(model.Resource{SystemCode: body.SystemCode, Code: body.Code}).Limit(1)
	})
	if err != nil {
		logger.Errorf(ctx, "failed to query, err: %v", err)
		return output.Failure(ctx, controller.ErrSystemError)
	} else if record == nil {
		logger.Infof(ctx, "failed to query, no matching record found, code: %s", body.Code)
		return output.Failure(ctx, controller.ErrRecordNotFound)
	}

	return output.Success(ctx, Resource{
		ID:          record.ID,
		Name:        record.Name,
		Description: record.Description,
		SystemCode:  record.SystemCode,
		Code:        record.Code,
		ModifiedBy:  record.ModifiedBy,
		UpdatedAt:   record.UpdatedAt,
	})
}
