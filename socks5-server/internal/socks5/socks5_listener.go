package socks5

import (
	"context"
	"net"

	"github.com/imightbuyaboat/SOCKS5-Proxy/pkg/config"
	"github.com/imightbuyaboat/SOCKS5-Proxy/pkg/logger"
	"go.uber.org/zap"
)

type SOCKS5Listener struct {
	config *config.Config

	logger *zap.Logger

	storage Storage

	listener net.Listener
}

func NewSOCKS5Listener(config *config.Config, logger *zap.Logger, storage Storage) *SOCKS5Listener {
	return &SOCKS5Listener{
		config:  config,
		logger:  logger,
		storage: storage,
	}
}

func (s *SOCKS5Listener) Start(ctx context.Context) {
	listener, err := net.Listen("tcp", s.config.SOCKS5ServerAddress)
	if err != nil {
		s.logger.Fatal("failed to start SOCKS5Listener on",
			zap.String("socks5_server_address", s.config.SOCKS5ServerAddress),
			zap.Error(err))
	}
	s.listener = listener

	s.logger.Info("SOCKS5 proxy listening on",
		zap.String("socks5_server_address", s.config.SOCKS5ServerAddress))

	go func() {
		<-ctx.Done()
		s.logger.Info("shutting down SOCKS5Listener")
		s.listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return

			default:
				s.logger.Error("failed to accept connection from client",
					zap.Error(err))
				continue
			}
		}

		s.logger.Info("successfully accepted conncetion",
			zap.String("client_address", conn.RemoteAddr().String()))

		go s.handleConnection(conn)
	}
}

func (s *SOCKS5Listener) UpdateConfig(newConfig *config.Config) {
	s.config.SOCKS5ServerAddress = newConfig.SOCKS5ServerAddress
	s.config.TCPRelayServerAddress = newConfig.TCPRelayServerAddress
	s.config.UDPRelayServerAddress = newConfig.UDPRelayServerAddress
	s.config.AllowNoAuth = newConfig.AllowNoAuth
}

func (s *SOCKS5Listener) GetConfig() *config.Config {
	return s.config
}

func (l *SOCKS5Listener) GetLogs() string {
	return logger.GetLogBuffer()
}

func (l *SOCKS5Listener) ClearLogs() {
	logger.ClearLogBuffer()
}
