package udp

import (
	"net"

	"github.com/imightbuyaboat/SOCKS5-Proxy/pkg/block"
	"github.com/imightbuyaboat/SOCKS5-Proxy/pkg/crypto"
	"github.com/imightbuyaboat/SOCKS5-Proxy/pkg/udp_header"
	"go.uber.org/zap"
)

func (l *UDPAssociateListener) handleUDPRelay(conn net.Conn) {
	defer conn.Close()

	// генерируем ключ
	key, err := crypto.GenerateSharedSecret(conn, false)
	if err != nil {
		l.logger.Error("failed to generate key",
			zap.String("address", conn.RemoteAddr().String()),
			zap.Error(err))
		return
	}

	// устанавливаем защищенное соединение с прокси-клиентом
	secureConn, err := crypto.NewSecureConn(conn, key)
	if err != nil {
		l.logger.Error("failed to create secure connection",
			zap.String("address", conn.RemoteAddr().String()),
			zap.Error(err))
		return
	}

	for {
		buf := make([]byte, block.BLOCK_SIZE)

		// читаем пакет
		n, err := secureConn.Read(buf)
		if err != nil {
			l.logger.Error("failed to read UDP packet",
				zap.String("address", conn.RemoteAddr().String()),
				zap.Error(err))
			return
		}

		// парсим пакет
		header, payload, err := udp_header.ParseUDPPacket(buf[:n])
		if err != nil {
			l.logger.Error("failed to parse UDP packet",
				zap.String("address", conn.RemoteAddr().String()),
				zap.Error(err))
			return
		}

		dstAddr := header.DST()

		l.logger.Info("successfully read and parse packet",
			zap.String("address", conn.RemoteAddr().String()),
			zap.Int("length", n),
			zap.String("target_address", dstAddr))

		// устанавливаем содениние с целевым адресом
		remoteConn, err := createRemoteUDPConnection(dstAddr)
		if err != nil {
			l.logger.Error("failed to create remote connection",
				zap.String("target_address", dstAddr),
				zap.Error(err))
			return
		}
		defer remoteConn.Close()

		l.logger.Info("successfully create connection to target address",
			zap.String("target_address", dstAddr))

		// отправляем полезную нагрузку
		_, err = remoteConn.Write(payload)
		if err != nil {
			l.logger.Error("failed to write payload to remote connection",
				zap.String("target_address", dstAddr),
				zap.Error(err))
			return
		}

		repsonse := make([]byte, block.BLOCK_SIZE)

		// читаем полезную нагрузку
		n, err = remoteConn.Read(repsonse)
		if err != nil {
			l.logger.Error("failed to read data from remote connection",
				zap.String("target_address", dstAddr),
				zap.Error(err))
			return
		}

		l.logger.Info("read response from target address",
			zap.Int("length", n))

		var packet []byte
		packet = append(packet, header.Bytes()...)
		packet = append(packet, repsonse[:n]...)

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
