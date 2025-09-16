package socks5

import "github.com/imightbuyaboat/SOCKS5-Proxy/client/internal/user"

type Storage interface {
	CheckUser(u *user.User) error
}
