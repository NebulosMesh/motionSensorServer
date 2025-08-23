package mesh

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	EventStore "github.com/superbrobenji/motionServer/eventStore"
	"go.bug.st/serial"
)

// MeshServer manages the mesh network communication
type MeshServer struct {
	serialComm     *SerialComm
	nodeRegistry   *NodeRegistry
	messageBuilder *MessageBuilder
	eventStore     EventStore.EventStore_interface
	
	// Configuration
	serialPort     string
	baudRate       int
	healthTimeout  time.Duration
	
	// Runtime state
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	mu         sync.RWMutex
	running    bool
}

// MeshServerConfig holds configuration for the mesh server
type MeshServerConfig struct {
	SerialPort    string
	BaudRate      int
	HealthTimeout time.Duration
	EventStore    EventStore.EventStore_interface
}

// NewMeshServer creates a new mesh server
func NewMeshServer(config MeshServerConfig) *MeshServer {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &MeshServer{
		nodeRegistry:   NewNodeRegistry(),
		messageBuilder: NewMessageBuilder(),
		eventStore:     config.EventStore,
		serialPort:     config.SerialPort,
		baudRate:       config.BaudRate,
		healthTimeout:  config.HealthTimeout,
		ctx:            ctx,
		cancel:         cancel,
	}
}

// Start starts the mesh server
func (ms *MeshServer) Start() error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	
	if ms.running {
		return fmt.Errorf("mesh server is already running")
	}

	// Open serial port
	mode := &serial.Mode{
		BaudRate: ms.baudRate,
		Parity:   serial.NoParity,
		DataBits: 8,
		StopBits: serial.OneStopBit,
	}

	port, err := serial.Open(ms.serialPort, mode)
	if err != nil {
		return fmt.Errorf("failed to open serial port %s: %w", ms.serialPort, err)
	}

	ms.serialComm = NewSerialComm(port)
	ms.running = true

	// Start message processing goroutine
	ms.wg.Add(1)
	go ms.messageProcessor()

	log.Printf("Mesh server started on serial port %s at %d baud", ms.serialPort, ms.baudRate)
	return nil
}

// Stop stops the mesh server
func (ms *MeshServer) Stop() error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	
	if !ms.running {
		return fmt.Errorf("mesh server is not running")
	}

	ms.cancel()
	ms.running = false

	if ms.serialComm != nil {
		ms.serialComm.Close()
	}

	ms.wg.Wait()
	log.Printf("Mesh server stopped")
	return nil
}

// messageProcessor processes incoming messages from the serial port
func (ms *MeshServer) messageProcessor() {
	defer ms.wg.Done()
	
	consecutiveErrors := 0
	maxConsecutiveErrors := 10
	
	for {
		select {
		case <-ms.ctx.Done():
			return
		default:
			msg, err := ms.serialComm.ReadFrame()
			if err != nil {
				consecutiveErrors++
				if consecutiveErrors <= maxConsecutiveErrors {
					log.Printf("Error reading frame: %v", err)
				} else if consecutiveErrors == maxConsecutiveErrors+1 {
					log.Printf("Too many consecutive frame errors, suppressing further error messages. Last error: %v", err)
					log.Printf("Note: If you see 'frame length too large' with ASCII characters (like 'un', 't:', '--'), the ESP32 might be sending text data instead of binary protobuf frames.")
					log.Printf("Check ESP32 firmware and ensure it's configured for mesh protocol, not debug output.")
				}
				// Brief pause to prevent tight error loop
				select {
				case <-ms.ctx.Done():
					return
				case <-time.After(100 * time.Millisecond):
				}
				continue
			}
			
			// Reset error counter on successful read
			if consecutiveErrors > 0 {
				if consecutiveErrors > maxConsecutiveErrors {
					log.Printf("Frame reading recovered after %d consecutive errors", consecutiveErrors)
				}
				consecutiveErrors = 0
			}

			if err := ms.handleMessage(msg); err != nil {
				log.Printf("Error handling message: %v", err)
			}
		}
	}
}

// handleMessage processes a received mesh message
func (ms *MeshServer) handleMessage(msg *MeshMessage) error {
	// Log the message to Kafka
	if err := ms.logMessageToKafka(msg, "incoming"); err != nil {
		log.Printf("Failed to log incoming message to Kafka: %v", err)
	}

	switch msg.MessageType {
	case MessageTypeAdapterData:
		return ms.handleAdapterData(msg)
	case MessageTypeMasterBeacon:
		return ms.handleMasterBeacon(msg)
	default:
		log.Printf("Unknown message type: %d", msg.MessageType)
	}

	return nil
}

// handleAdapterData processes adapter data messages
func (ms *MeshServer) handleAdapterData(msg *MeshMessage) error {
	switch msg.DataType {
	case AdapterTypeSerial:
		return ms.handleSerialData(msg)
	case AdapterTypePIR:
		return ms.handlePIRData(msg)
	default:
		log.Printf("Received adapter data - Type: %s, Origin: %s, Data: %x", 
			GetAdapterTypeName(msg.DataType), 
			macToString(msg.OriginMacAddress), 
			msg.Data)
	}

	return nil
}

// handleSerialData processes serial control messages
func (ms *MeshServer) handleSerialData(msg *MeshMessage) error {
	if len(msg.Data) == 0 {
		return fmt.Errorf("empty serial data")
	}

	opcode := msg.Data[0]
	switch opcode {
	case OpHealthReport:
		return ms.handleHealthReport(msg)
	default:
		log.Printf("Unknown serial opcode: 0x%02x", opcode)
	}

	return nil
}

