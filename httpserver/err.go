package httpserver

import (
	"net/http"

	"github.com/pkg/errors"
)

type Code int

const (
	CodeOK Code = iota

	CodeValidateRuleFailed
	CodeInternalServerError
	CodeBadRequest
	CodeUnauthorized
	CodeNotFound
	CodeForbidden
	CodeTimeout

	CodeCreated
	CodeAccepted
	CodeNoContent
	CodeResetContent
	CodeAuthorizationFailed
)

var code2MessageM = map[Code]string{
	CodeOK:                  "OK",
	CodeInternalServerError: "Internal Server Error",
	CodeBadRequest:          "Bad Request",
	CodeUnauthorized:        "Unauthorized",
	CodeNotFound:            "Not Found",
	CodeForbidden:           "Forbidden",
	CodeTimeout:             "Timeout",
	CodeCreated:             "Created",
	CodeAccepted:            "Accepted",
	CodeNoContent:           "No Content",
	CodeResetContent:        "Reset Content",
	CodeValidateRuleFailed:  "Validate Rule Failed",
}

var code2StatusM = map[Code]int{
	CodeOK:                  http.StatusOK,
	CodeValidateRuleFailed:  http.StatusBadRequest,
	CodeInternalServerError: http.StatusInternalServerError,
	CodeBadRequest:          http.StatusBadRequest,
	CodeUnauthorized:        http.StatusUnauthorized,
	CodeNotFound:            http.StatusNotFound,
	CodeForbidden:           http.StatusForbidden,
	CodeTimeout:             http.StatusGatewayTimeout,
	CodeCreated:             http.StatusCreated,
	CodeAccepted:            http.StatusAccepted,
	CodeNoContent:           http.StatusNoContent,
	CodeResetContent:        http.StatusResetContent,
	CodeAuthorizationFailed: http.StatusUnauthorized,
}

func statusOfCode(code Code) int {
	if status, ok := code2StatusM[code]; ok {
		return status
	}
	return http.StatusInternalServerError
}

type Err struct {
	Status int
	Code   Code
	Err    error
}

var _ error = &Err{}

func (e *Err) Error() string {
	return e.Err.Error()
}

func (e *Err) WithStatus(status int) *Err {
	e.Status = status
	return e
}

func ErrorWithCode(code Code) *Err {
	msg, ok := code2MessageM[code]
	if !ok {
		msg = code2MessageM[CodeInternalServerError]
	}
	return &Err{
		Status: statusOfCode(code),
		Code:   code,
		Err:    errors.New(msg),
	}
}

func NewError(code Code, msg string) *Err {
	return &Err{
		Status: statusOfCode(code),
		Code:   code,
		Err:    errors.New(msg),
	}
}

func ErrorWithBadRequest() *Err {
	return ErrorWithCode(CodeBadRequest)
}

func ErrorWithInternalServer() *Err {
	return ErrorWithCode(CodeInternalServerError)
}

func ErrWithUnAuthorized() *Err {
	return ErrorWithCode(CodeUnauthorized)
}

func ErrorWithAuthorizationFailed(reason string) *Err {
	return &Err{
		Status: http.StatusUnauthorized,
		Code:   CodeAuthorizationFailed,
		Err:    errors.New(reason),
	}
}
