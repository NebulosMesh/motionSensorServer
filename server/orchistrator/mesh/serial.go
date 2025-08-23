package mesh

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"

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
	log.Printf("[SERIAL_TX] Preparing to send message - Type: %d, DataType: %d, Origin: %x, Target: %x, HopCount: %d, DataLen: %d", 
		msg.MessageType, msg.DataType, msg.OriginMacAddress, msg.TargetMacAddress, msg.HopCount, len(msg.Data))
	
	if len(msg.Data) > 0 {
		log.Printf("[SERIAL_TX] Message data: %x", msg.Data)
	}

	// Marshal the protobuf message
	data, err := proto.Marshal(msg)
	if err != nil {
		log.Printf("[SERIAL_TX] Failed to marshal message: %v", err)
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	log.Printf("[SERIAL_TX] Marshaled protobuf data (%d bytes): %x", len(data), data)

	// Create 2-byte little-endian length header
	header := make([]byte, 2)
	binary.LittleEndian.PutUint16(header, uint16(len(data)))

	log.Printf("[SERIAL_TX] Header to send: %02x %02x (length: %d)", header[0], header[1], len(data))

	// Write header
	if _, err := s.port.Write(header); err != nil {
		log.Printf("[SERIAL_TX] Failed to write header: %v", err)
		return fmt.Errorf("failed to write header: %w", err)
	}

	log.Printf("[SERIAL_TX] Header sent successfully")

	// Write data
	if _, err := s.port.Write(data); err != nil {
		log.Printf("[SERIAL_TX] Failed to write data: %v", err)
		return fmt.Errorf("failed to write data: %w", err)
	}

	log.Printf("[SERIAL_TX] Data sent successfully - Total frame size: %d bytes (2-byte header + %d data bytes)", 
		len(header)+len(data), len(data))

	return nil
}

// ReadFrame reads a protobuf message with 2-byte little-endian length prefix
func (s *SerialComm) ReadFrame() (*MeshMessage, error) {
	// Read 2-byte header
	header := make([]byte, 2)
	if _, err := io.ReadFull(s.port, header); err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	log.Printf("[SERIAL_RX] Header received: %02x %02x", header[0], header[1])

	// Parse length
	length := binary.LittleEndian.Uint16(header)
	if length == 0 {
		return nil, fmt.Errorf("invalid frame length: 0 (header bytes: %02x %02x)", header[0], header[1])
	}

	// Validate reasonable frame length (prevent memory exhaustion)
	if length > 4096 {
		return nil, fmt.Errorf("frame length too large: %d (header bytes: %02x %02x)", length, header[0], header[1])
	}

	log.Printf("[SERIAL_RX] Frame length: %d bytes", length)

	// Read data
	data := make([]byte, length)
	if _, err := io.ReadFull(s.port, data); err != nil {
		return nil, fmt.Errorf("failed to read data: %w", err)
	}

	log.Printf("[SERIAL_RX] Raw data received (%d bytes): %x", len(data), data)

	// Unmarshal protobuf message
	var msg MeshMessage
	if err := proto.Unmarshal(data, &msg); err != nil {
		log.Printf("[SERIAL_RX] Failed to unmarshal protobuf data: %x", data)
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	log.Printf("[SERIAL_RX] Successfully parsed message - Type: %d, DataType: %d, Origin: %x, Target: %x, HopCount: %d, DataLen: %d", 
		msg.MessageType, msg.DataType, msg.OriginMacAddress, msg.TargetMacAddress, msg.HopCount, len(msg.Data))
	
	if len(msg.Data) > 0 {
		log.Printf("[SERIAL_RX] Message data: %x", msg.Data)
	}

	return &msg, nil
}

// Close closes the serial port
func (s *SerialComm) Close() error {
	return s.port.Close()
}
