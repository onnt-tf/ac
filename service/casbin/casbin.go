package casbin

import (
	"ac/custom/define"
	"fmt"
	"time"

	casebinV2 "github.com/casbin/casbin/v2"
	casebinModel "github.com/casbin/casbin/v2/model"
	gormAdapterV3 "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"
)

func NewEnforcer(db *gorm.DB) (*casebinV2.Enforcer, error) {
	adapter, err := gormAdapterV3.NewAdapterByDB(db)
	if err != nil {
		return nil, fmt.Errorf("failed to create adapter, err: %w", err)
	}

	// 定义 Casbin 模型
	modelText := `
		[request_definition]
			r = user, resource, action
			
		[policy_definition]
			p = subject, resource, action, begin_time, end_time
			
		[role_definition]
			g = _, _
			
		[policy_effect]
			e = some(where (p.eft == allow))
			
		[matchers]
			m = (r.user == p.subject || g(r.user, p.subject)) \
				&& keyMatch(r.resource, p.resource) \
				&& actionMatch(r.action, p.action) \
				&& timeMatch(p.begin_time, p.end_time)
	`

	model, err := casebinModel.NewModelFromString(modelText)
	if err != nil {
		return nil, fmt.Errorf("failed to create model, err: %w", err)
	}

	enforcer, err := casebinV2.NewEnforcer(model, adapter)
	if err != nil {
		return nil, fmt.Errorf("failed to create enforcer, err: %w", err)
	}

	enforcer.AddFunction("actionMatch", actionMatch)
	enforcer.AddFunction("timeMatch", timeMatch)

	if err := enforcer.LoadPolicy(); err != nil {
		return nil, fmt.Errorf("failed to load policy, err: %w", err)
	}

	return enforcer, nil
}

func timeMatch(args ...interface{}) (interface{}, error) {
	if len(args) < 2 {
		return false, fmt.Errorf("insufficient arguments: expected begin_time and end_time")
	}

	now := time.Now()
	layout := time.RFC3339

	parseTime := func(arg interface{}) (time.Time, error) {
		str, ok := arg.(string)
		if !ok {
			return time.Time{}, fmt.Errorf("invalid type for time argument: expected string, got %T", arg)
		}
		parsedTime, err := time.Parse(layout, str)
		if err != nil {
			return time.Time{}, fmt.Errorf("failed to parse time, err: %w", err)
		}
		return parsedTime, nil
	}

	beginTime, err := parseTime(args[0])
	if err != nil {
		return false, fmt.Errorf("failed to parse begin_time, err: %w", err)
	}
	endTime, err := parseTime(args[1])
	if err != nil {
		return false, fmt.Errorf("failed to parse end_time, err: %w", err)
	}

	if now.Equal(beginTime) || now.Equal(endTime) || (now.After(beginTime) && now.Before(endTime)) {
		return true, nil
	}

	return false, nil
}

func actionMatch(args ...interface{}) (interface{}, error) {
	if len(args) < 2 {
		return false, fmt.Errorf("insufficient arguments: expected requestedAction and policyAction")
	}

	requestedAction, ok1 := args[0].(string)
	policyAction, ok2 := args[1].(string)

	if !ok1 || !ok2 {
		return false, fmt.Errorf("invalid argument types: expected strings for requestedAction and policyAction")
	}

	requestedLevel, requestedExists := define.ValidAction2Level[requestedAction]
	policyLevel, policyExists := define.ValidAction2Level[policyAction]

	if !requestedExists {
		return false, fmt.Errorf("invalid requestedAction: %s", requestedAction)
	}
	if !policyExists {
		return false, fmt.Errorf("invalid policyAction: %s", policyAction)
	}

	return policyLevel >= requestedLevel, nil
}
