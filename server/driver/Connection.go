package driver

import (
	"context"
	"encoding/base64"
	"errors"
	"net"
	"prismServer/logger"
	"prismServer/utils"
	"strconv"
	"strings"
	"sync"
	"time"
)

type managedConnection struct {
	connection net.Conn
	lastUsed   time.Time
}

type connectionManager struct {
	connections map[string]*managedConnection
	mutex       sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
}

var connections connectionManager

func init() {
	connections.connections = make(map[string]*managedConnection)
	connections.ctx, connections.cancel = context.WithCancel(context.Background())
	go cleanUp()
}

type instrument struct {
	IPAddress       string
	PortNo          int
	AlternatePortNo int
	ReadPortNo      int
	DopplerPortNo   int
	Timeout         int
	seperator       string
	teminator       string
	opcAllowed      bool
	scpi            bool
	conn            *managedConnection
}

func ShutdownAllConnections() {
	connections.mutex.Lock()
	defer connections.mutex.Unlock()
	for k, v := range connections.connections {
		v.connection.Close()
		delete(connections.connections, k)
	}
	connections.cancel()
}

func (lan *instrument) Connect(port string) bool {
	d := net.Dialer{Timeout: time.Duration(lan.Timeout) * time.Millisecond}
	var connString = ""
	switch strings.ToLower(port) {
	case "alternate":
		connString = lan.IPAddress + ":" + strconv.FormatInt(int64(lan.AlternatePortNo), 10)
	case "read":
		connString = lan.IPAddress + ":" + strconv.FormatInt(int64(lan.ReadPortNo), 10)
	case "doppler":
		connString = lan.IPAddress + ":" + strconv.FormatInt(int64(lan.DopplerPortNo), 10)
	default:
		connString = lan.IPAddress + ":" + strconv.FormatInt(int64(lan.PortNo), 10)
	}
	connections.mutex.RLock()
	managed, ok := connections.connections[connString]
	connections.mutex.RUnlock()
	if !ok {
		logger.Log.Debug("Driver Connecting", "address", connString)
		conn, err := d.Dial("tcp", connString)
		if err != nil {
			logger.Log.Error(err.Error())
			return false
		}
		managed = &managedConnection{connection: conn, lastUsed: time.Now()}
		connections.mutex.Lock()
		connections.connections[connString] = managed
		connections.mutex.Unlock()
	}
	managed.lastUsed = time.Now()
	lan.conn = managed
	return true
}

func (lan *instrument) communicateWithDevice(command string, delay time.Duration, toBeRead bool, readBinary bool) (string, error) {
	ok := lan.write(command)
	if !ok {
		return "", errors.New("write timeout")
	}
	if delay != 0 {
		time.Sleep(delay * time.Millisecond)
	}
	var data = ""
	if toBeRead {
		data, ok = lan.read(readBinary)
		if !ok {
			return "", errors.New("read timeout")
		}
	}
	return data, nil
}

func (lan *instrument) communicateWithDeviceBinary(packet []byte, delay time.Duration) (string, error) {
	ok := lan.writeBinary(packet)
	if !ok {
		return "", errors.New("write timeout")
	}
	if delay != 0 {
		time.Sleep(delay * time.Millisecond)
	}
	var data = ""
	data, ok = lan.read(true)
	if !ok {
		return "", errors.New("read timeout")
	}

	return data, nil
}

func (lan *instrument) write(value string) bool {
	err := lan.conn.connection.SetWriteDeadline(time.Now().Add(time.Duration(lan.Timeout) * time.Millisecond))
	if err != nil {
		return false
	}
	logger.Log.Debug("Driver TX", "IP", lan.IPAddress, "command", strings.TrimSpace(value))
	_, err = lan.conn.connection.Write([]byte(value))
	if err != nil {
		return false
	}
	return true
}

