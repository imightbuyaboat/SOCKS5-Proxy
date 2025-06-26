package tcp

import (
	"net"

	"github.com/imightbuyaboat/SOCKS5-Proxy/pkg/config"
	"go.uber.org/zap"
)

type SOCKS5ListenerTCP struct {
	config *config.Config
	logger *zap.Logger
}

func NewSOCKS5ListenerTCP(config *config.Config, logger *zap.Logger) *SOCKS5ListenerTCP {
	return &SOCKS5ListenerTCP{
		config: config,
		logger: logger,
	}
}

func (s *SOCKS5ListenerTCP) Start() {
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
