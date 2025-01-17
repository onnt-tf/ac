package system

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
	"ac/service/system"
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

func RegisterRoutes(g *echo.Group) {
	g.POST("/add", addItem)
	g.POST("/update", updateItem)
	g.POST("/delete", deleteItem)
	g.GET("/query", query)
	g.GET("/get", GetItem)
}

func addItem(ctx echo.Context) error {
	body := struct {
		Name        string `json:"name" validate:"required,gt=0"`
		Description string `json:"description"`
	}{}
	if err := input.BindAndValidate(ctx, &body); err != nil {
		return output.Failure(ctx, controller.ErrInvalidInput.WithMsg(err.Error()))
	}
	var code string
	for i := 0; i < 3; i++ {
		tmpCode := util.GenerateCode(define.PrefixSystem)

		ok, err := system.IsCodeAvailable(ctx, tmpCode)
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
	newValue := &model.System{
		Name:        body.Name,
		Description: body.Description,
		Code:        code,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := dal.NewRepo[model.System]().Insert(ctx, database.DB, newValue); err != nil {
		logger.Errorf(ctx, "failed to insert record, err: %v", err)
		return output.Failure(ctx, controller.ErrSystemError)
	}
	return output.Success(ctx, System{ID: newValue.ID, Code: newValue.Code, Name: newValue.Name})
}

func updateItem(ctx echo.Context) error {
	body := struct {
		Code        string `json:"code" validate:"required,gt=0"`
		Name        string `json:"name" validate:"required,gt=0"`
		Description string `json:"description"`
	}{}
	if err := input.BindAndValidate(ctx, &body); err != nil {
		return output.Failure(ctx, controller.ErrInvalidInput.WithMsg(err.Error()))
	}

	if ok, err := system.Validate(ctx, body.Code); !ok {
		if err != nil {
			logger.Errorf(ctx, "failed to validate system, err: %v, code: %s", err, body.Code)
		}
		return output.Failure(ctx, controller.ErrSystemError.WithHint("Invalid system code"))
	}

	now := util.UTCNow()
	newValue := &model.System{
		Name:        body.Name,
		Description: body.Description,
		UpdatedAt:   now,
	}
	if err := dal.NewRepo[model.System]().Update(ctx, database.DB, newValue, func(db *gorm.DB) *gorm.DB {
		return db.Where(model.System{Code: body.Code}).Limit(1)
	}); err != nil {
		logger.Errorf(ctx, "failed to update record, err: %v", err)
		return output.Failure(ctx, controller.ErrSystemError)
	}
	return output.Success(ctx, nil)
}

func deleteItem(ctx echo.Context) error {
	body := struct {
		Code string `json:"code" validate:"required,gt=0"`
	}{}
	if err := input.BindAndValidate(ctx, &body); err != nil {
		return output.Failure(ctx, controller.ErrInvalidInput.WithMsg(err.Error()))
	}

	if ok, err := system.Validate(ctx, body.Code); !ok {
		if err != nil {
			logger.Errorf(ctx, "failed to validate system, err: %v, code: %s", err, body.Code)
		}
		return output.Failure(ctx, controller.ErrSystemError.WithHint("Invalid system code"))
	}

	now := util.UTCNow()
	newValue := &model.System{
		DeletedAt: &now,
	}
	if err := dal.NewRepo[model.System]().Update(ctx, database.DB, newValue, func(db *gorm.DB) *gorm.DB {
		return db.Where(model.System{Code: body.Code}).Limit(1)
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

	recordList, err := dal.NewRepo[model.System]().QueryList(ctx, database.DB, dal.Paginate(body.Page, body.PageSize))
	if err != nil {
		logger.Errorf(ctx, "failed to query, err: %v", err)
		return output.Failure(ctx, controller.ErrSystemError)
	}

	list := make([]System, 0, len(recordList))
	for _, v := range recordList {
		list = append(list, System{
			ID:         v.ID,
			Name:       v.Name,
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
		Code string `json:"code" validate:"required,gt=0"`
	}{}
	if err := input.BindAndValidate(ctx, &body); err != nil {
		return output.Failure(ctx, controller.ErrInvalidInput.WithMsg(err.Error()))
	}

	record, err := dal.NewRepo[model.System]().Query(ctx, database.DB, func(db *gorm.DB) *gorm.DB {
		return db.Where(model.System{Code: body.Code})
	})
	if err != nil {
		logger.Errorf(ctx, "failed to query, err: %v", err)
		return output.Failure(ctx, controller.ErrSystemError)
	} else if record == nil {
		logger.Infof(ctx, "failed to query, no matching record found, code: %s", body.Code)
		return output.Failure(ctx, controller.ErrRecordNotFound)
	}

	return output.Success(ctx, System{
		ID:          record.ID,
		Code:        record.Code,
		Name:        record.Name,
		Description: record.Description,
		ModifiedBy:  record.ModifiedBy,
		UpdatedAt:   record.UpdatedAt,
	})
}
