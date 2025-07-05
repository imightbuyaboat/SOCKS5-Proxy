package tcp

import (
	"net"

	"github.com/imightbuyaboat/SOCKS5-Proxy/pkg/config"
	"go.uber.org/zap"
)

type TCPAssociateListener struct {
	config *config.Config
	logger *zap.Logger
}

func NewTCPAssociateListener(config *config.Config, logger *zap.Logger) *TCPAssociateListener {
	return &TCPAssociateListener{
		config: config,
		logger: logger,
	}
}

func (l *TCPAssociateListener) Start() {
	listener, err := net.Listen("tcp", l.config.RemoteTCPAddress)
	if err != nil {
		l.logger.Fatal("failed to start TCPAssociateListener on",
			zap.String("address", l.config.RemoteTCPAddress),
			zap.Error(err))
	}
	l.logger.Info("TCPAssociateListener listening on",
		zap.String("address", l.config.RemoteTCPAddress))

	for {
		conn, err := listener.Accept()
		if err != nil {
			l.logger.Error("failed to accept connection",
				zap.Error(err))
			continue
		}
		l.logger.Info("successfully accepted connection",
			zap.String("address", conn.RemoteAddr().String()))

		go l.handleTCPRelay(conn)
	}
}
