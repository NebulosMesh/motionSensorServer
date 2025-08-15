package mesh

import (
	"bytes"
	"testing"
)

// MockSerialPort implements SerialPort for testing
type MockSerialPort struct {
	readBuffer  *bytes.Buffer
	writeBuffer *bytes.Buffer
}

func NewMockSerialPort() *MockSerialPort {
	return &MockSerialPort{
		readBuffer:  &bytes.Buffer{},
		writeBuffer: &bytes.Buffer{},
	}
}

func (m *MockSerialPort) Read(p []byte) (int, error) {
	return m.readBuffer.Read(p)
}

func (m *MockSerialPort) Write(p []byte) (int, error) {
	return m.writeBuffer.Write(p)
}

func (m *MockSerialPort) Close() error {
	return nil
}

func (m *MockSerialPort) AddReadData(data []byte) {
	m.readBuffer.Write(data)
}

func (m *MockSerialPort) GetWrittenData() []byte {
	return m.writeBuffer.Bytes()
}

func TestMessageBuilder(t *testing.T) {
	builder := NewMessageBuilder()

	t.Run("BuildConfigSetMessage", func(t *testing.T) {
		mac := []byte{0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF}
		msg, err := builder.BuildConfigSetMessage(mac, AdapterTypePIR)
		
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if msg.MessageType != MessageTypeAdapterData {
			t.Errorf("Expected MessageType %d, got %d", MessageTypeAdapterData, msg.MessageType)
		}

		if msg.DataType != AdapterTypeSerial {
			t.Errorf("Expected DataType %d, got %d", AdapterTypeSerial, msg.DataType)
		}

		if msg.Data[0] != OpConfigSet {
			t.Errorf("Expected opcode %02x, got %02x", OpConfigSet, msg.Data[0])
		}

		if !bytes.Equal(msg.Data[1:7], mac) {
			t.Errorf("Expected MAC %x, got %x", mac, msg.Data[1:7])
		}

		if msg.Data[7] != byte(AdapterTypePIR) {
			t.Errorf("Expected adapter type %d, got %d", AdapterTypePIR, msg.Data[7])
		}
	})

	t.Run("BuildHealthRequestMessage", func(t *testing.T) {
		msg := builder.BuildHealthRequestMessage()
		
		if msg.MessageType != MessageTypeAdapterData {
			t.Errorf("Expected MessageType %d, got %d", MessageTypeAdapterData, msg.MessageType)
		}

		if msg.DataType != AdapterTypeSerial {
			t.Errorf("Expected DataType %d, got %d", AdapterTypeSerial, msg.DataType)
		}

		if msg.Data[0] != OpHealthReq {
			t.Errorf("Expected opcode %02x, got %02x", OpHealthReq, msg.Data[0])
		}
	})

	t.Run("ParseHealthReport", func(t *testing.T) {
		// Create a mock health report message
		data := make([]byte, MaxDataLength)
		data[0] = OpHealthReport
		data[1] = byte(AdapterTypePIR)
		// MAC address
		mac := []byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66}
		copy(data[2:8], mac)
		// Uptime (little-endian)
		data[8] = 0x10  // 4112 seconds
		data[9] = 0x10
		data[10] = 0x00
		data[11] = 0x00

		msg := &MeshMessage{
			MessageType: MessageTypeAdapterData,
			DataType:    AdapterTypeSerial,
			Data:        data,
			HopCount:    2,
		}

		report, err := builder.ParseHealthReport(msg)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if !bytes.Equal(report.MAC, mac) {
			t.Errorf("Expected MAC %x, got %x", mac, report.MAC)
		}

		if report.AdapterType != AdapterTypePIR {
			t.Errorf("Expected adapter type %d, got %d", AdapterTypePIR, report.AdapterType)
		}

		if report.Uptime != 4112 {
			t.Errorf("Expected uptime 4112, got %d", report.Uptime)
		}

		if report.HopCount != 2 {
			t.Errorf("Expected hop count 2, got %d", report.HopCount)
		}
	})
}

