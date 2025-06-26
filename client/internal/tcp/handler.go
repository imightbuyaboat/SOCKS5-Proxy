package tcp

import (
	"io"
	"net"

	"github.com/imightbuyaboat/SOCKS5-Proxy/client/internal/socks5"
	"go.uber.org/zap"
)

func (s *SOCKS5ListenerTCP) handleConnection(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 512)

	// handshake request
	n, err := conn.Read(buf)
	if err != nil {
		s.logger.Error("failed to read handshake request",
			zap.String("address", conn.RemoteAddr().String()),
			zap.Error(err))
		return
	}
	if n == 0 {
		s.logger.Error("empty handshake request",
			zap.String("address", conn.RemoteAddr().String()))
		return
	}

	if err = socks5.ParseHandshake(buf); err != nil {
		s.logger.Error("failed to parse handshake request",
			zap.String("address", conn.RemoteAddr().String()),
			zap.Error(err))
		return
	}

	conn.Write([]byte{0x05, 0x00})

	// connection request
	n, err = conn.Read(buf)
	if err != nil {
		s.logger.Error("failed to read connect request",
			zap.String("address", conn.RemoteAddr().String()),
			zap.Error(err))
		return
	}
	if n == 0 {
		s.logger.Error("empty connect request",
			zap.String("address", conn.RemoteAddr().String()))
		return
	}

	targetAddr, err := socks5.ParseConnectRequest(buf)
	if err != nil {
		sendReply(conn, 0x01)
		s.logger.Error("failed to parse connect request",
			zap.String("address", conn.RemoteAddr().String()),
			zap.Error(err))
		return
	}

	// подключение к прокси-серверу
	remoteConn, err := createRemoteTCPConnection(s.config.RemoteTCPAddress, targetAddr, s.config.Key)
	if err != nil {
		sendReply(conn, 0x01)
		s.logger.Error("failed to create remote connection",
			zap.String("address", conn.RemoteAddr().String()),
			zap.String("remote_address", s.config.RemoteTCPAddress),
			zap.Error(err))
		return
	}
	defer remoteConn.Close()

	s.logger.Info("successfully create remote connection",
		zap.String("remote_tcp_address", s.config.RemoteTCPAddress))

	sendReply(conn, 0x00) // успешное подключение к прокси-серверу

	go io.Copy(remoteConn, conn)
	io.Copy(conn, remoteConn)
}

func sendReply(conn net.Conn, rep byte) {
	reply := []byte{0x05, rep, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	conn.Write(reply)
}
