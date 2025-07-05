package udp

import (
	"net"
)

func createRemoteUDPConnection(targetAddr string) (net.Conn, error) {
	addr, err := net.ResolveUDPAddr("udp", targetAddr)
	if err != nil {
		return nil, err
	}

	remoteConn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return nil, err
	}

	return remoteConn, nil
}
