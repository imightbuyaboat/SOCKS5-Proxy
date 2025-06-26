package tcp

import (
	"net"
)

func createRemoteTCPConnection(targetAddr string) (net.Conn, error) {
	remoteConn, err := net.Dial("tcp", targetAddr)
	if err != nil {
		return nil, err
	}

	return remoteConn, nil
}
