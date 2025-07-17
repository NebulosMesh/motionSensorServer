package eventstore

import (
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type store struct {
	broker   string
	groupId  string
	producer *kafka.Producer
	consumer *kafka.Consumer
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
	fmt.Printf("connecting Producer: %v\n", store.broker)
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": store.broker})
	if err != nil {
		return err
	}
	store.producer = p
	fmt.Printf("connection sucessful with producer : %v\n", store.broker)

	fmt.Printf("connecting consumer: %v\n, on Group: %v\n", store.broker, store.groupId)
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": store.broker,
		"group.id":          store.groupId,
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		return err
	}
	store.consumer = c
	fmt.Printf("connection sucessful with consumer : %v\n, on Group: %v\n", store.broker, store.groupId)

	return nil
}

func (store *store) WriteMessage(event string, topic string) error {
	fmt.Printf("Delivering %v\n", event)
	go func() {
		for e := range store.producer.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					fmt.Printf("Delivery failed: %v\n", ev.TopicPartition)
				} else {
					fmt.Printf("Delivered message to %v\n", ev.TopicPartition)
				}
			}
		}
	}()

	store.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          []byte(event),
	}, nil)
	return nil
}

func (store *store) SubscribeToEvents(topic string) error {
	store.consumer.Subscribe(topic, nil)

	for {
		msg, err := store.consumer.ReadMessage(-1)
		if err == nil {
			fmt.Printf("Message on %s: %s\n", msg.TopicPartition, string(msg.Value))
		} else {
			fmt.Printf("Consumer error: %v (%v)\n", err, msg)
			store.consumer.Close()
			return err
		}
	}
}
