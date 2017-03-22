package main

import (
	"fmt"
)

const (
	ErrCodeOK                 = 1200
	ErrCodeBadRequest         = 1400
	ErrCodeInvalidParam       = 14001
	ErrCodeActionNotSupport   = 14003
	ErrCodeInvalidToken       = 14004
	ErrCodeUnauthorized       = 1401
	ErrCodeForbidden          = 1403
	ErrCodePermissionDenied   = 14030
	ErrCodeNotFound           = 1404
	ErrCodePlanNotFound       = 14040
	ErrCodeRegionNotFound     = 14041
	ErrCodeUserNotFound       = 14042
	ErrCodeMethodNotAllowed   = 1405
	ErrCodeTimeout            = 1408
	ErrCodeConflict           = 1409
	ErrCodeAdminNotPresented  = 15000
	ErrCodeServiceUnavailable = 1503

	ErrCodeUnknownError = 140010
)

var errText = map[int]string{
	ErrCodeOK:                 "OK",
	ErrCodeBadRequest:         "Bad request",
	ErrCodeInvalidParam:       "Invalid parameter",
	ErrCodeActionNotSupport:   "Not supported action",
	ErrCodeInvalidToken:       "Invalid token",
	ErrCodeUnauthorized:       "Unauthorized",
	ErrCodeForbidden:          "Forbidden",
	ErrCodePermissionDenied:   "Permission denied",
	ErrCodeNotFound:           "Not found",
	ErrCodePlanNotFound:       "No such plan",
	ErrCodeRegionNotFound:     "Region not exist",
	ErrCodeUserNotFound:       "User not exist",
	ErrCodeMethodNotAllowed:   "Method not allowed",
	ErrCodeTimeout:            "Request timeout",
	ErrCodeConflict:           "Already exist",
	ErrCodeAdminNotPresented:  "Admin not presented",
	ErrCodeServiceUnavailable: "Service unavailable",

	ErrCodeUnknownError: "Unknown error",
}

func ErrText(code int) string {
	return errText[code]
}

type ErrorMessage struct {
	Code    int
	Message string
}

func (e *ErrorMessage) Error() string {
	return fmt.Sprintf("[%v] %v", e.Code, e.Message)
}

func (e *ErrorMessage) New(code int) error {
	e.Code = code
	e.Message = ErrText(code)

	return e
}

func (e *ErrorMessage) NewMsg(code int, msg string) error {
	e.Code = code
	e.Message = msg

	return e
}

func ErrorNew(code int) error {
	var e ErrorMessage
	return e.New(code)
}

func ErrorNewMsg(code int, msg string) error {
	var e ErrorMessage
	return e.NewMsg(code, msg)
}
