package rule

import (
	"ac/bootstrap/database"
	"ac/custom/define"
	"ac/custom/util"
	"ac/dal"
	"ac/model"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

var ErrDuplicateRule = errors.New("rule already exists")
var ErrRuleNotFound = errors.New("rule not found")

type Rule struct {
	PType string    `json:"p_type"`
	V0    string    `json:"v0"`
	V1    string    `json:"v1"`
	V2    string    `json:"v2"`
	V3    time.Time `json:"v3"`
	V4    time.Time `json:"v4"`
}

func (r *Rule) validate() error {
	if strings.TrimSpace(r.V0) == "" {
		return errors.New("v0 is empty")
	}
	if strings.TrimSpace(r.V1) == "" {
		return errors.New("v1 is empty")
	}

	if _, ok := define.ValidAction2Level[r.V2]; !ok {
		return errors.New("invalid v2")
	}

	if r.PType == model.PTypePolicy {
		if r.V3.IsZero() {
			return errors.New("v3 is zero")
		}
		if r.V4.IsZero() {
			return errors.New("v4 is zero")
		}
		if r.V4.Before(r.V3) {
			return errors.New("v4 must be after v3")
		}
	}

	return nil
}

func Add(ctx echo.Context, ruleList []Rule) error {
	now := util.UTCNow()
	ruleListToAdd := make([]*model.CasbinRule, 0, len(ruleList))
	for _, v := range ruleList {
		if err := v.validate(); err != nil {
			return fmt.Errorf("rule is invalid , err: %w", err)
		}
		if v.V3.Before(now) && v.V4.Before(now) {
			return errors.New("rule has expired")
		}
		ruleListToAdd = append(ruleListToAdd, &model.CasbinRule{
			PType: v.PType,
			V0:    v.V0,
			V1:    v.V1,
			V2:    v.V2,
			V3:    v.V3.Format(time.RFC3339),
			V4:    v.V4.Format(time.RFC3339),
		})
	}
	logContent, err := sonic.MarshalString(ruleListToAdd)
	if err != nil {
		return fmt.Errorf("failed to marshal rule list, err: %w", err)
	}

	err = database.DB.WithContext(ctx.Request().Context()).Transaction(func(tx *gorm.DB) error {
		log := &model.CasbinRuleLog{
			Operate:   model.OperateAdd,
			Content:   logContent,
			CreatedAt: now,
		}
		err = dal.NewRepo[model.CasbinRuleLog]().Insert(ctx, tx, log)
		if err != nil {
			return fmt.Errorf("failed to add log, err: %w", err)
		}

		for _, v := range ruleListToAdd {
			rerourd, err := dal.NewRepo[model.CasbinRule]().Query(ctx, tx, func(db *gorm.DB) *gorm.DB {
				return db.Where(&model.CasbinRule{
					PType: v.PType,
					V0:    v.V0,
					V1:    v.V1,
				})
			})
			if err != nil {
				return fmt.Errorf("failed to query rule, err: %w", err)
			}
			if rerourd != nil {
				return ErrDuplicateRule
			}
			err = dal.NewRepo[model.CasbinRule]().Insert(ctx, tx, v)
			if err != nil {
				return fmt.Errorf("failed to add rule, err: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to commit rule, err: %w", err)
	}
	return nil
}

func Delete(ctx echo.Context, ruleList []Rule) error {
	ruleListToDelete := make([]*model.CasbinRule, 0, len(ruleList))
	for _, v := range ruleList {
		if err := v.validate(); err != nil {
			return fmt.Errorf("rule is invalid , err: %w", err)
		}
		ruleListToDelete = append(ruleListToDelete, &model.CasbinRule{
			PType: v.PType,
			V0:    v.V0,
			V1:    v.V1,
			V2:    v.V2,
			V3:    v.V3.Format(time.RFC3339),
			V4:    v.V4.Format(time.RFC3339),
		})
	}
	logContent, err := sonic.MarshalString(ruleListToDelete)
	if err != nil {
		return fmt.Errorf("failed to marshal rule list, err: %w", err)
	}
	now := util.UTCNow()
	err = database.DB.WithContext(ctx.Request().Context()).Transaction(func(tx *gorm.DB) error {
		log := &model.CasbinRuleLog{
			Operate:   model.OperateDelete,
			Content:   logContent,
			CreatedAt: now,
		}
		err = dal.NewRepo[model.CasbinRuleLog]().Insert(ctx, tx, log)
		if err != nil {
			return fmt.Errorf("failed to add log, err: %w", err)
		}

		for _, v := range ruleListToDelete {
			record, err := dal.NewRepo[model.CasbinRule]().Query(ctx, tx, func(db *gorm.DB) *gorm.DB {
				return db.Where(v)
			})
			if err != nil {
				return fmt.Errorf("failed to query rule, err: %w", err)
			}
			if record == nil {
				return ErrRuleNotFound
			}
			err = dal.NewRepo[model.CasbinRule]().Delete(ctx, tx, func(db *gorm.DB) *gorm.DB {
				return db.Where(record).Limit(1)
			})
			if err != nil {
				return fmt.Errorf("failed to delete rule, err: %w", err)
			}
		}

		deletedRuleList := make([]*model.CasbinRuleDeleted, 0, len(ruleListToDelete))
		for _, v := range ruleListToDelete {
			deletedRuleList = append(deletedRuleList, &model.CasbinRuleDeleted{
				LogID:     log.ID,
				PType:     v.PType,
				V0:        v.V0,
				V1:        v.V1,
				V2:        v.V2,
				V3:        v.V3,
				V4:        v.V4,
				CreatedAt: now,
			})
		}
		err = dal.NewRepo[model.CasbinRuleDeleted]().BatchInsert(ctx, tx, deletedRuleList, 20)
		if err != nil {
			return fmt.Errorf("failed to add deleted rule, err: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to commit rule, err: %w", err)
	}
	return nil
}

func Set(ctx echo.Context, ruleList []Rule) error {
	ruleListToSet := make([]*model.CasbinRule, 0, len(ruleList))
	now := util.UTCNow()
	for _, v := range ruleList {
		if err := v.validate(); err != nil {
			return fmt.Errorf("invalid rule: %w", err)
		}
		ruleListToSet = append(ruleListToSet, &model.CasbinRule{
			PType: v.PType,
			V0:    v.V0,
			V1:    v.V1,
			V2:    v.V2,
			V3:    v.V3.Format(time.RFC3339),
			V4:    v.V4.Format(time.RFC3339),
		})
	}

	logContent, err := sonic.MarshalString(ruleListToSet)
	if err != nil {
		return fmt.Errorf("failed to marshal rules: %w", err)
	}

	err = database.DB.WithContext(ctx.Request().Context()).Transaction(func(tx *gorm.DB) error {
		log := &model.CasbinRuleLog{
			Operate:   model.OperateSet,
			Content:   logContent,
			CreatedAt: now,
		}
		err = dal.NewRepo[model.CasbinRuleLog]().Insert(ctx, tx, log)
		if err != nil {
			return fmt.Errorf("failed to log operation: %w", err)
		}

		deletedRuleList := make([]*model.CasbinRuleDeleted, 0, len(ruleListToSet))

		for _, v := range ruleListToSet {
			condition := &model.CasbinRule{
				PType: v.PType,
				V0:    v.V0,
				V1:    v.V1,
			}
			record, err := dal.NewRepo[model.CasbinRule]().Query(ctx, tx, func(db *gorm.DB) *gorm.DB {
				return db.Where(condition)
			})
			if err != nil {
				return fmt.Errorf("failed to query rule, err: %w", err)
			}
			if record == nil {
				err = dal.NewRepo[model.CasbinRule]().Insert(ctx, tx, v)
				if err != nil {
					return fmt.Errorf("failed to add rule, err: %w", err)
				}
			} else {
				err = dal.NewRepo[model.CasbinRule]().Update(ctx, tx, &model.CasbinRule{
					V2: v.V2,
					V3: v.V3,
					V4: v.V4,
				}, func(db *gorm.DB) *gorm.DB {
					return db.Where(condition).Limit(1)
				})
				if err != nil {
					return fmt.Errorf("failed to update rule, err: %w", err)
				}
				deletedRuleList = append(deletedRuleList, &model.CasbinRuleDeleted{
					LogID:     record.ID,
					PType:     record.PType,
					V0:        record.V0,
					V1:        record.V1,
					V2:        record.V2,
					V3:        record.V3,
					V4:        record.V4,
					CreatedAt: now,
				})
			}
		}

		if len(deletedRuleList) > 0 {
			err = dal.NewRepo[model.CasbinRuleDeleted]().BatchInsert(ctx, tx, deletedRuleList, 20)
			if err != nil {
				return fmt.Errorf("failed to add deleted rule, err: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to set policies: %w", err)
	}

	return nil
}
