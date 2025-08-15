package eventstore

import (
	"context"
	"fmt"

	"github.com/segmentio/kafka-go"
)

type store struct {
	broker string
	groupId string
	writer *kafka.Writer
	reader *kafka.Reader
}

func New(broker string, groupId string) EventStore_interface {
	var eventStore EventStore_interface
	eventStore = &store{
		broker:  broker,
		groupId: groupId,
	}
	return eventStore
}

func (store *store) Connect() error {
	fmt.Printf("connecting to Kafka: %v\n", store.broker)
	
	// Create writer for producing messages
	store.writer = &kafka.Writer{
		Addr:     kafka.TCP(store.broker),
		Balancer: &kafka.LeastBytes{},
	}
	
	// Don't create reader here - it will be created when subscribing to specific topics
	store.reader = nil
	
	fmt.Printf("connection successful with Kafka: %v\n", store.broker)
	return nil
}

func (store *store) WriteMessage(event string, topic string) error {
	fmt.Printf("Delivering %v to topic %v\n", event, topic)
	
	err := store.writer.WriteMessages(context.Background(),
		kafka.Message{
			Topic: topic,
			Value: []byte(event),
		},
	)
	
	if err != nil {
		fmt.Printf("Delivery failed: %v\n", err)
		return err
	}
	
	fmt.Printf("Delivered message to topic %v\n", topic)
	return nil
}

func (store *store) SubscribeToEvents(topic string) error {
	// Close existing reader if any
	if store.reader != nil {
		store.reader.Close()
	}
	
	// Configure reader for specific topic
	store.reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{store.broker},
		Topic:   topic,
		GroupID: store.groupId,
	})

	fmt.Printf("Subscribed to topic: %s\n", topic)

	for {
		msg, err := store.reader.ReadMessage(context.Background())
		if err == nil {
			fmt.Printf("Message on topic %s: %s\n", msg.Topic, string(msg.Value))
		} else {
			fmt.Printf("Consumer error: %v\n", err)
			store.reader.Close()
			return err
		}
	}
}
