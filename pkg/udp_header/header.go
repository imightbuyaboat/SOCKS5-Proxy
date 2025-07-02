package udp_header

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
)

type Socks5UDPHeader struct {
	atyp    byte
	dstAddr []byte
	dstPort []byte
}

func (h *Socks5UDPHeader) Bytes() []byte {
	var header []byte
	header = append(header, 0x00, 0x00, 0x00)

	if h.atyp == 0x01 {
		header = append(header, 0x01)
		header = append(header, h.dstAddr...)
	} else {
		header = append(header, 0x03)
		header = append(header, byte(len(h.dstAddr)))
		header = append(header, h.dstAddr...)
	}

	header = append(header, h.dstPort...)

	return header
}

func (h *Socks5UDPHeader) DST() string {
	var host string

	switch h.atyp {
	case 0x01:
		host = net.IP(h.dstAddr).String()
	case 0x03:
		host = string(h.dstAddr)
	}

	port := binary.BigEndian.Uint16(h.dstPort)

	return net.JoinHostPort(host, strconv.Itoa(int(port)))
}

func BuildSocks5UDPHeader(addr string) (*Socks5UDPHeader, error) {
	var header Socks5UDPHeader

	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}

	port, _ := strconv.Atoi(portStr)
	if port < 0 || port > 65535 {
		return nil, fmt.Errorf("invalid port")
	}

	ip := net.ParseIP(host).To4()
	if ip != nil {
		// IPv4
		header.atyp = 0x01
		header.dstAddr = net.ParseIP(host)[12:]

	} else {
		// домен
		header.atyp = 0x03
		header.dstAddr = []byte(host)
	}

	portBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(portBytes, uint16(port))
	header.dstPort = portBytes

	return &header, nil
}

func ParseUDPPacket(packet []byte) (*Socks5UDPHeader, []byte, error) {
	if len(packet) < 4 {
		return nil, nil, fmt.Errorf("packet too short")
	}

	var header Socks5UDPHeader
	var payload []byte
	header.atyp = packet[3]

	switch header.atyp {
	case 0x01:
		// IPv4
		if len(packet) < 10 {
			return nil, nil, fmt.Errorf("invalid IPv4 address")
		}
		header.dstAddr = packet[4:8]
		header.dstPort = packet[8:10]

		if 10 > len(packet) {
			return nil, nil, fmt.Errorf("invalid packet structure")
		}
		payload = packet[10:]

	case 0x03:
		// домен
		if len(packet) < 5 {
			return nil, nil, fmt.Errorf("invalid domain length")
		}
		domainLen := int(packet[4])
		if len(packet) < 7+domainLen {
			return nil, nil, fmt.Errorf("invalid domain address")
		}
		header.dstAddr = packet[5 : 5+domainLen]
		header.dstPort = packet[5+domainLen : 7+domainLen]

		if 7+domainLen > len(packet) {
			return nil, nil, fmt.Errorf("invalid packet structure")
		}
		payload = packet[7+domainLen:]

	default:
		return nil, nil, fmt.Errorf("unsupported address type: %d", packet[3])
	}

	return &header, payload, nil
}
