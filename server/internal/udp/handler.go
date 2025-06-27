package udp

import (
	"net"

	"github.com/imightbuyaboat/SOCKS5-Proxy/pkg/constants"
	"github.com/imightbuyaboat/SOCKS5-Proxy/pkg/crypto"
	"github.com/imightbuyaboat/SOCKS5-Proxy/pkg/udp_header"
	"go.uber.org/zap"
)

func (l *UDPAssociateListener) HandleConnection(conn net.Conn) {
	defer conn.Close()

	// устанавливаем защищенное соединение с прокси-клиентом
	secureConn, err := crypto.NewSecureConn(conn, l.config.Key)
	if err != nil {
		l.logger.Error("failed to create secure connection",
			zap.String("address", conn.RemoteAddr().String()),
			zap.Error(err))
		return
	}

	for {
		buf := make([]byte, constants.BLOCK_SIZE)

		// читаем пакет
		n, err := secureConn.Read(buf)
		if err != nil {
			l.logger.Error("failed to read UDP packet",
				zap.String("address", conn.RemoteAddr().String()),
				zap.Error(err))
			return
		}

		// парсим пакет
		srcAddr, targetAddr, payload, err := udp_header.ParseSocks5UDPHeader(buf[:n])
		if err != nil {
			l.logger.Error("failed to parse UDP packet",
				zap.String("address", conn.RemoteAddr().String()),
				zap.Error(err))
			return
		}

		l.logger.Info("successfully read and parse packet",
			zap.String("address", conn.RemoteAddr().String()),
			zap.Int("length", n),
			zap.String("target_address", targetAddr))

		// устанавливаем содениние с целевым адресом
		remoteConn, err := createRemoteUDPConnection(string(targetAddr))
		if err != nil {
			l.logger.Error("failed to create remote connection",
				zap.String("target_address", string(targetAddr)),
				zap.Error(err))
			return
		}
		defer remoteConn.Close()

		l.logger.Info("successfully create connection to target address",
			zap.String("target_address", string(targetAddr)))

		// отправляем полезную нагрузку
		_, err = remoteConn.Write(payload)
		if err != nil {
			l.logger.Error("failed to write payload to remote connection",
				zap.String("target_address", string(targetAddr)),
				zap.Error(err))
			return
		}

		// читаем полезную нагрузку
		n, err = remoteConn.Read(buf)
		if err != nil {
			l.logger.Error("failed to read data from remote connection",
				zap.String("target_address", string(targetAddr)),
				zap.Error(err))
			return
		}

		// генерируем пакет с заголовком
		packet, err := udp_header.BuildSocks5UDPHeader(targetAddr, srcAddr)
		if err != nil {
			l.logger.Error("failed to build UDP header",
				zap.Error(err))
			return
		}
		packet = append(packet, buf[:n]...)

		// отправляем пакет прокси-клиенту
		_, err = secureConn.Write(packet)
		if err != nil {
			l.logger.Error("failed to write to proxy-client",
				zap.Error(err))
			return
		}

		l.logger.Info("successfully send packet to proxy-client",
			zap.String("address", conn.RemoteAddr().String()))
	}
}
