package tcp

import (
	"context"
	"net"

	"github.com/imightbuyaboat/SOCKS5-Proxy/pkg/config"
	"github.com/imightbuyaboat/SOCKS5-Proxy/pkg/logger"
	"go.uber.org/zap"
)

type TCPAssociateListener struct {
	config   *config.Config
	logger   *zap.Logger
	listener net.Listener
}

func NewTCPAssociateListener(config *config.Config, logger *zap.Logger) *TCPAssociateListener {
	return &TCPAssociateListener{
		config: config,
		logger: logger,
	}
}

func (l *TCPAssociateListener) Start(ctx context.Context) {
	listener, err := net.Listen("tcp", l.config.TCPRelayServerAddress)
	if err != nil {
		l.logger.Fatal("failed to start TCPAssociateListener on",
			zap.String("tcp_relay_server_address", l.config.TCPRelayServerAddress),
			zap.Error(err))
	}
	l.listener = listener

	l.logger.Info("TCPAssociateListener listening on",
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
				return

			default:
				l.logger.Error("failed to accept connection from socks5-server",
					zap.Error(err))
				continue
			}
		}
		l.logger.Info("successfully accepted connection",
			zap.String("socks5_server_address", conn.RemoteAddr().String()))

		go l.handleTCPRelay(conn)
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
