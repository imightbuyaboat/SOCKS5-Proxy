package config

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"

	"golang.org/x/crypto/chacha20poly1305"
)

type Config struct {
	ListenAddress    string `json:"listen_address"`
	RemoteTCPAddress string `json:"remote_tcp_address"`
	Key              []byte
}

func LoadConfig() (*Config, error) {
	file, err := os.OpenFile("config.json", os.O_RDONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %v", err)
	}
	defer file.Close()

	var config Config
	if err = json.NewDecoder(file).Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %v", err)
	}

	encoded := os.Getenv("PRIVATE_KEY")

	key, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 key: %v", err)
	}

	config.Key = key

	if err = config.validateConfig(); err != nil {
		return nil, fmt.Errorf("failed to validate config: %v", err)
	}

	return &config, nil
}

func (c *Config) validateConfig() error {
	host, portStr, err := net.SplitHostPort(c.ListenAddress)
	if err != nil {
		return err
	}
	if err = validateHost(host); err != nil {
		return err
	}
	if err = validatePort(portStr); err != nil {
		return err
	}

	host, portStr, err = net.SplitHostPort(c.RemoteTCPAddress)
	if err != nil {
		return err
	}
	if err = validateHost(host); err != nil {
		return err
	}
	if err = validatePort(portStr); err != nil {
		return err
	}

	if len(c.Key) != chacha20poly1305.KeySize {
		return fmt.Errorf("incorrect length of key")
	}

	return nil
}

func validateHost(host string) error {
	ip := net.ParseIP(host)
	if ip == nil {
		return fmt.Errorf("invalid ip")
	}
	return nil
}

func validatePort(portStr string) error {
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("invalid port: %v", err)
	}

	if port < 0 || port > 65535 {
		return fmt.Errorf("invalid port")
	}
	return nil
}
