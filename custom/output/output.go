package output

import (
	"ac/controller"
	"net/http"

	"github.com/labstack/echo/v4"
)

// Response defines the common structure for API responses.
type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Hint string      `json:"hint"`
	Data interface{} `json:"data"`
}

func Success(c echo.Context, data interface{}) error {
	if data == nil {
		data = struct{}{}
	}
	return c.JSON(http.StatusOK, Response{
		Code: 0,
		Msg:  "success",
		Data: data,
	})
}

func Failure(c echo.Context, err error) error {
	code := -1
	msg := ""
	hint := ""

	// Check if the error is of type *controller.Error
	if customErr, ok := err.(*controller.Error); ok {
		// Extract the code, message, and hint from the custom error
		code = customErr.Code
		msg = customErr.Msg
		hint = customErr.Hint
	} else if err != nil {
		// If it's not a *controller.Error, fallback to the generic error message
		msg = err.Error()
	}

	// Return a JSON response with the error details
	return c.JSON(http.StatusOK, Response{
		Code: code,
		Msg:  msg,
		Hint: hint,
		Data: struct{}{}, // Empty data structure
	})

}
