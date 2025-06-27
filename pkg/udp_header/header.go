package udp_header

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
)

type Socks5UDPHeader struct {
	atyp    byte
	dstAddr string
	dstPort uint16
}

func (h *Socks5UDPHeader) ToBytes() []byte {
	var header []byte
	header = append(header, 0x00, 0x00, 0x00)

	portBytes := make([]byte, 2)

	if h.atyp == 0x01 {
		header = append(header, 0x01)
		header = append(header, net.ParseIP(h.dstAddr)...)
		binary.BigEndian.PutUint16(portBytes, h.dstPort)
		header = append(header, portBytes...)
	} else {
		header = append(header, 0x03)
		header = append(header, byte(len(h.dstAddr)))
		header = append(header, []byte(h.dstAddr)...)
		binary.BigEndian.PutUint16(portBytes, h.dstPort)
		header = append(header, portBytes...)
	}

	return header
}

func (h *Socks5UDPHeader) DST() string {
	return net.JoinHostPort(h.dstAddr, strconv.Itoa(int(h.dstPort)))
}

func BuildSocks5UDPHeader(addr string) (*Socks5UDPHeader, error) {
	var header Socks5UDPHeader

	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}

	port, _ := strconv.Atoi(portStr)

	ip := net.ParseIP(host).To4()
	if ip != nil {
		// IPv4
		header.atyp = 0x01
		header.dstAddr = ip.String()
		header.dstPort = uint16(port)

	} else {
		// домен
		header.atyp = 0x03
		header.dstAddr = host
		header.dstPort = uint16(port)
	}

	return &header, nil
}

func ParseUDPPacket(packet []byte) (*Socks5UDPHeader, []byte, error) {
	if len(packet) < 4 {
		return nil, nil, fmt.Errorf("packet too short")
	}

	var header Socks5UDPHeader
	header.atyp = packet[3]

	var i int

	switch header.atyp {
	case 0x01:
		// IPv4
		if len(packet) < 10 {
			return nil, nil, fmt.Errorf("invalid IPv4 address")
		}
		header.dstAddr = net.IP(packet[4:8]).String()
		header.dstPort = binary.BigEndian.Uint16(packet[8:10])
		i = 10

	case 0x03:
		// домен
		if len(packet) < 5 {
			return nil, nil, fmt.Errorf("invalid domain length")
		}
		domainLen := int(packet[4])
		if len(packet) < 7+domainLen {
			return nil, nil, fmt.Errorf("invalid domain address")
		}
		header.dstAddr = string(packet[5 : 5+domainLen])
		header.dstPort = binary.BigEndian.Uint16(packet[5+domainLen : 7+domainLen])
		i = 7 + domainLen

	default:
		return nil, nil, fmt.Errorf("unsupported address type: %d", packet[5])
	}

	if i > len(packet) {
		return nil, nil, fmt.Errorf("invalid packet structure")
	}
	payload := packet[i:]

	return &header, payload, nil
}
