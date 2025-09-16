package socks5

import (
	"net"

	"github.com/imightbuyaboat/SOCKS5-Proxy/pkg/config"
	"go.uber.org/zap"
)

type SOCKS5Listener struct {
	config *config.Config

	logger *zap.Logger

	storage Storage
}

func NewSOCKS5Listener(config *config.Config, logger *zap.Logger, storage Storage) *SOCKS5Listener {
	return &SOCKS5Listener{
		config:  config,
		logger:  logger,
		storage: storage,
	}
}

func (s *SOCKS5Listener) Start() {
	listener, err := net.Listen("tcp", s.config.SOCKS5ServerAddress)
	if err != nil {
		s.logger.Fatal("failed to start SOCKS5Listener on",
			zap.String("socks5_server_address", s.config.SOCKS5ServerAddress),
			zap.Error(err))
	}
	s.logger.Info("SOCKS5 proxy listening on",
		zap.String("socks5_server_address", s.config.SOCKS5ServerAddress))

	for {
		conn, err := listener.Accept()
		if err != nil {
			s.logger.Error("failed to accept connection from client",
				zap.Error(err))
			continue
		}
		s.logger.Info("successfully accepted conncetion",
			zap.String("client_address", conn.RemoteAddr().String()))

		go s.handleConnection(conn)
	}
}
