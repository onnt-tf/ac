package subject

import (
	"ac/model"

	"github.com/labstack/echo/v4"
)

func ValidateRole(ctx echo.Context, systemCode, code string) (bool, error) {
	return validate(ctx, systemCode, model.SubjectTypeRole, code)
}

func ValidateRoleBatch(ctx echo.Context, systemCode string, codeList []string) (map[string]bool, error) {
	return validateBatch(ctx, systemCode, model.SubjectTypeRole, codeList)
}

func IsRoleCodeAvailable(ctx echo.Context, code string) (bool, error) {
	return isCodeAvailable(ctx, model.SubjectTypeRole, code)
}
