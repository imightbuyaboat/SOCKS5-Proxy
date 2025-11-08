package tcp

import (
	"context"
	"net"
)

func createRemoteTCPConnection(ctx context.Context, targetAddr string) (net.Conn, error) {
	dialer := &net.Dialer{}
	return dialer.DialContext(ctx, "tcp", targetAddr)
}
