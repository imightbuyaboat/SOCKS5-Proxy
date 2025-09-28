package udp

import (
	"context"
	"net"

	"github.com/imightbuyaboat/SOCKS5-Proxy/pkg/config"
	"go.uber.org/zap"
)

type UDPAssociateListener struct {
	config   *config.Config
	logger   *zap.Logger
	listener net.Listener
}

func NewUDPAssociateListener(config *config.Config, logger *zap.Logger) *UDPAssociateListener {
	return &UDPAssociateListener{
		config: config,
		logger: logger,
	}
}

func (l *UDPAssociateListener) Start(ctx context.Context) {
	listener, err := net.Listen("tcp", l.config.UDPRelayServerAddress)
	if err != nil {
		l.logger.Fatal("failed to start UDPAssociateListener on",
			zap.String("udp_relay_server_address", l.config.UDPRelayServerAddress),
			zap.Error(err))
	}
	l.listener = listener

	l.logger.Info("UDPAssociateListener listening on",
		zap.String("udp_relay_server_address", l.config.UDPRelayServerAddress))

	go func() {
		<-ctx.Done()
		l.logger.Info("shutting down SOCKS5Listener")
		l.listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return

			default:
				l.logger.Error("failed to accept connection from socks5-server",
					zap.Error(err))
				continue
			}
		}
		l.logger.Info("successfully accepted connection",
			zap.String("socks5_server_address", conn.RemoteAddr().String()))

		go l.handleUDPRelay(conn)
	}
}

func (l *UDPAssociateListener) UpdateAddress(addr string) {
	l.config.UDPRelayServerAddress = addr
}

func (l *UDPAssociateListener) GetAddress() string {
	return l.config.UDPRelayServerAddress
}
