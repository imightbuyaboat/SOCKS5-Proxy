package udp

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/imightbuyaboat/SOCKS5-Proxy/server/internal/config"
	"go.uber.org/zap"
)

type UDPAssociateListener struct {
	config   *config.UDPListenerConfig
	listener net.Listener
	wg       *sync.WaitGroup
	logger   *zap.Logger
}

func NewUDPAssociateListener(config *config.UDPListenerConfig, logger *zap.Logger) *UDPAssociateListener {
	return &UDPAssociateListener{
		config: config,
		wg:     &sync.WaitGroup{},
		logger: logger,
	}
}

func (l *UDPAssociateListener) Start(ctx context.Context) {
	listener, err := net.Listen("tcp", l.config.UDPRelayServerAddress)
	if err != nil {
		l.logger.Fatal("failed to start UDPAssociateListener",
			zap.String("udp_relay_server_address", l.config.UDPRelayServerAddress),
			zap.Error(err))
	}
	l.listener = listener

	l.logger.Info("UDPAssociateListener started",
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
		go l.handleUDPRelay(ctx, conn)
	}
}

func (l *UDPAssociateListener) UpdateAddress(addr string) {
	l.config.UDPRelayServerAddress = addr
}

func (l *UDPAssociateListener) GetAddress() string {
	return l.config.UDPRelayServerAddress
}
