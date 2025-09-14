package tcp

import (
	"io"
	"net"

	"github.com/imightbuyaboat/SOCKS5-Proxy/pkg/block"
	"github.com/imightbuyaboat/SOCKS5-Proxy/pkg/crypto"
	"go.uber.org/zap"
)

func (l *TCPAssociateListener) handleTCPRelay(conn net.Conn) {
	defer conn.Close()

	// генерируем разделяемый секрет
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

	buf := make([]byte, block.BLOCK_SIZE)

	// читаем длину целевого адреса
	n, err := secureConn.Read(buf)
	if err != nil {
		l.logger.Error("failed to read length of target address",
			zap.String("socks5_server_address", conn.RemoteAddr().String()),
			zap.Error(err))
		return
	}
	if n == 0 {
		l.logger.Error("empty target address",
			zap.String("socks5_server_address", conn.RemoteAddr().String()))
		return
	}

	length := int(buf[0])

	// читаем адрес
	n, err = secureConn.Read(buf)
	if err != nil {
		l.logger.Error("failed to read target address",
			zap.String("socks5_server_address", conn.RemoteAddr().String()),
			zap.Error(err))
		return
	}
	if n == 0 {
		l.logger.Error("empty target address",
			zap.String("socks5_server_address", conn.RemoteAddr().String()))
		return
	}

	targetAddr := buf[:length]

	l.logger.Info("successfully read target address",
		zap.Int("length", length),
		zap.String("target_address", string(targetAddr)))

	// устанавливаем соединение с целевым сервером
	remoteConn, err := createRemoteTCPConnection(string(targetAddr))
	if err != nil {
		l.logger.Error("failed to create connection",
			zap.String("socks5_server_address", conn.RemoteAddr().String()),
			zap.String("target_address", string(targetAddr)),
			zap.Error(err))
		return
	}
	defer remoteConn.Close()

	l.logger.Info("successfully create connection to target server",
		zap.String("target_address", string(targetAddr)))

	go io.Copy(remoteConn, secureConn)
	io.Copy(secureConn, remoteConn)
}
