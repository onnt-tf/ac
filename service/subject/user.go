package subject

import (
	"ac/model"

	"github.com/labstack/echo/v4"
)

func ValidateUser(ctx echo.Context, systemCode, code string) (bool, error) {
	return validate(ctx, systemCode, model.SubjectTypeUser, code)
}

func ValidateUserBatch(ctx echo.Context, systemCode string, codeList []string) (map[string]bool, error) {
	return validateBatch(ctx, systemCode, model.SubjectTypeUser, codeList)

}

func IsUserCodeAvailable(ctx echo.Context, code string) (bool, error) {
	return isCodeAvailable(ctx, model.SubjectTypeUser, code)
}
