package role

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

	"ac/service/subject"
	"ac/service/system"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type Role struct {
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

	var code string
	for i := 0; i < 3; i++ {
		tmpCode := util.GenerateCode(define.PrefixRole)

		ok, err := subject.IsRoleCodeAvailable(ctx, tmpCode)
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
	newValue := &model.Subject{
		SystemCode:  body.SystemCode,
		Type:        model.SubjectTypeRole,
		Name:        body.Name,
		Description: body.Description,
		Code:        code,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := dal.NewRepo[model.Subject]().Insert(ctx, database.DB, newValue); err != nil {
		logger.Errorf(ctx, "failed to insert record, err: %v", err)
		return output.Failure(ctx, controller.ErrSystemError)
	}
	return output.Success(ctx, Role{ID: newValue.ID, SystemCode: newValue.SystemCode, Code: newValue.Code})
}

func updateItem(ctx echo.Context) error {
	body := struct {
		SystemCode  string `json:"system_code" validate:"required,gt=0"`
		Code        string `json:"code" validate:"required,gt=0"`
		Name        string `json:"name" validate:"required,gt=0"`
		Description string `json:"description"`
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

	if ok, err := subject.ValidateRole(ctx, body.SystemCode, body.Code); !ok {
		if err != nil {
			logger.Errorf(ctx, "failed to validate role, err: %v, system code: %s, user code: %s", err, body.SystemCode, body.Code)
		}
		return output.Failure(ctx, controller.ErrSystemError.WithHint("Invalid role code"))
	}

	now := util.UTCNow()
	newValue := &model.Subject{
		Name:        body.Name,
		Description: body.Description,
		UpdatedAt:   now,
	}
	if err := dal.NewRepo[model.Subject]().Update(ctx, database.DB, newValue, func(db *gorm.DB) *gorm.DB {
		return db.Where(model.Subject{SystemCode: body.SystemCode, Code: body.Code, Type: model.SubjectTypeRole}).Limit(1)
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

	if ok, err := subject.ValidateRole(ctx, body.SystemCode, body.Code); !ok {
		if err != nil {
			logger.Errorf(ctx, "failed to validate role, err: %v, system code: %s, user code: %s", err, body.SystemCode, body.Code)
		}
		return output.Failure(ctx, controller.ErrSystemError.WithHint("Invalid role code"))
	}

	now := util.UTCNow()
	newValue := &model.Subject{
		DeletedAt: &now,
	}
	if err := dal.NewRepo[model.Subject]().Update(ctx, database.DB, newValue, func(db *gorm.DB) *gorm.DB {
		return db.Where(model.Subject{SystemCode: body.SystemCode, Code: body.Code, Type: model.SubjectTypeRole}).Limit(1)
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

	recordList, err := dal.NewRepo[model.Subject]().QueryList(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Where(model.Subject{Type: model.SubjectTypeRole})
	}, dal.Paginate(body.Page, body.PageSize))
	if err != nil {
		logger.Errorf(ctx, "failed to query, err: %v", err)
		return output.Failure(ctx, controller.ErrSystemError)
	}

	list := make([]Role, 0, len(recordList))
	for _, v := range recordList {
		list = append(list, Role{
			ID:         v.ID,
			Name:       v.Name,
			SystemCode: v.SystemCode,
			Code:       v.Code,
			ModifiedBy: v.ModifiedBy,
			UpdatedAt:  v.UpdatedAt,
		})
	}

	return output.Success(ctx, map[string]interface{}{
		"list": list,
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

	record, err := dal.NewRepo[model.Subject]().Query(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Where(model.Subject{SystemCode: body.SystemCode, Code: body.Code, Type: model.SubjectTypeRole})
	})
	if err != nil {
		logger.Errorf(ctx, "failed to query role, err: %v, system code: %s, code: %s", err, body.SystemCode, body.Code)
		return output.Failure(ctx, controller.ErrSystemError)
	}
	if record == nil {
		logger.Infof(ctx, "failed to query, no matching record found, system code: %s, code: %s", body.SystemCode, body.Code)
		return output.Failure(ctx, controller.ErrRecordNotFound)
	}

	return output.Success(ctx, Role{
		ID:          record.ID,
		Name:        record.Name,
		Description: record.Description,
		SystemCode:  record.SystemCode,
		Code:        record.Code,
		ModifiedBy:  record.ModifiedBy,
		UpdatedAt:   record.UpdatedAt,
	})
}
