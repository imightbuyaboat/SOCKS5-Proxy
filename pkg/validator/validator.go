package validator

import (
	"fmt"
	"net"
	"strconv"
)

func ValidateAddress(addr string) error {
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return err
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return err
	}

	if port < 0 || port > 65535 {
		return fmt.Errorf("invalid port")
	}

	ip := net.ParseIP(host).To4()
	if ip == nil {
		return fmt.Errorf("invalid format of address")
	}

	return nil
}
