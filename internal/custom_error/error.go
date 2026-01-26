package custom_error

import (
	"errors"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrIncorrectPassword = errors.New("incorrect password")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidAdminToken = errors.New("admin token invalid")
	ErrInvalidLogin      = errors.New("invalid login")
	ErrInvalidPassword   = errors.New("invalid password")

	ErrDocumentNotFound = errors.New("saga not found")
)
