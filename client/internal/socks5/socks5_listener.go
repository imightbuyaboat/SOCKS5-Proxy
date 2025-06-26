package socks5

import (
	"net"

	"github.com/imightbuyaboat/SOCKS5-Proxy/client/internal/config"
	"go.uber.org/zap"
)

type SOCKS5Listener struct {
	config *config.Config
	logger *zap.Logger
}

func NewSOCKS5Listener(config *config.Config, logger *zap.Logger) *SOCKS5Listener {
	return &SOCKS5Listener{
		config: config,
		logger: logger,
	}
}

func (s *SOCKS5Listener) Start() {
	listener, err := net.Listen("tcp", s.config.ListenAddress)
	if err != nil {
		s.logger.Fatal("failed to start SOCKS5Listener on",
			zap.String("address", s.config.ListenAddress),
			zap.Error(err))
	}
	s.logger.Info("SOCKS5 proxy listening on",
		zap.String("address", s.config.ListenAddress))

	for {
		conn, err := listener.Accept()
		if err != nil {
			s.logger.Error("failed to accept connection",
				zap.Error(err))
			continue
		}
		s.logger.Info("successfully accepted conncetion",
			zap.String("address", conn.RemoteAddr().String()))

		go s.handleConnection(conn)
	}
}
