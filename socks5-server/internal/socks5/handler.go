package socks5

import (
	"errors"
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
			zap.String("client_address", conn.RemoteAddr().String()),
			zap.Error(err))
		return
	}

	method, err := parser.ParseHandshake(buf[:n], s.config.AllowNoAuth)
	if err != nil {
		if errors.Is(err, parser.ErrNoAcceptableMethods) {
			conn.Write([]byte{0x05, 0xFF})
			s.logger.Error("no acceptable methods",
				zap.String("client_address", conn.RemoteAddr().String()),
				zap.Error(err))
			return
		}

		s.logger.Error("failed to parse handshake request",
			zap.String("client_address", conn.RemoteAddr().String()),
			zap.Error(err))
		return
	}

	switch method {
	// no auth
	case 0x00:
		conn.Write([]byte{0x05, 0x00})

	// auth
	case 0x02:
		conn.Write([]byte{0x05, 0x02})

		n, err = conn.Read(buf)
		if err != nil {
			s.logger.Error("failed to read auth request",
				zap.String("client_address", conn.RemoteAddr().String()),
				zap.Error(err))
			return
		}

		user, err := parser.ParseAuthRequest(buf[:n])
		if err != nil {
			conn.Write([]byte{0x01, 0x01})
			s.logger.Error("invalid auth request",
				zap.String("client_address", conn.RemoteAddr().String()),
				zap.Error(err))
			return
		}

		if err = s.storage.CheckUser(user); err != nil {
			conn.Write([]byte{0x01, 0x01})
			s.logger.Error("invalid auth request",
				zap.String("client_address", conn.RemoteAddr().String()),
				zap.Error(err))
			return
		}

		conn.Write([]byte{0x01, 0x00})
	}

	// connect request
	n, err = conn.Read(buf)
	if err != nil {
		s.logger.Error("failed to read connect request",
			zap.String("client_address", conn.RemoteAddr().String()),
			zap.Error(err))
		return
	}

	cmd, targetAddr, err := parser.ParseConnectRequest(buf[:n])
	if err != nil {
		conn.Write([]byte{0x05, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})

		s.logger.Error("failed to parse connect request",
			zap.String("client_address", conn.RemoteAddr().String()),
			zap.Error(err))
		return
	}

	// определение адреса прокси-сервера
	var remoteAddr string
	switch cmd {
	case 0x01:
		remoteAddr = s.config.TCPRelayServerAddress
	case 0x03:
		remoteAddr = s.config.UDPRelayServerAddress
	}

	// подключение к relay-серверу
	remoteConn, err := createRemoteConnectionToRelayServer(remoteAddr)
	if err != nil {
		conn.Write([]byte{0x05, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})

		s.logger.Error("failed to create connection to relay-server",
			zap.String("relay_server_address", remoteAddr),
			zap.Error(err))
		return
	}
	defer remoteConn.Close()

	// генерируем ключ
	key, err := crypto.GenerateSharedSecret(remoteConn, true)
	if err != nil {
		conn.Write([]byte{0x05, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})

		s.logger.Error("failed to generate shared secret",
			zap.String("relay_server_address", remoteAddr),
			zap.Error(err))
		return
	}

	// создаем защищенное подключение
	secureRemoteConn, err := crypto.NewSecureConn(remoteConn, key)
	if err != nil {
		conn.Write([]byte{0x05, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})

		s.logger.Error("failed to create secure connection to relay-server",
			zap.String("relay_server_address", remoteAddr),
			zap.Error(err))
		return
	}

	s.logger.Info("successfully create remote connection to relay-server",
		zap.String("relay_server_address", remoteAddr))

	// вызов обработчика
	switch cmd {
	case 0x01:
		tcp.NewTCPAssociateHandler(s.logger).HandleTCPAssociateConn(targetAddr, secureRemoteConn, conn)

	case 0x03:
		udp.NewUDPAssociateHandler(s.logger).HandleUDPAssociateConn(secureRemoteConn, conn)
	}
}
