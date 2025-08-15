package mesh

// MockEventStore provides a mock implementation for testing
type MockEventStore struct {
	messages []string
	topics   []string
}

// NewMockEventStore creates a new mock event store
func NewMockEventStore() *MockEventStore {
	return &MockEventStore{
		messages: make([]string, 0),
		topics:   make([]string, 0),
	}
}

// Connect implements EventStore_interface
func (m *MockEventStore) Connect() error {
	return nil
}

// WriteMessage implements EventStore_interface
func (m *MockEventStore) WriteMessage(event string, topic string) error {
	m.messages = append(m.messages, event)
	m.topics = append(m.topics, topic)
	return nil
}

// SubscribeToEvents implements EventStore_interface
func (m *MockEventStore) SubscribeToEvents(topic string) error {
	return nil
}

// GetMessages returns all written messages (for testing)
func (m *MockEventStore) GetMessages() []string {
	return m.messages
}

// GetTopics returns all written topics (for testing)
func (m *MockEventStore) GetTopics() []string {
	return m.topics
}
