package socks5

import (
	"net"

	"github.com/imightbuyaboat/SOCKS5-Proxy/client/internal/parser"
	"github.com/imightbuyaboat/SOCKS5-Proxy/client/internal/tcp"
	"github.com/imightbuyaboat/SOCKS5-Proxy/client/internal/udp"
	"github.com/imightbuyaboat/SOCKS5-Proxy/pkg/block"
	"github.com/imightbuyaboat/SOCKS5-Proxy/pkg/crypto"
	"go.uber.org/zap"
)

func (s *SOCKS5Listener) handleConnection(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, block.BLOCK_SIZE)

	// handshake request
	n, err := conn.Read(buf)
	if err != nil {
		s.logger.Error("failed to read handshake request",
			zap.String("address", conn.RemoteAddr().String()),
			zap.Error(err))
		return
	}

	if err = parser.ParseHandshake(buf[:n]); err != nil {
		s.logger.Error("failed to parse handshake request",
			zap.String("address", conn.RemoteAddr().String()),
			zap.Error(err))
		return
	}

	conn.Write([]byte{0x05, 0x00})

	// connect request
	n, err = conn.Read(buf)
	if err != nil {
		s.logger.Error("failed to read connect request",
			zap.String("address", conn.RemoteAddr().String()),
			zap.Error(err))
		return
	}

	cmd, targetAddr, err := parser.ParseConnectRequest(buf[:n])
	if err != nil {
		conn.Write([]byte{0x05, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})

		s.logger.Error("failed to parse connect request",
			zap.String("address", conn.RemoteAddr().String()),
			zap.Error(err))
		return
	}

	// определение адреса прокси-сервера
	var remoteAddr string
	switch cmd {
	case 0x01:
		remoteAddr = s.config.RemoteTCPAddress
	case 0x03:
		remoteAddr = s.config.RemoteUDPAddress
	}

	// подключение к прокси-серверу
	remoteConn, err := createRemoteConnectionToProxyServer(remoteAddr)
	if err != nil {
		conn.Write([]byte{0x05, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})

		s.logger.Error("failed to create remote connection",
			zap.String("address", conn.RemoteAddr().String()),
			zap.String("remote_address", remoteAddr),
			zap.Error(err))
		return
	}
	defer remoteConn.Close()

	// генерируем ключ
	key, err := crypto.GenerateSharedSecret(remoteConn, true)
	if err != nil {
		conn.Write([]byte{0x05, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})

		s.logger.Error("failed to generate key",
			zap.String("remote_address", remoteAddr),
			zap.Error(err))
		return
	}

	// создаем защищенное подключение
	secureRemoteConn, err := crypto.NewSecureConn(remoteConn, key)
	if err != nil {
		conn.Write([]byte{0x05, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})

		s.logger.Error("failed to create secure remote connection",
			zap.String("address", conn.RemoteAddr().String()),
			zap.String("remote_address", remoteAddr),
			zap.Error(err))
		return
	}

	s.logger.Info("successfully create remote connection to proxy-server",
		zap.String("remote_tcp_address", remoteAddr))

	// вызов обработчика
	switch cmd {
	case 0x01:
		tcp.NewTCPAssociateHandler(s.logger).HandleTCPAssociateConn(targetAddr, secureRemoteConn, conn)

	case 0x03:
		udp.NewUDPAssociateHandler(s.logger).HandleUDPAssociateConn(secureRemoteConn, conn)
	}
}
