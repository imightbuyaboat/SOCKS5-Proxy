package models

import (
	"context"
)

type User struct {
	Username string
	Password string
}

type Storage interface {
	CheckUser(ctx context.Context, u *User) error
}
