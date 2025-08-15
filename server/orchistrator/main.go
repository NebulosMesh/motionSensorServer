package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	EventStore "github.com/superbrobenji/motionServer/eventStore"
	"github.com/superbrobenji/motionServer/mesh"
)

var (
	broker  = "kafka:9094" // Use docker compose service name
	groupId = "1"
)

func main() {
	// Command line flags
	serialPort := flag.String("serial", "/dev/ttyUSB0", "Serial port for mesh communication")
	baudRate := flag.Int("baud", 115200, "Serial baud rate")
	apiPort := flag.Int("port", 8080, "HTTP API port")
	flag.Parse()

	log.Printf("Starting Planetopia Motion Sensor Server")
	log.Printf("Serial: %s @ %d baud", *serialPort, *baudRate)
	log.Printf("API Port: %d", *apiPort)
	log.Printf("Kafka Broker: %s", broker)

	// Setup event store
	eventStore := EventStore.New(broker, groupId)
	err := eventStore.Connect()
	if err != nil {
		log.Printf("Warning: Failed to connect to Kafka: %v", err)
		log.Printf("Continuing without Kafka integration...")
		eventStore = nil
	} else {
		log.Printf("Connected to Kafka successfully")
	}

	// Setup mesh server
	meshConfig := mesh.MeshServerConfig{
		SerialPort:    *serialPort,
		BaudRate:      *baudRate,
		HealthTimeout: 30 * time.Second,
		EventStore:    eventStore,
	}

	meshServer := mesh.NewMeshServer(meshConfig)

	// Start mesh server
	if err := meshServer.Start(); err != nil {
		log.Printf("Warning: Failed to start mesh server: %v", err)
		log.Printf("Mesh functionality will be disabled")
	} else {
		log.Printf("Mesh server started successfully")
		
		// Request initial health reports
		time.AfterFunc(2*time.Second, func() {
			if err := meshServer.RequestHealthReports(); err != nil {
				log.Printf("Failed to request initial health reports: %v", err)
			}
		})
	}

	// Start HTTP API server
	go func() {
		if err := mesh.StartAPIServer(meshServer, *apiPort); err != nil {
			log.Printf("API server error: %v", err)
		}
	}()

	// Setup graceful shutdown

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Printf("Server started successfully. Press Ctrl+C to shutdown.")
	
	// Wait for shutdown signal
	<-sigChan
	log.Printf("Shutdown signal received, stopping services...")

	// Stop mesh server
	if meshServer.IsRunning() {
		if err := meshServer.Stop(); err != nil {
			log.Printf("Error stopping mesh server: %v", err)
		}
	}

	log.Printf("Server shutdown complete")
}
