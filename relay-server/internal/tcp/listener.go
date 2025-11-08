package tcp

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/imightbuyaboat/SOCKS5-Proxy/pkg/logger"
	"github.com/imightbuyaboat/SOCKS5-Proxy/server/internal/config"
	"go.uber.org/zap"
)

type TCPAssociateListener struct {
	config   *config.TCPListenerConfig
	listener net.Listener
	wg       *sync.WaitGroup
	logger   *zap.Logger
}

func NewTCPAssociateListener(config *config.TCPListenerConfig, logger *zap.Logger) *TCPAssociateListener {
	return &TCPAssociateListener{
		config: config,
		wg:     &sync.WaitGroup{},
		logger: logger,
	}
}

func (l *TCPAssociateListener) Start(ctx context.Context) {
	listener, err := net.Listen("tcp", l.config.TCPRelayServerAddress)
	if err != nil {
		l.logger.Fatal("failed to start TCPAssociateListener",
			zap.String("tcp_relay_server_address", l.config.TCPRelayServerAddress),
			zap.Error(err))
	}
	l.listener = listener

	l.logger.Info("TCPAssociateListener started",
		zap.String("tcp_relay_server_address", l.config.TCPRelayServerAddress))

	go func() {
		<-ctx.Done()
		l.logger.Info("shutting down TCPAssociateListener")
		l.listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				l.logger.Info("waiting for active connections to finish")

				done := make(chan struct{})
				go func() {
					l.wg.Wait()
					close(done)
				}()

				select {
				case <-done:
					l.logger.Info("all connections closed")
				case <-time.After(30 * time.Second):
					l.logger.Warn("shutdown timeout")
				}
				return

			default:
				l.logger.Error("failed to accept connection from socks5-server",
					zap.Error(err))
				continue
			}
		}
		l.logger.Info("successfully accepted connection",
			zap.String("socks5_server_address", conn.RemoteAddr().String()))

		l.wg.Add(1)
		go l.handleTCPRelay(ctx, conn)
	}
}

func (l *TCPAssociateListener) UpdateAddress(addr string) {
	l.config.TCPRelayServerAddress = addr
}

func (l *TCPAssociateListener) GetAddress() string {
	return l.config.TCPRelayServerAddress
}

func (l *TCPAssociateListener) GetLogs() string {
	return logger.GetLogBuffer()
}

func (l *TCPAssociateListener) ClearLogs() {
	logger.ClearLogBuffer()
}
