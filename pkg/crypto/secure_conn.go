package crypto

import (
	"crypto/cipher"
	"crypto/rand"
	"io"
	"net"

	"golang.org/x/crypto/chacha20poly1305"
)

type SecureConn struct {
	conn net.Conn
	aead cipher.AEAD
	key  []byte
}

func NewSecureConn(conn net.Conn, key []byte) (*SecureConn, error) {
	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, err
	}

	return &SecureConn{
		conn: conn,
		aead: aead,
		key:  key,
	}, nil
}

func (s *SecureConn) Read(p []byte) (int, error) {
	nonce := make([]byte, chacha20poly1305.NonceSizeX)
	if _, err := io.ReadFull(s.conn, nonce); err != nil {
		return 0, err
	}

	lengthBuf := make([]byte, 2)
	if _, err := io.ReadFull(s.conn, lengthBuf); err != nil {
		return 0, err
	}
	length := int(lengthBuf[0])<<1 | int(lengthBuf[1])

	cipherText := make([]byte, length)
	if _, err := io.ReadFull(s.conn, cipherText); err != nil {
		return 0, err
	}

	plainText, err := s.aead.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return 0, err
	}

	return copy(p, plainText), nil
}

func (s *SecureConn) Write(p []byte) (int, error) {
	nonce := make([]byte, chacha20poly1305.NonceSizeX)
	if _, err := rand.Read(nonce); err != nil {
		return 0, err
	}

	ciphertext := s.aead.Seal(nil, nonce, p, nil)

	length := len(ciphertext)
	header := []byte{
		byte(length >> 8), byte(length & 0xff),
	}

	_, err := s.conn.Write(append(append(nonce, header...), ciphertext...))
	if err != nil {
		return 0, err
	}

	return len(p), nil
}

func (s *SecureConn) Close() error {
	return s.conn.Close()
}
