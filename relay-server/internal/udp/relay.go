package udp

import (
	"context"
	"io"
	"net"
	"time"

	"github.com/imightbuyaboat/SOCKS5-Proxy/pkg/constants"
	"github.com/imightbuyaboat/SOCKS5-Proxy/pkg/crypto"
	"github.com/imightbuyaboat/SOCKS5-Proxy/pkg/udp_associate"
	"github.com/imightbuyaboat/SOCKS5-Proxy/server/internal/conn_operations"
	"go.uber.org/zap"
)

func (l *UDPAssociateListener) handleUDPRelay(ctx context.Context, conn net.Conn) {
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

	for {
		select {
		case <-ctx.Done():
			l.logger.Info("context cancelled, stopping UDP relay")
			return
		default:
		}

		buf := make([]byte, constants.BLOCK_SIZE)

		// читаем пакет
		secureConn.SetReadDeadline(time.Now().Add(constants.ReadWriteTimeout))
		n, err := secureConn.Read(buf)
		if err != nil {
			if err == io.EOF {
				l.logger.Debug("client closed connection")
				return
			}

			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				l.logger.Debug("idle timeout, closing connection")
				return
			}

			l.logger.Error("failed to read UDP packet",
				zap.String("socks5_server_address", conn.RemoteAddr().String()),
				zap.Error(err))
			return
		}
		secureConn.SetReadDeadline(time.Time{})

		// парсим пакет
		header, payload, err := udp_associate.ParseUDPPacket(buf[:n])
		if err != nil {
			l.logger.Error("failed to parse UDP packet",
				zap.String("socks5_server_address", conn.RemoteAddr().String()),
				zap.Error(err))
			return
		}

		dstAddr := header.DST()

		l.logger.Info("successfully read and parse packet",
			zap.String("socks5_server_address", conn.RemoteAddr().String()),
			zap.Int("length", n),
			zap.String("target_address", dstAddr))

		// устанавливаем соединение с целевым адресом
		remoteConn, err := conn_operations.CreateRemoteConnection(ctx, conn_operations.UDPNetwork, dstAddr)
		if err != nil {
			l.logger.Error("failed to create connection",
				zap.String("target_address", dstAddr),
				zap.Error(err))
			return
		}

		if err = l.handleSingleUDPExchange(ctx, secureConn, remoteConn, header, payload, dstAddr); err != nil {
			return
		}
	}
}

func (l *UDPAssociateListener) handleSingleUDPExchange(ctx context.Context, secureConn *crypto.SecureConn, remoteConn net.Conn, header *udp_associate.Socks5UDPAssociateHeader, payload []byte, dstAddr string) error {
	defer remoteConn.Close()

	// отправляем полезную нагрузку
	remoteConn.SetWriteDeadline(time.Now().Add(constants.ReadWriteTimeout))
	_, err := remoteConn.Write(payload)
	if err != nil {
		l.logger.Error("failed to write payload to connection",
			zap.String("target_address", dstAddr),
			zap.Error(err))
		return err
	}
	remoteConn.SetWriteDeadline(time.Time{})

	response := make([]byte, constants.BLOCK_SIZE)

	// читаем ответ
	remoteConn.SetReadDeadline(time.Now().Add(constants.ReadWriteTimeout))
	n, err := remoteConn.Read(response)
	if err != nil {
		select {
		case <-ctx.Done():
			l.logger.Info("read from target interrupted by context cancellation")
			return err
		default:
			l.logger.Error("failed to read data from target server",
				zap.String("target_address", dstAddr),
				zap.Error(err))
			return err
		}
	}
	remoteConn.SetReadDeadline(time.Time{})

	l.logger.Info("read response from target server", zap.Int("length", n))

	var packet []byte
	packet = append(packet, header.Bytes()...)
	packet = append(packet, response[:n]...)

	// отправляем пакет socks5-серверу
	secureConn.SetWriteDeadline(time.Now().Add(constants.ReadWriteTimeout))
	_, err = secureConn.Write(packet)
	if err != nil {
		l.logger.Error("failed to write to socks5-server", zap.Error(err))
		return err
	}
	secureConn.SetWriteDeadline(time.Time{})

	l.logger.Info("successfully send packet to socks5-server",
		zap.String("socks5_server_address", secureConn.RemoteAddr().String()))

	return nil
}
