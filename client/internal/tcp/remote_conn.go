package tcp

import (
	"net"

	"github.com/imightbuyaboat/SOCKS5-Proxy/pkg/crypto"
)

func CreateRemoteTCPConnection(remoteTCPAddress, targetAddr, key string) (*crypto.SecureConn, error) {
	remoteConn, err := net.Dial("tcp", remoteTCPAddress)
	if err != nil {
		return nil, err
	}

	secureRemoteConn, err := crypto.NewSecureConn(remoteConn, []byte(key))
	if err != nil {
		return nil, err
	}

	addrBytes := []byte(targetAddr)
	length := byte(len(addrBytes))

	_, err = secureRemoteConn.Write([]byte{length})
	if err != nil {
		return nil, err
	}

	_, err = secureRemoteConn.Write(addrBytes)
	if err != nil {
		return nil, err
	}

	return secureRemoteConn, nil
}
