package crypto

import (
	"context"
	"crypto/hkdf"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"net"
	"time"

	"github.com/imightbuyaboat/SOCKS5-Proxy/pkg/constants"
	"golang.org/x/crypto/curve25519"
)

func GenerateSharedSecret(ctx context.Context, remoteConn net.Conn, initiator bool) ([]byte, error) {
	var localPriv [32]byte
	_, err := rand.Read(localPriv[:])
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %v", err)
	}

	localPub, err := curve25519.X25519(localPriv[:], curve25519.Basepoint)
	if err != nil {
		return nil, fmt.Errorf("failed to generate public key: %v", err)
	}

	var remotePub [32]byte

	if initiator {
		remoteConn.SetWriteDeadline(time.Now().Add(constants.ReadWriteTimeout))
		if _, err = remoteConn.Write(localPub); err != nil {
			return nil, fmt.Errorf("failed to send public key to %s: %v", remoteConn.RemoteAddr().String(), err)
		}
		remoteConn.SetWriteDeadline(time.Time{})

		remoteConn.SetReadDeadline(time.Now().Add(constants.ReadWriteTimeout))
		if _, err = remoteConn.Read(remotePub[:]); err != nil {
			return nil, fmt.Errorf("failed to recieve public key from %s: %v", remoteConn.RemoteAddr().String(), err)
		}
		remoteConn.SetReadDeadline(time.Time{})
	} else {
		remoteConn.SetReadDeadline(time.Now().Add(constants.ReadWriteTimeout))
		if _, err = remoteConn.Read(remotePub[:]); err != nil {
			return nil, fmt.Errorf("failed to recieve public key from %s: %v", remoteConn.RemoteAddr().String(), err)
		}
		remoteConn.SetReadDeadline(time.Time{})

		remoteConn.SetWriteDeadline(time.Now().Add(constants.ReadWriteTimeout))
		if _, err = remoteConn.Write(localPub); err != nil {
			return nil, fmt.Errorf("failed to send public key to %s: %v", remoteConn.RemoteAddr().String(), err)
		}
		remoteConn.SetWriteDeadline(time.Time{})
	}

	sharedSecret, err := curve25519.X25519(localPriv[:], remotePub[:])
	if err != nil {
		return nil, fmt.Errorf("failed to generate shared secret: %v", err)
	}

	key, err := hkdf.Key(sha256.New, sharedSecret, nil, "", 32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key from shared secret: %v", err)
	}

	return key, nil
}
