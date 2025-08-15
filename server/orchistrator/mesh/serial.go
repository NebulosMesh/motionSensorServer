package mesh

import (
	"encoding/binary"
	"fmt"
	"io"

	"google.golang.org/protobuf/proto"
)

// SerialPort interface for serial communication
type SerialPort interface {
	io.ReadWriter
	Close() error
}

// SerialComm handles serial communication with framing
type SerialComm struct {
	port SerialPort
}

// NewSerialComm creates a new serial communication handler
func NewSerialComm(port SerialPort) *SerialComm {
	return &SerialComm{port: port}
}

// WriteFrame writes a protobuf message with 2-byte little-endian length prefix
func (s *SerialComm) WriteFrame(msg *MeshMessage) error {
	// Marshal the protobuf message
	data, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Create 2-byte little-endian length header
	header := make([]byte, 2)
	binary.LittleEndian.PutUint16(header, uint16(len(data)))

	// Write header
	if _, err := s.port.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write data
	if _, err := s.port.Write(data); err != nil {
		return fmt.Errorf("failed to write data: %w", err)
	}

	return nil
}

// ReadFrame reads a protobuf message with 2-byte little-endian length prefix
func (s *SerialComm) ReadFrame() (*MeshMessage, error) {
	// Read 2-byte header
	header := make([]byte, 2)
	if _, err := io.ReadFull(s.port, header); err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	// Parse length
	length := binary.LittleEndian.Uint16(header)
	if length == 0 {
		return nil, fmt.Errorf("invalid frame length: 0")
	}

	// Read data
	data := make([]byte, length)
	if _, err := io.ReadFull(s.port, data); err != nil {
		return nil, fmt.Errorf("failed to read data: %w", err)
	}

	// Unmarshal protobuf message
	var msg MeshMessage
	if err := proto.Unmarshal(data, &msg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	return &msg, nil
}

// Close closes the serial port
func (s *SerialComm) Close() error {
	return s.port.Close()
}
