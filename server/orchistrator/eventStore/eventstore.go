package eventstore

type EventStore_interface interface {
	Connect() error
	WriteMessage(event string, topic string) error
	SubscribeToEvents(topic string) error
}
