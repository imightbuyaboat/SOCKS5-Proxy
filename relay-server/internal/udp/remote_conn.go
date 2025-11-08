package udp

import (
	"context"
	"net"
)

func createRemoteUDPConnection(ctx context.Context, targetAddr string) (net.Conn, error) {
	dialer := &net.Dialer{}
	return dialer.DialContext(ctx, "udp", targetAddr)
}
