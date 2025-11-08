package tcp

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/imightbuyaboat/SOCKS5-Proxy/pkg/constants"
	"github.com/imightbuyaboat/SOCKS5-Proxy/pkg/crypto"
	"go.uber.org/zap"
)

func (l *TCPAssociateListener) handleTCPRelay(ctx context.Context, conn net.Conn) {
	defer l.wg.Done()
	defer conn.Close()

	// устанавливаем защищенное соединение с socks5-сервером
	secureConn, err := l.createSecureConnToSOCK5Server(ctx, conn)
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
	remoteConn, err := createRemoteTCPConnection(ctx, string(targetAddr))
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

		_, err := io.Copy(remoteConn, secureConn)
		if tcpConn, ok := remoteConn.(*net.TCPConn); ok {
			tcpConn.CloseWrite()
		}

		if err != nil {
			select {
			case <-ctx.Done():
				l.logger.Debug("context closed")
			default:
				l.logger.Error("error while wrtiting to remote conn")
			}
		} else {
			l.logger.Debug("successfully write tcp packet to target address")
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		_, err := io.Copy(secureConn, remoteConn)
		secureConn.CloseWrite()

		if err != nil {
			select {
			case <-ctx.Done():
				l.logger.Debug("context closed")
			default:
				l.logger.Error("error while wrtiting to conn")
			}
		} else {
			l.logger.Debug("successfully read tcp packet from target address")
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
	}
}

func (l *TCPAssociateListener) createSecureConnToSOCK5Server(ctx context.Context, conn net.Conn) (*crypto.SecureConn, error) {
	// генерируем разделяемый секрет
	key, err := crypto.GenerateSharedSecret(ctx, conn, false)
	if err != nil {
		return nil, fmt.Errorf("failed to generate shared secret: %v", err)
	}

	// устанавливаем защищенное соединение с socks5-сервером
	secureConn, err := crypto.NewSecureConn(conn, key)
	if err != nil {
		return nil, fmt.Errorf("failed to create secure connection: %v", err)
	}

	return secureConn, nil
}
