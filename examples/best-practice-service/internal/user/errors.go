package user

import "errors"

var (
	ErrEmailExists  = errors.New("email already exists")
	ErrInvalidEmail = errors.New("email is required")
	ErrInvalidName  = errors.New("name is required")
)
