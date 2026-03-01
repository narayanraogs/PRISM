package driver

import (
	"fmt"
	"net"
	"strconv"
	"time"
)

type ArbitraryDevice struct {
	IPAddress string
	Port      int
	Timeout   time.Duration
}

func NewArbitraryDevice(ip string, port int) *ArbitraryDevice {
	return &ArbitraryDevice{
		IPAddress: ip,
		Port:      port,
		Timeout:   10 * time.Second,
	}
}

func (d *ArbitraryDevice) SendCommand(command string, read bool) (string, error) {
	address := net.JoinHostPort(d.IPAddress, strconv.Itoa(d.Port))
	conn, err := net.DialTimeout("tcp", address, d.Timeout)
	if err != nil {
		return "", fmt.Errorf("failed to connect to device: %v", err)
	}
	defer conn.Close()

	// Send command
	conn.SetWriteDeadline(time.Now().Add(d.Timeout))
	command = command + "\n"
	_, err = conn.Write([]byte(command))
	if err != nil {
		return "", fmt.Errorf("failed to send command: %v", err)
	}

	// Read response
	conn.SetReadDeadline(time.Now().Add(d.Timeout))
	var readValue string
	if read {
		time.Sleep(2 * time.Second)
		buffer := make([]byte, 1024)
		n, err := conn.Read(buffer)
		if err != nil {
			return "", fmt.Errorf("failed to read response: %v", err)
		}
		readValue = string(buffer[:n])
	}
	return readValue, nil
}
