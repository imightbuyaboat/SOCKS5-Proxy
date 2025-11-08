package socks5

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/imightbuyaboat/SOCKS5-Proxy/client/internal/config"
	"github.com/imightbuyaboat/SOCKS5-Proxy/client/internal/models"
	"github.com/imightbuyaboat/SOCKS5-Proxy/client/internal/parser"
	"github.com/imightbuyaboat/SOCKS5-Proxy/client/internal/tcp"
	"github.com/imightbuyaboat/SOCKS5-Proxy/client/internal/udp"
	"github.com/imightbuyaboat/SOCKS5-Proxy/pkg/logger"
	"go.uber.org/zap"
)

type SOCKS5Listener struct {
	config              *config.SOCKS5ListenerConfig
	p                   parser.Parser
	storage             models.Storage
	listener            net.Listener
	tcpAssociateHandler tcp.TCPAssociateHandler
	udpAssociateHandler udp.UDPAssociateHandler
	wg                  *sync.WaitGroup
	logger              *zap.Logger
}

func NewSOCKS5Listener(config *config.SOCKS5ListenerConfig, p parser.Parser, storage models.Storage, tcpAssociateHandler tcp.TCPAssociateHandler, udpAssociateHandler udp.UDPAssociateHandler, logger *zap.Logger) *SOCKS5Listener {
	return &SOCKS5Listener{
		config:              config,
		p:                   p,
		storage:             storage,
		tcpAssociateHandler: tcpAssociateHandler,
		udpAssociateHandler: udpAssociateHandler,
		wg:                  &sync.WaitGroup{},
		logger:              logger,
	}
}

func (s *SOCKS5Listener) Start(ctx context.Context) {
	listener, err := net.Listen("tcp", s.config.SOCKS5ServerAddress)
	if err != nil {
		s.logger.Fatal("failed to start SOCKS5Listener",
			zap.String("socks5_server_address", s.config.SOCKS5ServerAddress),
			zap.Error(err))
		return
	}
	s.listener = listener

	s.logger.Info("SOCKS5 proxy started",
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
				s.logger.Info("waiting for active connections to finish")

				done := make(chan struct{})
				go func() {
					s.wg.Wait()
					close(done)
				}()

				select {
				case <-done:
					s.logger.Info("all connections closed")
				case <-time.After(30 * time.Second):
					s.logger.Warn("shutdown timeout")
				}
				return

			default:
				s.logger.Error("failed to accept connection from client",
					zap.Error(err))
				continue
			}
		}

		s.logger.Info("successfully accepted conncetion",
			zap.String("client_address", conn.RemoteAddr().String()))

		s.wg.Add(1)
		go s.handleConnection(ctx, conn)
	}
}

func (s *SOCKS5Listener) UpdateConfig(newConfig *config.SOCKS5ListenerConfig) {
	s.config.SOCKS5ServerAddress = newConfig.SOCKS5ServerAddress
	s.config.TCPRelayServerAddress = newConfig.TCPRelayServerAddress
	s.config.UDPRelayServerAddress = newConfig.UDPRelayServerAddress
	s.config.AllowNoAuth = newConfig.AllowNoAuth
}

func (s *SOCKS5Listener) GetConfig() *config.SOCKS5ListenerConfig {
	return s.config
}

func (l *SOCKS5Listener) GetLogs() string {
	return logger.GetLogBuffer()
}

func (l *SOCKS5Listener) ClearLogs() {
	logger.ClearLogBuffer()
}
