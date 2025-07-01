package udp

import (
	"encoding/binary"
	"net"

	"github.com/imightbuyaboat/SOCKS5-Proxy/pkg/block"
	"github.com/imightbuyaboat/SOCKS5-Proxy/pkg/crypto"
	"github.com/imightbuyaboat/SOCKS5-Proxy/pkg/udp_header"
	"go.uber.org/zap"
)

type UDPAssociateHandler struct {
	logger *zap.Logger
}

func NewUDPAssociateHandler(logger *zap.Logger) *UDPAssociateHandler {
	return &UDPAssociateHandler{
		logger: logger,
	}
}

func (h *UDPAssociateHandler) HandleUDPAssociateConn(remoteConn *crypto.SecureConn, conn net.Conn) {
	// определяем адрес, с которого будем принимать UDP пакеты
	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		conn.Write([]byte{0x05, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})

		h.logger.Error("failed to resolve UDP address",
			zap.Error(err))
		return
	}

	// слушаем порт, на который приходят UDP пакеты
	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		conn.Write([]byte{0x05, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})

		h.logger.Error("failed to start UDP listener",
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

	h.logger.Info("listen udp packets",
		zap.Int("port", port))

	var clientAddr *net.UDPAddr

	// читаем пакеты с прокси-сервера
	go func() {
		for {
			buf := make([]byte, block.BLOCK_SIZE)

			n, err := remoteConn.Read(buf)
			if err != nil {
				h.logger.Error("failed to read UDP packet from remote connection",
					zap.Error(err))
				return
			}

			// парсим заголовок пакета
			_, payload, err := udp_header.ParseUDPPacket(buf[:n])
			if err != nil {
				h.logger.Error("failed to parse UDP header",
					zap.Error(err))
				return
			}

			if clientAddr == nil {
				h.logger.Warn("client address not set")
				continue
			}

			// пересылаем полезную нагрузку пользователю
			_, err = udpConn.WriteToUDP(payload, clientAddr)
			if err != nil {
				h.logger.Error("failed to write packet to UDP connection",
					zap.String("client_address", udpAddr.String()),
					zap.Error(err))
				return
			}
		}
	}()

	// отправляем пакеты на прокси-сервер
	for {
		buf := make([]byte, block.BLOCK_SIZE)

		n, addr, err := udpConn.ReadFromUDP(buf)
		if err != nil {
			h.logger.Error("failed to read packet from UDP connection",
				zap.String("client_address", addr.String()),
				zap.Error(err))
			return
		}

		if clientAddr == nil {
			clientAddr = addr
			h.logger.Info("set client addr",
				zap.String("client_address", addr.String()))
		}

		h.logger.Info("succesfully receive packet from client",
			zap.Int("length", n),
			zap.String("client_address", addr.String()))

		// отправляем пакеты на прокси-сервер
		_, err = remoteConn.Write(buf[:n])
		if err != nil {
			h.logger.Error("failed to write to remote connection",
				zap.Error(err))
			return
		}

		h.logger.Info("succesfully send packet to proxy-server",
			zap.Int("length", n),
			zap.String("client_address", addr.String()))
	}
}
