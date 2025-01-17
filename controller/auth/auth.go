package auth

import (
	"ac/bootstrap/database"
	"ac/bootstrap/logger"
	"ac/controller"
	"ac/custom/define"
	"ac/custom/input"
	"ac/custom/output"
	"ac/service/casbin"
	"ac/service/resource"
	"ac/service/subject"
	"strings"

	"github.com/labstack/echo/v4"
)

func RegisterRoutes(g *echo.Group) {
	g.POST("/authenticate", authenticate)
}

func authenticate(ctx echo.Context) error {
	body := struct {
		SystemCode    string `json:"system_code" validate:"required,gt=0"`
		UserCode      string `json:"user_code" validate:"required,gt=0"`
		ResourceIndex string `json:"resource_index" validate:"required,gt=0"`
		Action        string `json:"action" validate:"required,gt=0"`
	}{}
	if err := input.BindAndValidate(ctx, &body); err != nil {
		return output.Failure(ctx, controller.ErrInvalidInput.WithMsg(err.Error()))
	}
	if _, ok := define.ValidAction2Level[body.Action]; !ok {
		return output.Failure(ctx, controller.ErrSystemError.WithHint("Invalid action"))
	}
	if ok, err := subject.ValidateUser(ctx, body.SystemCode, body.UserCode); !ok {
		if err != nil {
			logger.Errorf(ctx, "failed to validate user, err: %v, system code: %s, code: %s", err, body.SystemCode, body.UserCode)
		}
		return output.Failure(ctx, controller.ErrSystemError.WithHint("Invalid resource code"))
	}
	partList := strings.Split(body.ResourceIndex, "/")

	resourceCodeList := make([]string, 0, len(partList)-1)
	for _, v := range partList {
		if !strings.HasPrefix(v, define.PrefixResource) {
			continue
		}
		resourceCodeList = append(resourceCodeList, v)
	}

	validateResult, err := resource.ValidateBatch(ctx, body.SystemCode, resourceCodeList)
	if err != nil {
		logger.Errorf(ctx, "failed to validate role, err: %v", err)
		return output.Failure(ctx, controller.ErrSystemError)
	}

	for _, v := range resourceCodeList {
		if valid, ok := validateResult[v]; ok && valid {
			continue
		}
		return output.Failure(ctx, controller.ErrSystemError.WithHint("Invalid resource index"))
	}

	enforcer, err := casbin.NewEnforcer(database.DB)
	if err != nil {
		return output.Failure(ctx, controller.ErrSystemError)
	}
	authorized, err := enforcer.Enforce(body.UserCode, body.SystemCode+body.ResourceIndex, body.Action)
	if err != nil {
		logger.Errorf(ctx, "failed to enforce, err: %v, system code: %s, user code: %s, resource code: %s", err, body.SystemCode, body.UserCode, body.ResourceIndex)
		return output.Failure(ctx, controller.ErrSystemError)
	}
	return output.Success(ctx, map[string]bool{
		"authorized": authorized,
	})
}
