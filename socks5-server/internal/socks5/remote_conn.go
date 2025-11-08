package socks5

import (
	"context"
	"net"
)

func createRemoteConnectionToRelayServer(ctx context.Context, remoteAddr string) (net.Conn, error) {
	// заглушка для docker
	if remoteAddr == "0.0.0.0:1081" {
		remoteAddr = "172.17.0.1:1081"
	} else if remoteAddr == "0.0.0.0:1082" {
		remoteAddr = "172.17.0.1:1082"
	}

	dialer := &net.Dialer{}
	return dialer.DialContext(ctx, "tcp", remoteAddr)
}
