package socks5_tests

import (
	"bytes"
	"encoding/binary"
	"net"
	"testing"
	"time"
)

func buildDNSQuery() []byte {
	return append([]byte{
		0x12, 0x34, // ID
		0x01, 0x00, // Стандартный запрос
		0x00, 0x01, // QDCOUNT
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // Zeros
		0x06, 'g', 'o', 'o', 'g', 'l', 'e',
		0x03, 'c', 'o', 'm', 0x00,
		0x00, 0x01, // Type A
		0x00, 0x01, // Class IN
	}, []byte{}...)
}

func TestSOCKS5_UDP_Associate(t *testing.T) {
	const (
		socks5Addr = "127.0.0.1:1080"
		targetIP   = "8.8.8.8"
		targetPort = 53
	)

	conn, err := net.Dial("tcp", socks5Addr)
	if err != nil {
		t.Fatalf("failed to connect to SOCKS5: %v", err)
	}
	defer conn.Close()

	// handshake
	conn.Write([]byte{0x05, 0x01, 0x00})
	resp := make([]byte, 2)
	conn.Read(resp)

	// connect
	var req bytes.Buffer
	req.Write([]byte{0x05, 0x03, 0x00, 0x01})
	req.Write([]byte{0, 0, 0, 0})
	req.Write([]byte{0, 0})
	conn.Write(req.Bytes())

	// ответ прокси-клиента
	reply := make([]byte, 10)
	_, err = conn.Read(reply)
	if err != nil {
		t.Fatalf("failed to read SOCKS5 UDP associate response: %v", err)
	}

	relayIP := net.IP(reply[4:8])
	relayPort := binary.BigEndian.Uint16(reply[8:10])
	relayAddr := &net.UDPAddr{
		IP:   relayIP,
		Port: int(relayPort),
	}

	// UDP-соединение к прокси-клиенту
	localAddr := &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0}
	udpConn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		t.Fatalf("failed to listen UDP: %v", err)
	}
	defer udpConn.Close()

	// формируем пакет
	var packet bytes.Buffer
	packet.Write([]byte{0x00, 0x00, 0x00, 0x01})
	packet.Write(net.ParseIP(targetIP).To4())
	portBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(portBytes, targetPort)
	packet.Write(portBytes)
	packet.Write(buildDNSQuery())

	// отправляем пакет
	_, err = udpConn.WriteTo(packet.Bytes(), relayAddr)
	if err != nil {
		t.Fatalf("failed to send UDP packet to relay: %v", err)
	}

	// ожидание ответа
	buf := make([]byte, 512)
	udpConn.SetReadDeadline(time.Now().Add(time.Second * 3))
	n, from, err := udpConn.ReadFrom(buf)
	if err != nil {
		t.Fatalf("no response from UDP target through SOCKS5: %v", err)
	}
	if n < 12 {
		t.Fatalf("unexpected short response: %d bytes", n)
	}

	t.Logf("Received %d bytes from %s", n, from.String())
}
