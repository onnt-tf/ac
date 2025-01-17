package dal

import (
	"ac/bootstrap/logger"
	"errors"
	"fmt"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type Entity interface {
	// model.UserChange | model.UserChangeRemind | model.Task | model.User | model.Revoke | model.Transfer | model.TaskV2 | model.SubtaskV2
}

type Repository[T Entity] interface {
	Insert(ctx echo.Context, db *gorm.DB, newValue *T) error
	BatchInsert(ctx echo.Context, db *gorm.DB, valuesToAdd []*T, batchSize int) error
	Update(ctx echo.Context, db *gorm.DB, newValue *T, funcs ...func(db *gorm.DB) *gorm.DB) error
	UpdateWithMap(ctx echo.Context, db *gorm.DB, newValue map[string]interface{}, funcs ...func(db *gorm.DB) *gorm.DB) error
	Query(ctx echo.Context, db *gorm.DB, funcs ...func(db *gorm.DB) *gorm.DB) (*T, error)
	QueryList(ctx echo.Context, db *gorm.DB, funcs ...func(db *gorm.DB) *gorm.DB) ([]T, error)
	Count(ctx echo.Context, db *gorm.DB, funcs ...func(db *gorm.DB) *gorm.DB) (int64, error)
}

var ErrMySQL = errors.New("MySQL error occurred")

type Repo[T Entity] struct{}

func NewRepo[T Entity]() *Repo[T] {
	return &Repo[T]{}
}

func logWithError(ctx echo.Context, operation string, err error) {
	if err == nil {
		return
	}
	logger.Errorf(ctx, "operation: %s, error: %s", operation, err)
}

func (r *Repo[T]) Insert(ctx echo.Context, db *gorm.DB, newValue *T) error {
	if newValue == nil {
		return fmt.Errorf("invalid argument: newValue is nil")
	}
	result := db.WithContext(ctx.Request().Context()).Create(newValue)
	if result.Error != nil {
		logWithError(ctx, "insert", result.Error)
		return errors.Join(ErrMySQL, fmt.Errorf("failed to insert record, err: %w", result.Error))
	}
	return nil
}

func (r *Repo[T]) BatchInsert(ctx echo.Context, db *gorm.DB, valuesToAdd []*T, batchSize int) error {
	if len(valuesToAdd) == 0 {
		return fmt.Errorf("invalid argument: valuesToAdd is empty")
	}
	for i, v := range valuesToAdd {
		if v == nil {
			return fmt.Errorf("invalid argument: value at index %d is nil", i)
		}
	}
	if batchSize <= 0 {
		batchSize = 10
	}
	result := db.WithContext(ctx.Request().Context()).CreateInBatches(valuesToAdd, batchSize)
	if result.Error != nil {
		logWithError(ctx, "batch insert", result.Error)
		return errors.Join(ErrMySQL, fmt.Errorf("failed to batch insert records, err: %w", result.Error))
	}
	return nil
}

func (r *Repo[T]) Update(ctx echo.Context, db *gorm.DB, newValue *T, funcs ...func(db *gorm.DB) *gorm.DB) error {
	if newValue == nil {
		return fmt.Errorf("invalid argument: newValue is nil")
	}
	result := db.WithContext(ctx.Request().Context()).Model(new(T)).Scopes(funcs...).Updates(newValue)
	if result.Error != nil {
		logWithError(ctx, "update", result.Error)
		return errors.Join(ErrMySQL, fmt.Errorf("failed to update record, err: %w", result.Error))
	}
	return nil
}

func (r *Repo[T]) UpdateWithMap(ctx echo.Context, db *gorm.DB, newValue map[string]interface{}, funcs ...func(db *gorm.DB) *gorm.DB) error {
	if newValue == nil {
		return fmt.Errorf("invalid argument: newValue is nil")
	}
	result := db.WithContext(ctx.Request().Context()).Model(new(T)).Scopes(funcs...).Updates(newValue)
	if result.Error != nil {
		logWithError(ctx, "update with map", result.Error)
		return errors.Join(ErrMySQL, fmt.Errorf("failed to update with map, err: %w", result.Error))
	}
	return nil
}

func (r *Repo[T]) Delete(ctx echo.Context, db *gorm.DB, funcs ...func(db *gorm.DB) *gorm.DB) error {
	result := db.WithContext(ctx.Request().Context()).Model(new(T)).Scopes(funcs...).Delete(new(T))
	if result.Error != nil {
		logWithError(ctx, "delete", result.Error)
		return errors.Join(ErrMySQL, fmt.Errorf("failed to delete record, err: %w", result.Error))
	}
	return nil
}

func (r *Repo[T]) Query(ctx echo.Context, db *gorm.DB, funcs ...func(db *gorm.DB) *gorm.DB) (*T, error) {
	var record T
	result := db.WithContext(ctx.Request().Context()).Scopes(funcs...).Limit(1).Find(&record)
	if result.Error != nil {
		logWithError(ctx, "query one", result.Error)
		return nil, errors.Join(ErrMySQL, fmt.Errorf("failed to query one record, err: %w", result.Error))
	}
	if result.RowsAffected == 0 {
		logger.Infof(ctx, "no matching record found during query one")
		return nil, nil
	}
	return &record, nil
}

func (r *Repo[T]) QueryList(ctx echo.Context, db *gorm.DB, funcs ...func(db *gorm.DB) *gorm.DB) ([]T, error) {
	var recordList []T
	result := db.WithContext(ctx.Request().Context()).Scopes(funcs...).Find(&recordList)
	if result.Error != nil {
		logWithError(ctx, "query list", result.Error)
		return nil, errors.Join(ErrMySQL, fmt.Errorf("failed to query list of records, err: %w", result.Error))
	}
	return recordList, nil
}

func (r *Repo[T]) Count(ctx echo.Context, db *gorm.DB, funcs ...func(db *gorm.DB) *gorm.DB) (int64, error) {
	var count int64
	result := db.WithContext(ctx.Request().Context()).Model(new(T)).Scopes(funcs...).Count(&count)
	if result.Error != nil {
		logWithError(ctx, "count", result.Error)
		return 0, errors.Join(ErrMySQL, fmt.Errorf("failed to count records, err: %w", result.Error))
	}
	return count, nil
}

func Paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
	const (
		defaultPageSize = 10
		maxPageSize     = 100
	)
	return func(db *gorm.DB) *gorm.DB {
		if page <= 0 {
			page = 1
		}
		if pageSize <= 0 {
			pageSize = defaultPageSize
		} else if pageSize > maxPageSize {
			pageSize = maxPageSize
		}
		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}
