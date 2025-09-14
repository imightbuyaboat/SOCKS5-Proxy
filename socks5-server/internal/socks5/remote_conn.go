package socks5

import (
	"net"
)

func createRemoteConnectionToRelayServer(remoteAddr string) (net.Conn, error) {
	// заглушка для docker
	if remoteAddr == "0.0.0.0:1081" {
		remoteAddr = "172.17.0.1:1081"
	} else if remoteAddr == "0.0.0.0:1082" {
		remoteAddr = "172.17.0.1:1082"
	}

	remoteConn, err := net.Dial("tcp", remoteAddr)
	if err != nil {
		return nil, err
	}

	return remoteConn, nil
}
