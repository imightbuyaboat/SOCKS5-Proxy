package conn_operations

import (
	"context"
	"fmt"
	"net"

	"github.com/imightbuyaboat/SOCKS5-Proxy/pkg/crypto"
)

const (
	TCPNetwork = "tcp"
	UDPNetwork = "udp"
)

func CreateRemoteConnection(ctx context.Context, network, targetAddr string) (net.Conn, error) {
	dialer := &net.Dialer{}
	return dialer.DialContext(ctx, network, targetAddr)
}

func UpgradeConnToSecureConn(ctx context.Context, conn net.Conn) (*crypto.SecureConn, error) {
	// генерируем разделяемый секрет
	key, err := crypto.GenerateSharedSecret(ctx, conn, false)
	if err != nil {
		return nil, fmt.Errorf("failed to generate shared secret: %v", err)
	}

	// устанавливаем защищенное соединение с socks5-сервером
	secureConn, err := crypto.NewSecureConn(conn, key)
	if err != nil {
		return nil, fmt.Errorf("failed to create secure connection: %v", err)
	}

	return secureConn, nil
}
