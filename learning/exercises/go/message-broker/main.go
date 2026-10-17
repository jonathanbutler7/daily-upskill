package main

type Message struct {
	Offset int
	Value  string
}

type Broker struct {
	// Choose your fields.
}

func NewBroker() *Broker {
	panic("implement me")
}

func (b *Broker) CreateTopic(name string) error {
	panic("implement me")
}

// Publish appends a message and returns its assigned offset.
// Offsets start at 0 within each topic.
func (b *Broker) Publish(topic string, value string) (int, error) {
	panic("implement me")
}

// Read returns up to limit messages, starting at offset.
// Reading does not remove messages.
func (b *Broker) Read(topic string, offset int, limit int) ([]Message, error) {
	panic("implement me")
}

type Consumer struct {
	// Choose your fields.
}

func NewConsumer(broker *Broker, topic string) *Consumer {
	panic("implement me")
}

// Poll reads from this consumer's current position.
// After a successful read, advance past the returned messages.
func (c *Consumer) Poll(limit int) ([]Message, error) {
	panic("implement me")
}

// Seek changes the position, allowing messages to be replayed.
func (c *Consumer) Seek(offset int) error {
	panic("implement me")
}