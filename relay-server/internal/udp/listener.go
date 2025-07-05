package udp

import (
	"net"

	"github.com/imightbuyaboat/SOCKS5-Proxy/pkg/config"
	"go.uber.org/zap"
)

type UDPAssociateListener struct {
	config *config.Config
	logger *zap.Logger
}

func NewUDPAssociateListener(config *config.Config, logger *zap.Logger) *UDPAssociateListener {
	return &UDPAssociateListener{
		config: config,
		logger: logger,
	}
}

func (l *UDPAssociateListener) Start() {
	listener, err := net.Listen("tcp", l.config.RemoteUDPAddress)
	if err != nil {
		l.logger.Fatal("failed to start UDPAssociateListener on",
			zap.String("address", l.config.RemoteUDPAddress),
			zap.Error(err))
	}
	l.logger.Info("UDPAssociateListener listening on",
		zap.String("address", l.config.RemoteUDPAddress))

	for {
		conn, err := listener.Accept()
		if err != nil {
			l.logger.Error("failed to accept connection",
				zap.Error(err))
			continue
		}
		l.logger.Info("successfully accepted connection",
			zap.String("address", conn.RemoteAddr().String()))

		go l.handleUDPRelay(conn)
	}
}
