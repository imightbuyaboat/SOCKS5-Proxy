package postgres

import "errors"

var (
	ErrUserNotExists     = errors.New("user not exists")
	ErrIncorrectPassword = errors.New("incorrect password")
)
