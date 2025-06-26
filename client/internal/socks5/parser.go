package socks5

import (
	"fmt"
	"net"
)

func parseHandshake(req []byte) error {
	ver := req[0]
	nMethods := int(req[1])

	if ver != 0x05 {
		return fmt.Errorf("invalid version in handshake request")
	}
	if nMethods < 1 || len(req) < 2+nMethods {
		return fmt.Errorf("invalid nMethods in handshake request")
	}

	// ищем метод без аутентификации (0x00)
	hanNoAuth := false
	for i := 0; i < nMethods; i++ {
		if req[2+i] == 0x00 {
			hanNoAuth = true
			break
		}
	}

	if !hanNoAuth {
		return fmt.Errorf("no supported authentication methods")
	}

	return nil
}

func parseConnectRequest(req []byte) (string, error) {
	ver := req[0]
	cmd := req[1]
	atyp := req[3]

	if ver != 0x05 || cmd != 0x01 {
		return "", fmt.Errorf("invalid connection request")
	}

	var addr string
	var port uint16

	switch atyp {
	case 0x01:
		ip := net.IP(req[4:8])
		port = uint16(req[8])<<8 | uint16(req[9])
		addr = fmt.Sprintf("%s:%d", ip, port)

	case 0x03:
		nameLen := req[4]
		host := string(req[5 : 5+nameLen])
		port = uint16(req[5+nameLen])<<8 | uint16(req[6+nameLen])
		addr = fmt.Sprintf("%s:%d", host, port)

	default:
		return "", fmt.Errorf("invalid connection request")
	}

	return addr, nil
}