func (lan *instrument) writeBinary(data []byte) bool {
	err := lan.conn.connection.SetWriteDeadline(time.Now().Add(time.Duration(lan.Timeout) * time.Millisecond))
	if err != nil {
		return false
	}
	logger.Log.Debug("Driver TX Binary", "IP", lan.IPAddress, "bytes", len(data))
	_, err = lan.conn.connection.Write(data)
	if err != nil {
		return false
	}
	return true
}

func (lan *instrument) read(binary bool) (string, bool) {
	var tbr = make([]byte, 0)

	packet := make([]byte, 1000000)
	var n = len(packet)
	var readValue string
	if !binary {
		err := lan.conn.connection.SetReadDeadline(time.Now().Add(time.Duration(lan.Timeout) * time.Millisecond))
		if err != nil {
			return "Cannot set read timeout", false
		}
		n, err = lan.conn.connection.Read(packet)
		if err != nil {
			return "", false
		}
		tbr = append(tbr, packet[:n]...)
		readValue = string(tbr)
	} else {
		err := lan.conn.connection.SetReadDeadline(time.Now().Add(2000 * time.Millisecond))
		if err != nil {
			return "Cannot set read timeout", false
		}
		for {
			n, err = lan.conn.connection.Read(packet)
			if err != nil {
				if len(tbr) == 0 {
					logger.Log.Error(err.Error())
				}
				break
			}
			tbr = append(tbr, packet[:n]...)
			packet = make([]byte, 100000)
		}
		readValue = base64.StdEncoding.EncodeToString(tbr)
	}
	if len(tbr) == 0 {
		return "", false
	}
	readValue = strings.TrimSpace(readValue)
	if binary {
		logger.Log.Debug("Driver RX Binary", "IP", lan.IPAddress, "bytes", len(tbr))
	} else {
		logger.Log.Debug("Driver RX", "IP", lan.IPAddress, "response", readValue)
	}
	return readValue, true
}

func (lan *instrument) Configure(sep string, term string, scpi bool, opcPresent bool) {
	lan.seperator = sep
	lan.teminator = term
	lan.opcAllowed = opcPresent
	lan.scpi = scpi
}

func (lan *instrument) Communicate(cmds []utils.Command) ([]string, error) {
	if lan.scpi {
		return lan.communicateSCPI(cmds)
	}
	return lan.communicateBinary(cmds)
}

func (lan *instrument) communicateSCPI(cmdList []utils.Command) ([]string, error) {
	var retVal = make([]string, 0)
	for _, cmd := range cmdList {
		command := cmd.Command
		if cmd.Argument {
			command = command + lan.seperator + cmd.ArgumentValue
		}

		if lan.opcAllowed && !cmd.Read {
			command = command + ";*opc?"
		}
		var toBeRead = cmd.Read || lan.opcAllowed
		command = command + lan.teminator
		delay := time.Duration(int64(cmd.Delay * 1000))

		data, err := lan.communicateWithDevice(command, delay, toBeRead, cmd.ReadBinary)
		if err != nil {
			logger.Log.Error(err.Error())
			return nil, err
		}
		if cmd.Read {
			retVal = append(retVal, data)
		}
	}
	lan.conn.lastUsed = time.Now()
	return retVal, nil
}

func (lan *instrument) communicateBinary(cmdList []utils.Command) ([]string, error) {
	var retVal = make([]string, 0)
	for _, cmd := range cmdList {
		command := cmd.Packet
		delay := time.Duration(cmd.Delay * 1000)

		data, err := lan.communicateWithDeviceBinary(command, delay)
		if err != nil {
			logger.Log.Error(err.Error())
			return nil, err
		}
		retVal = append(retVal, data)
	}
	lan.conn.lastUsed = time.Now()
	return retVal, nil
}

func cleanUp() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-connections.ctx.Done():
			return
		case <-ticker.C:
			connections.mutex.Lock()
			for k, v := range connections.connections {
				if time.Since(v.lastUsed) > 30*time.Second {
					v.connection.Close()
					delete(connections.connections, k)
				}
			}
			connections.mutex.Unlock()
		}
	}
}
