package errors

import "errors"

const (
	ErrSomethingWentWrong = "something went wrong"
	ErrUnauthorized       = "unauthorized"
	ErrInvalidJson        = "Invalid JSON"
	ErrMissingBody        = "missing body request"
	ErrInvalidID          = "ID is not in its proper form"
)

var (
	ErrPasswordIncorrect = errors.New("invalid credentials")
	ErrInvalidToken      = errors.New("invalid token")
)
