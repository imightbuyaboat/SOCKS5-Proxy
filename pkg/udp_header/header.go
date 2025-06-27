package udp_header

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
)

func BuildSocks5UDPHeader(srcAddr, dstAddr string) ([]byte, error) {
	var header []byte
	header = append(header, 0x00, 0x00, 0x00)

	var err error

	// src
	header, err = addAddressToHeader(header, srcAddr)
	if err != nil {
		return nil, err
	}

	// dst
	header, err = addAddressToHeader(header, dstAddr)
	if err != nil {
		return nil, err
	}

	return header, nil
}

func addAddressToHeader(header []byte, addr string) ([]byte, error) {
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}

	port, _ := strconv.Atoi(portStr)

	ip := net.ParseIP(host).To4()
	if ip != nil {
		// IPv4
		header = append(header, 0x01)
		header = append(header, ip...)

		portBytes := make([]byte, 2)
		binary.BigEndian.PutUint16(portBytes, uint16(port))
		header = append(header, portBytes...)

	} else {
		// домен
		header = append(header, 0x03)
		header = append(header, byte(len(host)))
		header = append(header, []byte(host)...)
		portBytes := make([]byte, 2)
		binary.BigEndian.PutUint16(portBytes, uint16(port))
		header = append(header, portBytes...)
	}

	return header, nil
}

func ParseSocks5UDPHeader(packet []byte) (string, string, []byte, error) {
	if len(packet) < 3 {
		return "", "", nil, fmt.Errorf("packet too short")
	}

	i := 3

	// dst
	dstAddr, n, err := parseAddress(packet[i:])
	if err != nil {
		return "", "", nil, err
	}
	i += n

	// src
	srcAddr, n, err := parseAddress(packet[i:])
	if err != nil {
		return "", "", nil, err
	}
	i += n

	if i > len(packet) {
		return "", "", nil, fmt.Errorf("invalid packet structure")
	}
	payload := packet[i:]

	return srcAddr, dstAddr, payload, nil
}

func parseAddress(data []byte) (string, int, error) {
	if len(data) < 1 {
		return "", 0, fmt.Errorf("address too short")
	}

	atyp := data[0]
	i := 1

	switch atyp {
	case 0x01: // IPv4
		if len(data) < i+6 {
			return "", 0, fmt.Errorf("invalid IPv4 address")
		}
		ip := net.IP(data[i : i+4]).String()
		i += 4
		port := binary.BigEndian.Uint16(data[i : i+2])
		i += 2
		return net.JoinHostPort(ip, strconv.Itoa(int(port))), i, nil

	case 0x03: // domain
		if len(data) < i+1 {
			return "", 0, fmt.Errorf("invalid domain length")
		}
		domainLen := int(data[i])
		i++
		if len(data) < i+domainLen+2 {
			return "", 0, fmt.Errorf("invalid domain address")
		}
		host := string(data[i : i+domainLen])
		i += domainLen
		port := binary.BigEndian.Uint16(data[i : i+2])
		i += 2
		return net.JoinHostPort(host, strconv.Itoa(int(port))), i, nil

	default:
		return "", 0, fmt.Errorf("unsupported address type: %d", atyp)
	}
}