func TestNodeRegistry(t *testing.T) {
	registry := NewNodeRegistry()

	mac1 := []byte{0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF}
	mac2 := []byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66}

	t.Run("UpdateAndGetNode", func(t *testing.T) {
		// Update node
		registry.UpdateNode(mac1, AdapterTypePIR, 1000, 1)

		// Get node
		node, exists := registry.GetNode(mac1)
		if !exists {
			t.Fatal("Expected node to exist")
		}

		if !bytes.Equal(node.MAC, mac1) {
			t.Errorf("Expected MAC %x, got %x", mac1, node.MAC)
		}

		if node.AdapterType != AdapterTypePIR {
			t.Errorf("Expected adapter type %d, got %d", AdapterTypePIR, node.AdapterType)
		}

		if node.Uptime != 1000 {
			t.Errorf("Expected uptime 1000, got %d", node.Uptime)
		}
	})

	t.Run("GetAllNodes", func(t *testing.T) {
		// Add second node
		registry.UpdateNode(mac2, AdapterTypeLED, 2000, 2)

		nodes := registry.GetAllNodes()
		if len(nodes) != 2 {
			t.Errorf("Expected 2 nodes, got %d", len(nodes))
		}
	})

	t.Run("NodeCount", func(t *testing.T) {
		count := registry.NodeCount()
		if count != 2 {
			t.Errorf("Expected count 2, got %d", count)
		}
	})
}

func TestSerialComm(t *testing.T) {
	mockPort := NewMockSerialPort()
	comm := NewSerialComm(mockPort)

	t.Run("WriteAndReadFrame", func(t *testing.T) {
		// Create test message
		originalMsg := &MeshMessage{
			MessageType: MessageTypeAdapterData,
			DataType:    AdapterTypePIR,
			Data:        []byte{0x01, 0x02, 0x03, 0x04},
		}

		// Write frame
		err := comm.WriteFrame(originalMsg)
		if err != nil {
			t.Fatalf("Expected no error writing frame, got %v", err)
		}

		// Get written data and add it to read buffer for testing
		writtenData := mockPort.GetWrittenData()
		mockPort.AddReadData(writtenData)

		// Read frame back
		readMsg, err := comm.ReadFrame()
		if err != nil {
			t.Fatalf("Expected no error reading frame, got %v", err)
		}

		// Compare messages
		if readMsg.MessageType != originalMsg.MessageType {
			t.Errorf("Expected MessageType %d, got %d", originalMsg.MessageType, readMsg.MessageType)
		}

		if readMsg.DataType != originalMsg.DataType {
			t.Errorf("Expected DataType %d, got %d", originalMsg.DataType, readMsg.DataType)
		}

		if !bytes.Equal(readMsg.Data, originalMsg.Data) {
			t.Errorf("Expected Data %x, got %x", originalMsg.Data, readMsg.Data)
		}
	})
}

func TestStringToMAC(t *testing.T) {
	testCases := []struct {
		input    string
		expected []byte
		hasError bool
	}{
		{"aa:bb:cc:dd:ee:ff", []byte{0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF}, false},
		{"11:22:33:44:55:66", []byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66}, false},
		{"aabbccddeeff", []byte{0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF}, false},
		{"invalid", nil, true},
		{"aa:bb:cc:dd:ee", nil, true}, // Too short
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result, err := StringToMAC(tc.input)
			
			if tc.hasError {
				if err == nil {
					t.Errorf("Expected error for input %s", tc.input)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for input %s: %v", tc.input, err)
				}
				
				if !bytes.Equal(result, tc.expected) {
					t.Errorf("Expected %x, got %x for input %s", tc.expected, result, tc.input)
				}
			}
		})
	}
}

func TestGetAdapterTypeName(t *testing.T) {
	testCases := []struct {
		adapterType int32
		expected    string
	}{
		{AdapterTypeUnknown, "Unknown"},
		{AdapterTypePIR, "PIR"},
		{AdapterTypeWIFI, "WiFi"},
		{AdapterTypeLED, "LED"},
		{AdapterTypeSerial, "Serial"},
		{99, "Unknown(99)"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			result := GetAdapterTypeName(tc.adapterType)
			if result != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, result)
			}
		})
	}
}
