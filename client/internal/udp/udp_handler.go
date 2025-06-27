package udp

import (
	"encoding/binary"
	"net"

	"github.com/imightbuyaboat/SOCKS5-Proxy/pkg/constants"
	"github.com/imightbuyaboat/SOCKS5-Proxy/pkg/crypto"
	"github.com/imightbuyaboat/SOCKS5-Proxy/pkg/udp_header"
	"go.uber.org/zap"
)

func HandleUDPAssociateConn(remoteConn *crypto.SecureConn, conn net.Conn, logger *zap.Logger) {
	// определяем адрес, с которого будем принимать UDP пакеты
	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		conn.Write([]byte{0x05, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})

		logger.Error("failed to resolve UDP address",
			zap.Error(err))
		return
	}

	// слушаем порт, на который приходят UDP пакеты
	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		conn.Write([]byte{0x05, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})

		logger.Error("failed to start UDP listener",
			zap.String("listen_address", udpAddr.String()),
			zap.Error(err))
		return
	}
	defer udpConn.Close()

	// успешное подключение к прокси-серверу
	port := udpConn.LocalAddr().(*net.UDPAddr).Port
	portBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(portBytes, uint16(port))
	conn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x7F, 0x00, 0x00, 0x01, portBytes[0], portBytes[1]})

	logger.Info("listen udp packets",
		zap.Int("port", port))

	// читаем пакеты с прокси-сервера
	go func() {
		for {
			buf := make([]byte, constants.BLOCK_SIZE)

			n, err := remoteConn.Read(buf)
			if err != nil {
				logger.Error("failed to read UDP packet from remote connection",
					zap.Error(err))
				return
			}

			// парсим заголовок пакета
			_, payload, err := udp_header.ParseSocks5UDPHeader(buf[:n])
			if err != nil {
				logger.Error("failed to parse UDP header",
					zap.Error(err))
				return
			}

			// пересылаем полезную нагрузку пользователю
			_, err = udpConn.Write(payload)
			if err != nil {
				logger.Error("failed to write packet to UDP connection",
					zap.String("client_address", udpAddr.String()),
					zap.Error(err))
				return
			}
		}
	}()

	// отправляем пакеты на прокси-сервер
	for {
		buf := make([]byte, constants.BLOCK_SIZE)

		n, clientAddr, err := udpConn.ReadFromUDP(buf)
		if err != nil {
			logger.Error("failed to read packet from UDP connection",
				zap.String("client_address", clientAddr.String()),
				zap.Error(err))
			return
		}

		logger.Info("succesfully receive packet from client",
			zap.Int("length", n),
			zap.String("client_address", clientAddr.String()))

		// отправляем пакеты на прокси-сервер
		_, err = remoteConn.Write(buf)
		if err != nil {
			logger.Error("failed to write to remote connection",
				zap.Error(err))
			return
		}

		logger.Info("succesfully send packet to proxy-server",
			zap.Int("length", len(buf)),
			zap.String("client_address", clientAddr.String()))
	}
}
