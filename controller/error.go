package controller

// Predefined errors
var (
	ErrSystemError    = NewError(1, "system error", "An unexpected system error occurred")
	ErrInvalidInput   = NewError(2, "invalid input", "Please check your input")
	ErrRecordNotFound = NewError(3, "record not found", "The record does not exist or is no longer available")
)

// Error struct defines the structure of an error
type Error struct {
	Code int    `json:"code"` // Error code, identifies the type of error
	Msg  string `json:"msg"`  // Error message, used for debugging and logging
	Hint string `json:"hint"` // Hint message, provides user-friendly information
}

// Error implements the error interface, returning the error message
func (e *Error) Error() string {
	return e.Msg
}

// NewError is a constructor function that creates a new Error instance
func NewError(code int, msg, hint string) *Error {
	return &Error{
		Code: code,
		Msg:  msg,
		Hint: hint,
	}
}

// WithMsg returns a new Error instance with a temporarily modified Msg field
func (e *Error) WithMsg(newMsg string) *Error {
	return &Error{
		Code: e.Code,
		Msg:  newMsg,
		Hint: e.Hint,
	}
}

// WithHint returns a new Error instance with a temporarily modified Hint field
func (e *Error) WithHint(newHint string) *Error {
	return &Error{
		Code: e.Code,
		Msg:  e.Msg,
		Hint: newHint,
	}
}
