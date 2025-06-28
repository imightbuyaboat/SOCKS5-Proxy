package tcp

import (
	"io"
	"net"

	"github.com/imightbuyaboat/SOCKS5-Proxy/pkg/constants"
	"github.com/imightbuyaboat/SOCKS5-Proxy/pkg/crypto"
	"go.uber.org/zap"
)

func (l *TCPAssociateListener) handleTCPRelay(conn net.Conn) {
	defer conn.Close()

	// устанавливаем защищенное соединение
	secureConn, err := crypto.NewSecureConn(conn, l.config.Key)
	if err != nil {
		l.logger.Error("failed to create secure connection",
			zap.String("address", conn.RemoteAddr().String()),
			zap.Error(err))
		return
	}

	buf := make([]byte, constants.BLOCK_SIZE)

	// читаем длину адреса
	n, err := secureConn.Read(buf)
	if err != nil {
		l.logger.Error("failed to read length of target address",
			zap.String("address", conn.RemoteAddr().String()),
			zap.Error(err))
		return
	}
	if n == 0 {
		l.logger.Error("empty target address length",
			zap.String("address", conn.RemoteAddr().String()))
		return
	}

	length := int(buf[0])

	// читаем адрес
	n, err = secureConn.Read(buf)
	if err != nil {
		l.logger.Error("failed to read target address",
			zap.String("address", conn.RemoteAddr().String()),
			zap.Error(err))
		return
	}
	if n == 0 {
		l.logger.Error("empty target address",
			zap.String("address", conn.RemoteAddr().String()))
		return
	}

	targetAddr := buf[:length]

	l.logger.Info("successfully read target address",
		zap.Int("length", length),
		zap.String("target_address", string(targetAddr)))

	// устанавливаем соединение с целевым адресом
	remoteConn, err := createRemoteTCPConnection(string(targetAddr))
	if err != nil {
		l.logger.Error("failed to create remote connection",
			zap.String("address", conn.RemoteAddr().String()),
			zap.String("target_address", string(targetAddr)),
			zap.Error(err))
		return
	}
	defer remoteConn.Close()

	l.logger.Info("successfully create connection to target address",
		zap.String("target_address", string(targetAddr)))

	go io.Copy(remoteConn, secureConn)
	io.Copy(secureConn, remoteConn)
}
