package parser

import (
	"fmt"
	"net"

	"github.com/imightbuyaboat/SOCKS5-Proxy/client/internal/user"
)

func ParseHandshake(req []byte, allowNoAuth bool) (byte, error) {
	if len(req) < 3 {
		return 0x00, fmt.Errorf("invalid handshake request")
	}

	ver := req[0]
	nMethods := int(req[1])

	if ver != 0x05 {
		return 0x00, fmt.Errorf("invalid version in handshake request")
	}
	if nMethods < 1 || len(req) < 2+nMethods {
		return 0x00, fmt.Errorf("invalid nMethods in handshake request")
	}

	// ищем метод без аутентификации (0x00) или с аутентификацией (0x02)
	hasNoAuth := false
	for i := 0; i < nMethods; i++ {
		switch req[2+i] {
		case 0x00:
			if allowNoAuth {
				hasNoAuth = true
			}

		case 0x02:
			return 0x02, nil
		}
	}

	if hasNoAuth {
		return 0x00, nil
	}

	return 0x00, ErrNoAcceptableMethods
}

func ParseAuthRequest(req []byte) (*user.User, error) {
	if len(req) < 5 {
		return nil, fmt.Errorf("invalid auth request")
	}

	ver := req[0]
	if ver != 0x01 {
		return nil, fmt.Errorf("invalid auth request")
	}

	uLen := int(req[1])
	if len(req) < 3+uLen {
		return nil, fmt.Errorf("invalid auth request")
	}
	uName := req[2 : 2+uLen]

	pLen := int(req[2+uLen])
	if len(req) != 3+uLen+pLen {
		return nil, fmt.Errorf("invalid auth request")
	}
	passwd := req[3+uLen:]

	return &user.User{
		Username: string(uName),
		Password: string(passwd),
	}, nil
}

func ParseConnectRequest(req []byte) (byte, string, error) {
	if len(req) < 4 {
		return 0x00, "", fmt.Errorf("invalid connect request")
	}

	ver := req[0]
	cmd := req[1]
	atyp := req[3]

	if ver != 0x05 || (cmd != 0x01 && cmd != 0x03) {
		return 0x00, "", fmt.Errorf("invalid connection request")
	}

	if cmd == 0x03 {
		return cmd, "", nil
	}

	var addr string
	var port uint16

	switch atyp {
	case 0x01:
		if len(req) != 10 {
			return 0x00, "", fmt.Errorf("invalid ipv4 in connect request")
		}

		ip := net.IP(req[4:8])
		port = uint16(req[8])<<8 | uint16(req[9])
		addr = fmt.Sprintf("%s:%d", ip, port)

	case 0x03:
		nameLen := req[4]

		if len(req) != 7+int(nameLen) {
			return 0x00, "", fmt.Errorf("invalid domain in connect request")
		}

		host := string(req[5 : 5+nameLen])
		port = uint16(req[5+nameLen])<<8 | uint16(req[6+nameLen])
		addr = fmt.Sprintf("%s:%d", host, port)

	default:
		return 0x00, "", fmt.Errorf("unsupported type of address")
	}

	return cmd, addr, nil
}
