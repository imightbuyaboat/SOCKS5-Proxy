package tcp

import (
	"context"
	"io"
	"net"
	"sync"
	"time"

	"github.com/imightbuyaboat/SOCKS5-Proxy/pkg/constants"
	"github.com/imightbuyaboat/SOCKS5-Proxy/pkg/crypto"
	"go.uber.org/zap"
)

type TCPAssociateHandler interface {
	HandleTCPAssociateConn(ctx context.Context, targetAddr string, remoteConn *crypto.SecureConn, conn net.Conn)
}

type tcpAssociateHandler struct {
	logger *zap.Logger
}

func NewTCPAssociateHandler(logger *zap.Logger) TCPAssociateHandler {
	return &tcpAssociateHandler{
		logger: logger,
	}
}

func (h *tcpAssociateHandler) HandleTCPAssociateConn(ctx context.Context, targetAddr string, remoteConn *crypto.SecureConn, conn net.Conn) {
	defer conn.Close()
	defer remoteConn.Close()

	// отправляем целевой адрес и его длину
	addrBytes := []byte(targetAddr)
	length := byte(len(addrBytes))

	remoteConn.SetWriteDeadline(time.Now().Add(constants.ReadWriteTimeout))
	if _, err := remoteConn.Write([]byte{length}); err != nil {
		h.logger.Error("failed to write length of target address",
			zap.String("target_address", targetAddr),
			zap.Int("length", len(addrBytes)),
			zap.Error(err))
		return
	}
	remoteConn.SetWriteDeadline(time.Time{})

	remoteConn.SetWriteDeadline(time.Now().Add(constants.ReadWriteTimeout))
	if _, err := remoteConn.Write(addrBytes); err != nil {
		h.logger.Error("failed to write target address",
			zap.String("target_address", targetAddr),
			zap.Error(err))
		return
	}
	remoteConn.SetWriteDeadline(time.Time{})

	h.logger.Debug("successfully write target address and length to relay-server",
		zap.String("target_address", targetAddr), zap.Int("length", len(addrBytes)))

	// успешное подключение к прокси-серверу
	conn.SetWriteDeadline(time.Now().Add(constants.ReadWriteTimeout))
	if _, err := conn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}); err != nil {
		h.logger.Error("failed to write success response",
			zap.String("target_address", targetAddr),
			zap.Error(err))
		return
	}
	conn.SetWriteDeadline(time.Time{})

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		_, err := io.Copy(remoteConn, conn)
		remoteConn.CloseWrite()

		if err != nil {
			select {
			case <-ctx.Done():
				h.logger.Debug("context closed")
			default:
				h.logger.Error("error while wrtiting to remote conn")
			}
		} else {
			h.logger.Debug("successfully write tcp packet to relay-server")
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		_, err := io.Copy(conn, remoteConn)
		if tcpConn, ok := conn.(*net.TCPConn); ok {
			tcpConn.CloseWrite()
		}

		if err != nil {
			select {
			case <-ctx.Done():
				h.logger.Debug("context closed")
			default:
				h.logger.Error("error while wrtiting to conn")
			}
		} else {
			h.logger.Debug("successfully read tcp packet from relay-server")
		}
	}()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		h.logger.Debug("successfully read and write to relay-server")
	case <-ctx.Done():
		h.logger.Debug("context closed")
	}
}
