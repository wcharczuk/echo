package web

import (
	"fmt"

	"github.com/blend/go-sdk/exception"
)

const (
	// ErrSessionIDEmpty is thrown if a session id is empty.
	ErrSessionIDEmpty exception.Class = "auth session id is empty"
	// ErrSessionIDTooLong is thrown if a session id is too long.
	ErrSessionIDTooLong exception.Class = "auth session id is too long"
	// ErrSecureSessionIDEmpty is an error that is thrown if a given secure session id is invalid.
	ErrSecureSessionIDEmpty exception.Class = "auth secure session id is empty"
	// ErrSecureSessionIDTooLong is an error that is thrown if a given secure session id is invalid.
	ErrSecureSessionIDTooLong exception.Class = "auth secure session id is too long"
	// ErrSecureSessionIDInvalid is an error that is thrown if a given secure session id is invalid.
	ErrSecureSessionIDInvalid exception.Class = "auth secure session id is invalid"
	// ErrUnsetViewTemplate is an error that is thrown if a given secure session id is invalid.
	ErrUnsetViewTemplate exception.Class = "view result template is unset"

	// ErrParameterMissing is an error on request validation.
	ErrParameterMissing exception.Class = "parameter is missing"
)

func newParameterMissingError(paramName string) error {
	return fmt.Errorf("`%s` parameter is missing", paramName)
}

// IsErrSessionInvalid returns if an error is a session invalid error.
func IsErrSessionInvalid(err error) bool {
	if err == nil {
		return false
	}
	switch err {
	case ErrSessionIDEmpty,
		ErrSessionIDTooLong,
		ErrSecureSessionIDEmpty,
		ErrSecureSessionIDTooLong,
		ErrSecureSessionIDInvalid:
		return true
	default:
		return false
	}
}
