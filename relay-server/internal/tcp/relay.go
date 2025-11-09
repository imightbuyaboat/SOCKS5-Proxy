package tcp

import (
	"context"
	"errors"
	"io"
	"net"
	"sync"
	"time"

	"github.com/imightbuyaboat/SOCKS5-Proxy/pkg/constants"
	"github.com/imightbuyaboat/SOCKS5-Proxy/server/internal/conn_operations"
	"go.uber.org/zap"
)

func (l *TCPAssociateListener) handleTCPRelay(ctx context.Context, conn net.Conn) {
	defer l.wg.Done()
	defer conn.Close()

	// устанавливаем защищенное соединение с socks5-сервером
	secureConn, err := conn_operations.UpgradeConnToSecureConn(ctx, conn)
	if err != nil {
		l.logger.Error("failed to create secure conn to SOCKS5 server",
			zap.String("socks5_server_address", conn.RemoteAddr().String()),
			zap.Error(err))
		return
	}

	buf := make([]byte, constants.BLOCK_SIZE)

	// читаем длину целевого адреса
	secureConn.SetReadDeadline(time.Now().Add(constants.ReadWriteTimeout))
	n, err := secureConn.Read(buf)
	if err != nil {
		l.logger.Error("failed to read length of target address",
			zap.String("socks5_server_address", conn.RemoteAddr().String()),
			zap.Error(err))
		return
	}
	secureConn.SetReadDeadline(time.Time{})

	if n == 0 {
		l.logger.Error("empty target address",
			zap.String("socks5_server_address", conn.RemoteAddr().String()))
		return
	}

	length := int(buf[0])

	// читаем адрес
	secureConn.SetReadDeadline(time.Now().Add(constants.ReadWriteTimeout))
	n, err = secureConn.Read(buf)
	if err != nil {
		l.logger.Error("failed to read target address",
			zap.String("socks5_server_address", conn.RemoteAddr().String()),
			zap.Error(err))
		return
	}
	secureConn.SetReadDeadline(time.Time{})

	if n == 0 {
		l.logger.Error("empty target address",
			zap.String("socks5_server_address", conn.RemoteAddr().String()))
		return
	}

	targetAddr := buf[:length]

	l.logger.Info("successfully read target address",
		zap.Int("length", length),
		zap.String("target_address", string(targetAddr)))

	// устанавливаем соединение с целевым сервером
	remoteConn, err := conn_operations.CreateRemoteConnection(ctx, conn_operations.TCPNetwork, string(targetAddr))
	if err != nil {
		l.logger.Error("failed to create connection",
			zap.String("socks5_server_address", conn.RemoteAddr().String()),
			zap.String("target_address", string(targetAddr)),
			zap.Error(err))
		return
	}
	defer remoteConn.Close()

	l.logger.Debug("successfully create connection to target server",
		zap.String("target_address", string(targetAddr)))

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer remoteConn.Close()

		buf := make([]byte, 1024)
		totalWritten := 0
		secureConn.SetReadDeadline(time.Now().Add(10 * time.Second))
		for {
			n, err := secureConn.Read(buf)
			secureConn.SetReadDeadline(time.Time{})
			if n > 0 {
				l.logger.Debug("read from SOCKS5", zap.Int("bytes", n))
				totalWritten += n
				remoteConn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if _, wErr := remoteConn.Write(buf[:n]); wErr != nil {
					l.logger.Error("write to target error", zap.String("target_address", string(targetAddr)), zap.Error(wErr))
					return
				}
				remoteConn.SetWriteDeadline(time.Time{})
				l.logger.Debug("wrote to target", zap.Int("bytes", n), zap.String("target_address", string(targetAddr)))
			}
			if err != nil {
				if errors.Is(err, io.EOF) {
					l.logger.Debug("EOF from SOCKS5 (end of client data)")
				} else if errors.Is(err, context.DeadlineExceeded) {
					l.logger.Warn("read from SOCKS5 timeout")
				} else {
					l.logger.Error("read from SOCKS5 error", zap.Error(err))
				}
				l.logger.Debug("total written to target", zap.Int("bytes", totalWritten), zap.String("target_address", string(targetAddr)))
				return
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer secureConn.Close()

		buf := make([]byte, 1024)
		totalRead := 0
		remoteConn.SetReadDeadline(time.Now().Add(10 * time.Second))
		for {
			n, err := remoteConn.Read(buf)
			remoteConn.SetReadDeadline(time.Time{})
			if n > 0 {
				l.logger.Debug("read from target", zap.String("target_address", string(targetAddr)), zap.Int("bytes", n))
				totalRead += n
				secureConn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if _, wErr := secureConn.Write(buf[:n]); wErr != nil {
					l.logger.Error("write to SOCKS5 error", zap.Error(wErr))
					return
				}
				secureConn.SetWriteDeadline(time.Time{})
				l.logger.Debug("wrote to SOCKS5", zap.Int("bytes", n))
			}
			if err != nil {
				if errors.Is(err, io.EOF) {
					l.logger.Debug("EOF from target (normal close)", zap.String("target_address", string(targetAddr)))
				} else if errors.Is(err, context.DeadlineExceeded) {
					l.logger.Warn("read from target timeout", zap.String("target_address", string(targetAddr)))
				} else {
					l.logger.Error("read from target error", zap.String("target_address", string(targetAddr)), zap.Error(err))
				}
				l.logger.Debug("total read from target", zap.Int("bytes", totalRead), zap.String("target_address", string(targetAddr)))
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
		l.logger.Debug("successfully read and write to socks5-server")
	case <-ctx.Done():
		l.logger.Debug("context closed")
	case <-time.After(60 * time.Second):
		l.logger.Warn("timeout on relay goroutines")
	}
}
