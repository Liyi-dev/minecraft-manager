package rcon

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

const (
	loginRequest  int32 = 3
	commandRequest int32 = 2
	commandResponse int32 = 0
	maxPacketSize = 4096
)

var (
	ErrNotConnected = errors.New("rcon: not connected")
	ErrAuthFailed   = errors.New("rcon: authentication failed")
	ErrTimeout      = errors.New("rcon: command timed out")
)

type RCONClient struct {
	host     string
	port     int
	password string
	conn     net.Conn
	mu       sync.Mutex
	timeout  time.Duration
	reqID    int32
}

func New(host string, port int, password string) *RCONClient {
	return &RCONClient{
		host:     host,
		port:     port,
		password: password,
		timeout:  5 * time.Second,
	}
}

func (c *RCONClient) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		c.conn.Close()
	}

	addr := fmt.Sprintf("%s:%d", c.host, c.port)
	conn, err := net.DialTimeout("tcp", addr, c.timeout)
	if err != nil {
		return fmt.Errorf("rcon: dial failed: %w", err)
	}
	c.conn = conn

	// Authenticate
	if err := c.authenticate(); err != nil {
		c.conn.Close()
		c.conn = nil
		return err
	}

	c.reqID = 0
	return nil
}

func (c *RCONClient) authenticate() error {
	c.reqID++
	packet := buildPacket(c.reqID, loginRequest, c.password)
	if _, err := c.conn.Write(packet); err != nil {
		return fmt.Errorf("rcon: auth write failed: %w", err)
	}

	// Read response
	resp, id, err := c.readPacket()
	if err != nil {
		return fmt.Errorf("rcon: auth read failed: %w", err)
	}

	// Auth failure: server sends request ID = -1
	if id == -1 {
		return ErrAuthFailed
	}

	// Some servers echo the login request; consume it
	_ = resp
	return nil
}

// Execute sends a command and returns the response.
func (c *RCONClient) Execute(cmd string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return "", ErrNotConnected
	}

	c.reqID++
	packet := buildPacket(c.reqID, commandRequest, cmd)

	// Set write deadline
	c.conn.SetWriteDeadline(time.Now().Add(c.timeout))
	if _, err := c.conn.Write(packet); err != nil {
		// Connection broken, mark as disconnected
		c.conn.Close()
		c.conn = nil
		return "", fmt.Errorf("rcon: write failed: %w", err)
	}

	// Set read deadline
	c.conn.SetReadDeadline(time.Now().Add(c.timeout))
	resp, _, err := c.readPacket()
	if err != nil {
		c.conn.Close()
		c.conn = nil
		return "", fmt.Errorf("rcon: read failed: %w", err)
	}

	return resp, nil
}

// ExecuteWithRetry attempts the command with retries on failure.
func (c *RCONClient) ExecuteWithRetry(cmd string, retries int) (string, error) {
	var lastErr error
	for i := 0; i <= retries; i++ {
		if c.conn == nil {
			if err := c.Connect(); err != nil {
				lastErr = err
				time.Sleep(500 * time.Millisecond)
				continue
			}
		}

		result, err := c.Execute(cmd)
		if err == nil {
			return result, nil
		}
		lastErr = err
		time.Sleep(500 * time.Millisecond)
	}
	return "", lastErr
}

func (c *RCONClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		err := c.conn.Close()
		c.conn = nil
		return err
	}
	return nil
}

func (c *RCONClient) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn != nil
}

// readPacket reads a complete RCON packet from the connection.
func (c *RCONClient) readPacket() (string, int32, error) {
	// Read length (4 bytes, little-endian int32)
	lenBuf := make([]byte, 4)
	if _, err := io.ReadFull(c.conn, lenBuf); err != nil {
		return "", 0, fmt.Errorf("read length: %w", err)
	}
	length := int32(binary.LittleEndian.Uint32(lenBuf))
	if length < 10 || length > maxPacketSize {
		return "", 0, fmt.Errorf("invalid packet length: %d", length)
	}

	// Read the rest of the packet
	data := make([]byte, length)
	if _, err := io.ReadFull(c.conn, data); err != nil {
		return "", 0, fmt.Errorf("read data: %w", err)
	}

	reqID := int32(binary.LittleEndian.Uint32(data[0:4]))
	respType := int32(binary.LittleEndian.Uint32(data[4:8]))

	// Payload is null-terminated string(s), strip trailing nulls
	payload := string(data[8 : length-2])

	// If it's a login request echoed back, just return
	if respType == loginRequest {
		return payload, reqID, nil
	}

	return payload, reqID, nil
}

// buildPacket creates an RCON packet.
func buildPacket(reqID int32, packetType int32, payload string) []byte {
	payloadBytes := []byte(payload)
	// Length: 4 (reqID) + 4 (type) + len(payload) + 2 (null terminators)
	length := int32(10 + len(payloadBytes))

	buf := make([]byte, 4+length)
	binary.LittleEndian.PutUint32(buf[0:4], uint32(length))
	binary.LittleEndian.PutUint32(buf[4:8], uint32(reqID))
	binary.LittleEndian.PutUint32(buf[8:12], uint32(packetType))
	copy(buf[12:], payloadBytes)
	// Last two bytes are already zero (null terminators)

	return buf
}
