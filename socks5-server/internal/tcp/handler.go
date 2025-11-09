package tcp

import (
	"context"
	"errors"
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
		defer remoteConn.Close()

		buf := make([]byte, 1024)
		totalWritten := 0
		conn.SetReadDeadline(time.Now().Add(10 * time.Second))
		for {
			n, err := conn.Read(buf)
			conn.SetReadDeadline(time.Time{})
			if n > 0 {
				h.logger.Debug("read from client", zap.String("client_address", conn.RemoteAddr().String()), zap.Int("bytes", n))
				totalWritten += n
				remoteConn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if _, wErr := remoteConn.Write(buf[:n]); wErr != nil {
					h.logger.Error("write to relay error", zap.Error(wErr))
					return
				}
				remoteConn.SetWriteDeadline(time.Time{})
				h.logger.Debug("wrote to relay", zap.Int("bytes", n))
			}
			if err != nil {
				if errors.Is(err, io.EOF) {
					h.logger.Debug("EOF from client (end of request)", zap.String("client_address", conn.RemoteAddr().String()))
				} else if errors.Is(err, context.DeadlineExceeded) {
					h.logger.Warn("read from client timeout", zap.String("client_address", conn.RemoteAddr().String()))
				} else {
					h.logger.Error("read from client error", zap.Error(err), zap.String("client_address", conn.RemoteAddr().String()))
				}
				h.logger.Debug("total written to relay", zap.Int("bytes", totalWritten), zap.String("client_address", conn.RemoteAddr().String()))
				return
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer conn.Close()

		buf := make([]byte, 1024)
		totalRead := 0
		remoteConn.SetReadDeadline(time.Now().Add(10 * time.Second))
		for {
			n, err := remoteConn.Read(buf)
			remoteConn.SetReadDeadline(time.Time{})
			if n > 0 {
				h.logger.Debug("read from relay", zap.Int("bytes", n))
				totalRead += n
				conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if _, wErr := conn.Write(buf[:n]); wErr != nil {
					h.logger.Error("write to client error", zap.String("client_address", conn.RemoteAddr().String()), zap.Error(wErr))
					return
				}
				conn.SetWriteDeadline(time.Time{})
				h.logger.Debug("wrote to client", zap.Int("bytes", n), zap.String("client_address", conn.RemoteAddr().String()))
			}
			if err != nil {
				if errors.Is(err, io.EOF) {
					h.logger.Debug("EOF from relay (normal close)")
				} else if errors.Is(err, context.DeadlineExceeded) {
					h.logger.Warn("read from relay timeout")
				} else {
					h.logger.Error("read from relay error", zap.Error(err))
				}
				h.logger.Debug("total read from relay", zap.Int("bytes", totalRead))
				return
			}
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
	case <-time.After(60 * time.Second):
		h.logger.Warn("timeout waiting for relay goroutines")
	}
}
