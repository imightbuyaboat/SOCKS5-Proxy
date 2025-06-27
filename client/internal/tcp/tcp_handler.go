package tcp

import (
	"io"
	"net"

	"github.com/imightbuyaboat/SOCKS5-Proxy/pkg/crypto"
	"go.uber.org/zap"
)

func HandleTCPAssociateConn(targetAddr string, remoteConn *crypto.SecureConn, conn net.Conn, logger *zap.Logger) {
	// отправляем целевой адрес и его длину
	addrBytes := []byte(targetAddr)
	length := byte(len(addrBytes))

	_, err := remoteConn.Write([]byte{length})
	if err != nil {
		logger.Error("failed to write length of target address",
			zap.String("target_address", targetAddr),
			zap.Int("length", int(length)),
			zap.Error(err))
		return
	}

	_, err = remoteConn.Write(addrBytes)
	if err != nil {
		logger.Error("failed to write target address",
			zap.String("target_address", targetAddr),
			zap.Error(err))
		return
	}

	// успешное подключение к прокси-серверу
	conn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})

	go io.Copy(remoteConn, conn)
	io.Copy(conn, remoteConn)
}