// handleHealthReport processes health report messages
func (ms *MeshServer) handleHealthReport(msg *MeshMessage) error {
	healthReport, err := ms.messageBuilder.ParseHealthReport(msg)
	if err != nil {
		return fmt.Errorf("failed to parse health report: %w", err)
	}

	// Update node registry
	ms.nodeRegistry.UpdateNode(
		healthReport.MAC,
		healthReport.AdapterType,
		healthReport.Uptime,
		healthReport.HopCount,
	)

	log.Printf("Health report from %s: Type=%s, Uptime=%ds, Hops=%d",
		macToString(healthReport.MAC),
		GetAdapterTypeName(healthReport.AdapterType),
		healthReport.Uptime,
		healthReport.HopCount)

	return nil
}

// handlePIRData processes PIR sensor data
func (ms *MeshServer) handlePIRData(msg *MeshMessage) error {
	log.Printf("PIR motion detected from %s (hops: %d)", 
		macToString(msg.OriginMacAddress), 
		msg.HopCount)

	// Log PIR event to Kafka with more specific topic
	pirEvent := map[string]interface{}{
		"type":      "pir_motion",
		"mac":       macToString(msg.OriginMacAddress),
		"timestamp": time.Now().Unix(),
		"hopCount":  msg.HopCount,
		"data":      msg.Data,
	}

	eventJSON, _ := json.Marshal(pirEvent)
	if err := ms.eventStore.WriteMessage(string(eventJSON), "motion-trigger"); err != nil {
		log.Printf("Failed to log PIR event to Kafka: %v", err)
	}

	return nil
}

// handleMasterBeacon processes master beacon messages
func (ms *MeshServer) handleMasterBeacon(msg *MeshMessage) error {
	log.Printf("Master beacon from %s", macToString(msg.OriginMacAddress))
	return nil
}

// SendMessage sends a message to the mesh network
func (ms *MeshServer) SendMessage(msg *MeshMessage) error {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	
	if !ms.running {
		return fmt.Errorf("mesh server is not running")
	}

	// Log the outgoing message
	if err := ms.logMessageToKafka(msg, "outgoing"); err != nil {
		log.Printf("Failed to log outgoing message to Kafka: %v", err)
	}

	return ms.serialComm.WriteFrame(msg)
}

// ConfigureNode sets the adapter type for a specific node
func (ms *MeshServer) ConfigureNode(targetMAC []byte, adapterType int32) error {
	msg, err := ms.messageBuilder.BuildConfigSetMessage(targetMAC, adapterType)
	if err != nil {
		return fmt.Errorf("failed to build config message: %w", err)
	}

	log.Printf("Configuring node %s to adapter type %s", 
		macToString(targetMAC), 
		GetAdapterTypeName(adapterType))

	return ms.SendMessage(msg)
}

// ConfigureAllNodes sets the adapter type for all nodes
func (ms *MeshServer) ConfigureAllNodes(adapterType int32) error {
	msg, err := ms.messageBuilder.BuildConfigSetBroadcastMessage(adapterType)
	if err != nil {
		return fmt.Errorf("failed to build broadcast config message: %w", err)
	}

	log.Printf("Configuring all nodes to adapter type %s", 
		GetAdapterTypeName(adapterType))

	return ms.SendMessage(msg)
}

// RequestHealthReports requests health reports from all nodes
func (ms *MeshServer) RequestHealthReports() error {
	msg := ms.messageBuilder.BuildHealthRequestMessage()
	
	log.Printf("Requesting health reports from all nodes")
	return ms.SendMessage(msg)
}

// BroadcastData broadcasts data to all nodes
func (ms *MeshServer) BroadcastData(dataType int32, data []byte) error {
	msg, err := ms.messageBuilder.BuildBroadcastMessage(dataType, data)
	if err != nil {
		return fmt.Errorf("failed to build broadcast message: %w", err)
	}

	log.Printf("Broadcasting data: Type=%s, Length=%d", 
		GetAdapterTypeName(dataType), 
		len(data))

	return ms.SendMessage(msg)
}

// GetNodeRegistry returns the node registry
func (ms *MeshServer) GetNodeRegistry() *NodeRegistry {
	return ms.nodeRegistry
}

// IsRunning returns whether the server is running
func (ms *MeshServer) IsRunning() bool {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return ms.running
}

// logMessageToKafka logs messages to Kafka for debugging and monitoring
func (ms *MeshServer) logMessageToKafka(msg *MeshMessage, direction string) error {
	if ms.eventStore == nil {
		return nil // Event store not configured
	}

	logEntry := map[string]interface{}{
		"timestamp":   time.Now().Unix(),
		"direction":   direction,
		"messageType": msg.MessageType,
		"dataType":    msg.DataType,
		"origin":      macToString(msg.OriginMacAddress),
		"target":      macToString(msg.TargetMacAddress),
		"lastHop":     macToString(msg.LastHopMacAddress),
		"hopCount":    msg.HopCount,
		"dataLength":  len(msg.Data),
	}

	// Add specific fields for health reports
	if ms.messageBuilder.IsHealthReport(msg) {
		if healthReport, err := ms.messageBuilder.ParseHealthReport(msg); err == nil {
			logEntry["healthReport"] = map[string]interface{}{
				"mac":         macToString(healthReport.MAC),
				"adapterType": GetAdapterTypeName(healthReport.AdapterType),
				"uptime":      healthReport.Uptime,
			}
		}
	}

	logJSON, err := json.Marshal(logEntry)
	if err != nil {
		return fmt.Errorf("failed to marshal log entry: %w", err)
	}

	return ms.eventStore.WriteMessage(string(logJSON), "mesh-messages")
}
