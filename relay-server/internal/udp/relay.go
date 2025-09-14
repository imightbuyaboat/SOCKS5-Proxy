package udp

import (
	"net"

	"github.com/imightbuyaboat/SOCKS5-Proxy/pkg/block"
	"github.com/imightbuyaboat/SOCKS5-Proxy/pkg/crypto"
	"github.com/imightbuyaboat/SOCKS5-Proxy/pkg/udp_associate"
	"go.uber.org/zap"
)

func (l *UDPAssociateListener) handleUDPRelay(conn net.Conn) {
	defer conn.Close()

	// генерируем разделяесый секрет
	key, err := crypto.GenerateSharedSecret(conn, false)
	if err != nil {
		l.logger.Error("failed to generate shared secret",
			zap.String("socks5_server_address", conn.RemoteAddr().String()),
			zap.Error(err))
		return
	}

	// устанавливаем защищенное соединение с socks5-сервером
	secureConn, err := crypto.NewSecureConn(conn, key)
	if err != nil {
		l.logger.Error("failed to create secure connection",
			zap.String("socks5_server_address", conn.RemoteAddr().String()),
			zap.Error(err))
		return
	}

	for {
		buf := make([]byte, block.BLOCK_SIZE)

		// читаем пакет
		n, err := secureConn.Read(buf)
		if err != nil {
			l.logger.Error("failed to read UDP packet",
				zap.String("socks5_server_address", conn.RemoteAddr().String()),
				zap.Error(err))
			return
		}

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

		// устанавливаем соедниние с целевым адресом
		remoteConn, err := createRemoteUDPConnection(dstAddr)
		if err != nil {
			l.logger.Error("failed to create connection",
				zap.String("target_address", dstAddr),
				zap.Error(err))
			return
		}
		defer remoteConn.Close()

		l.logger.Info("successfully create connection to target server",
			zap.String("target_address", dstAddr))

		// отправляем полезную нагрузку
		_, err = remoteConn.Write(payload)
		if err != nil {
			l.logger.Error("failed to write payload to connection",
				zap.String("target_address", dstAddr),
				zap.Error(err))
			return
		}

		repsonse := make([]byte, block.BLOCK_SIZE)

		// читаем полезную нагрузку
		n, err = remoteConn.Read(repsonse)
		if err != nil {
			l.logger.Error("failed to read data from target server",
				zap.String("target_address", dstAddr),
				zap.Error(err))
			return
		}

		l.logger.Info("read response from target server",
			zap.Int("length", n))

		var packet []byte
		packet = append(packet, header.Bytes()...)
		packet = append(packet, repsonse[:n]...)

		// отправляем пакет socks5-серверу
		_, err = secureConn.Write(packet)
		if err != nil {
			l.logger.Error("failed to write to socks5-server",
				zap.Error(err))
			return
		}

		l.logger.Info("successfully send packet to socks5-server",
			zap.String("socks5_server_address", conn.RemoteAddr().String()))
	}
}
