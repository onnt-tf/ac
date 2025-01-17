package input

import (
	"github.com/labstack/echo/v4"
)

func BindAndValidate(c echo.Context, input interface{}) error {
	if err := c.Bind(input); err != nil {
		return err
	}
	if err := c.Validate(input); err != nil {
		return err
	}
	return nil
}
