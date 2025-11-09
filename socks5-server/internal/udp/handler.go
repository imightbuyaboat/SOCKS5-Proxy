package udp

import (
	"context"
	"encoding/binary"
	"net"
	"sync"
	"time"

	"github.com/imightbuyaboat/SOCKS5-Proxy/pkg/constants"
	"github.com/imightbuyaboat/SOCKS5-Proxy/pkg/crypto"
	"github.com/imightbuyaboat/SOCKS5-Proxy/pkg/udp_associate"
	"go.uber.org/zap"
)

type UDPAssociateHandler interface {
	HandleUDPAssociateConn(ctx context.Context, remoteConn *crypto.SecureConn, conn net.Conn)
}

type udpAssociateHandler struct {
	logger *zap.Logger
}

func NewUDPAssociateHandler(logger *zap.Logger) UDPAssociateHandler {
	return &udpAssociateHandler{
		logger: logger,
	}
}

func (h *udpAssociateHandler) HandleUDPAssociateConn(ctx context.Context, remoteConn *crypto.SecureConn, conn net.Conn) {
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
			zap.String("udp_listen_address", udpAddr.String()),
			zap.Error(err))
		return
	}
	defer udpConn.Close()

	// успешное подключение к relay-серверу
	port := udpConn.LocalAddr().(*net.UDPAddr).Port
	portBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(portBytes, uint16(port))
	if _, err := conn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x7F, 0x00, 0x00, 0x01, portBytes[0], portBytes[1]}); err != nil {
		return
	}

	h.logger.Debug("successfully start UDP listener", zap.Int("port", port))

	var clientAddr *net.UDPAddr
	var clientAddrMu sync.Mutex

	var wg sync.WaitGroup

	// читаем пакеты с relay-сервера
	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			buf := make([]byte, constants.BLOCK_SIZE)

			remoteConn.SetReadDeadline(time.Now().Add(constants.ReadWriteTimeout))
			n, err := remoteConn.Read(buf)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					h.logger.Debug("UDP read timeout from relay-server, closing connection")
					return
				}
				h.logger.Error("failed to read UDP packet from relay-server",
					zap.Error(err))
				return
			}
			remoteConn.SetReadDeadline(time.Time{})

			// парсим заголовок пакета
			_, payload, err := udp_associate.ParseUDPPacket(buf[:n])
			if err != nil {
				h.logger.Error("failed to parse UDP header",
					zap.Error(err))
				return
			}

			h.logger.Info("read and parse packet from relay-server",
				zap.Int("length", n),
				zap.Int("payload_length", len(payload)))

			clientAddrMu.Lock()
			if clientAddr == nil {
				h.logger.Warn("client address not set")
				continue
			}
			currentClientAddr := clientAddr
			clientAddrMu.Unlock()

			// формируем новый заголовок
			header, err := udp_associate.BuildSocks5UDPHeader(currentClientAddr.String())
			if err != nil {
				h.logger.Error("failed to build UDP Associate header",
					zap.Error(err))
				return
			}

			packet := header.Bytes()
			packet = append(packet, payload...)

			// пересылаем пакет клиенту
			udpConn.SetWriteDeadline(time.Now().Add(constants.ReadWriteTimeout))
			_, err = udpConn.WriteToUDP(packet, clientAddr)
			if err != nil {
				h.logger.Error("failed to write UDP packet to client",
					zap.String("client_address", udpAddr.String()),
					zap.Error(err))
				return
			}
			udpConn.SetWriteDeadline(time.Time{})
		}
	}()

	// отправляем пакеты на relay-сервер
	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			buf := make([]byte, constants.BLOCK_SIZE)

			udpConn.SetReadDeadline(time.Now().Add(constants.ReadWriteTimeout))
			n, addr, err := udpConn.ReadFromUDP(buf)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					h.logger.Debug("UDP read timeout, no more packets from client")
					return
				}
				h.logger.Error("failed to read UDP packet from client",
					zap.String("client_address", addr.String()),
					zap.Error(err))
				return
			}
			udpConn.SetReadDeadline(time.Time{})

			clientAddrMu.Lock()
			if clientAddr == nil {
				clientAddr = addr
				h.logger.Info("set client address",
					zap.String("client_address", addr.String()))
			}
			clientAddrMu.Unlock()

			h.logger.Info("read UDP packet from client",
				zap.String("client_address", addr.String()),
				zap.Int("length", n))

			// отправляем пакеты на relay-сервер
			remoteConn.SetWriteDeadline(time.Now().Add(constants.ReadWriteTimeout))
			_, err = remoteConn.Write(buf[:n])
			if err != nil {
				h.logger.Error("failed to write UDP packet to relay-server",
					zap.Error(err))
				return
			}
			remoteConn.SetWriteDeadline(time.Time{})

			h.logger.Info("send UDP packet to relay-server")
		}
	}()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		h.logger.Debug("successfully finish reading and writing to relay-server")
	case <-ctx.Done():
		h.logger.Debug("context closed")
	}
}
