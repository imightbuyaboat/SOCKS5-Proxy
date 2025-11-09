package socks5

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/imightbuyaboat/SOCKS5-Proxy/client/internal/parser"
	"github.com/imightbuyaboat/SOCKS5-Proxy/pkg/constants"
	"github.com/imightbuyaboat/SOCKS5-Proxy/pkg/crypto"
	"go.uber.org/zap"
)

func (s *SOCKS5Listener) handleConnection(ctx context.Context, conn net.Conn) {
	defer s.wg.Done()
	defer conn.Close()

	buf := make([]byte, constants.BLOCK_SIZE)

	// handshake request
	conn.SetReadDeadline(time.Now().Add(constants.ReadWriteTimeout))
	n, err := conn.Read(buf)
	if err != nil {
		s.logger.Error("failed to read handshake request",
			zap.String("client_address", conn.RemoteAddr().String()),
			zap.Error(err))
		return
	}
	conn.SetReadDeadline(time.Time{})

	s.logger.Debug("successfully read handhsake request",
		zap.String("client_address", conn.RemoteAddr().String()))

	method, err := s.p.ParseHandshake(buf[:n], s.config.AllowNoAuth)
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

	if err := s.handleAuthMethod(ctx, conn, method, buf); err != nil {
		s.logger.Error("failed to handle auth method from hanshake request",
			zap.String("client_address", conn.RemoteAddr().String()), zap.ByteString("method", []byte{method}))
		return
	}

	s.logger.Debug("successfully parse handhsake request",
		zap.String("client_address", conn.RemoteAddr().String()))

	// connect request
	conn.SetReadDeadline(time.Now().Add(constants.ReadWriteTimeout))
	n, err = conn.Read(buf)
	if err != nil {
		s.logger.Error("failed to read connect request",
			zap.String("client_address", conn.RemoteAddr().String()),
			zap.Error(err))
		return
	}
	conn.SetReadDeadline(time.Time{})

	s.logger.Debug("successfully read connect request",
		zap.String("client_address", conn.RemoteAddr().String()))

	cmd, targetAddr, err := s.p.ParseConnectRequest(buf[:n])
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
	default:
		s.logger.Error("unexpected cmd in connect request",
			zap.String("client_address", conn.RemoteAddr().String()), zap.ByteString("cmd", []byte{cmd}))

		conn.Write([]byte{0x05, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
		return
	}

	s.logger.Debug("successfully parse connect request",
		zap.String("client_address", conn.RemoteAddr().String()))

	secureRemoteConn, err := s.createSecureRemoteConnToRelayServer(ctx, remoteAddr)
	if err != nil {
		conn.Write([]byte{0x05, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})

		s.logger.Error("failed to create secure connection to relay-server",
			zap.String("relay_server_address", remoteAddr),
			zap.Error(err))
		return
	}
	defer secureRemoteConn.Close()

	s.logger.Info("successfully create secure remote connection to relay-server",
		zap.String("relay_server_address", remoteAddr))

	// вызов обработчика
	switch cmd {
	case 0x01:
		s.tcpAssociateHandler.HandleTCPAssociateConn(ctx, targetAddr, secureRemoteConn, conn)

	case 0x03:
		s.udpAssociateHandler.HandleUDPAssociateConn(ctx, secureRemoteConn, conn)
	}
}

func (s *SOCKS5Listener) handleAuthMethod(ctx context.Context, conn net.Conn, method byte, buf []byte) error {
	switch method {
	case 0x00: // no auth
		conn.SetWriteDeadline(time.Now().Add(constants.ReadWriteTimeout))
		if _, err := conn.Write([]byte{0x05, 0x00}); err != nil {
			s.logger.Error("failed to write no-auth response", zap.Error(err))
			return err
		}
		conn.SetWriteDeadline(time.Time{})

	case 0x02: // auth
		conn.SetWriteDeadline(time.Now().Add(constants.ReadWriteTimeout))
		if _, err := conn.Write([]byte{0x05, 0x02}); err != nil {
			s.logger.Error("failed to write auth method response", zap.Error(err))
			//s.clearWriteDeadline(conn)
			return err
		}
		conn.SetWriteDeadline(time.Time{})

		// read auth request
		//s.setReadDeadline(conn, constants.ReadWriteTimeout)
		conn.SetReadDeadline(time.Now().Add(constants.ReadWriteTimeout))
		n, err := conn.Read(buf)
		if err != nil {
			s.logger.Error("failed to read auth request", zap.Error(err))
			conn.SetReadDeadline(time.Time{})
			return err
		}
		conn.SetReadDeadline(time.Time{})

		user, err := s.p.ParseAuthRequest(buf[:n])
		if err != nil {
			conn.Write([]byte{0x01, 0x01})
			s.logger.Error("invalid auth request", zap.Error(err))
			return err
		}

		if err = s.storage.CheckUser(ctx, user); err != nil {
			conn.Write([]byte{0x01, 0x01})
			s.logger.Error("user authentication failed", zap.Error(err))
			return err
		}

		// send auth success
		//s.setWriteDeadline(conn, constants.ReadWriteTimeout)
		conn.SetWriteDeadline(time.Now().Add(constants.ReadWriteTimeout))
		if _, err := conn.Write([]byte{0x01, 0x00}); err != nil {
			s.logger.Error("failed to write auth success response", zap.Error(err))
			conn.SetWriteDeadline(time.Time{})
			return err
		}
		conn.SetWriteDeadline(time.Time{})

	default:
		conn.Write([]byte{0x05, 0xFF})
		return errors.New("unsupported auth method")
	}

	return nil
}

func (s *SOCKS5Listener) createSecureRemoteConnToRelayServer(ctx context.Context, remoteAddr string) (*crypto.SecureConn, error) {
	// подключение к relay-серверу
	remoteConn, err := createRemoteConnectionToRelayServer(ctx, remoteAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to create remote connection: %v", err)
	}

	// генерируем ключ
	key, err := crypto.GenerateSharedSecret(ctx, remoteConn, true)
	if err != nil {
		remoteConn.Close()
		return nil, fmt.Errorf("failed to generate shared secret: %v", err)
	}

	// создаем защищенное подключение
	secureRemoteConn, err := crypto.NewSecureConn(remoteConn, key)
	if err != nil {
		remoteConn.Close()
		return nil, fmt.Errorf("failed to create secure connection to relay-server: %v", err)
	}

	return secureRemoteConn, nil
}
