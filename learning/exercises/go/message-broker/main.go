package main

import (
	"fmt"
	"slices"
)

type Message[T any] struct {
	Offset int
	Value  T
}

type Broker[T any] struct {
	topics map[string][]Message[T]
}

type Admin[T any] struct {
	broker *Broker[T]
}

type Producer[T any] struct {
	broker *Broker[T]
}

type Consumer[T any] struct {
	broker           *Broker[T]
	topicsSubscribed []string
	// later TopicPartition
}

func NewBroker[T any]() *Broker[T] {
	return &Broker[T]{
		topics: make(map[string][]Message[T]),
	}
}

func NewAdmin[T any](broker *Broker[T]) *Admin[T] {
	return &Admin[T]{
		broker: broker,
	}
}

func NewProducer[T any](broker *Broker[T]) *Producer[T] {
	return &Producer[T]{
		broker: broker,
	}
}
func NewConsumer[T any](broker *Broker[T]) *Consumer[T] {
	return &Consumer[T]{
		broker:           broker,
		topicsSubscribed: []string{},
	}
}

func main() {
	b := NewBroker[int]()
	a := NewAdmin(b)
	p := NewProducer(b)
	c := NewConsumer(b)
	c.Subscribe([]string{"users"})
	a.CreateTopics([]string{"orders", "users"})
	p.Send(ProducerRecord[int]{topic: "orders", value: 100})
}

func (a *Admin[T]) CreateTopics(names []string) error {
	for _, name := range names {
		_, exists := a.broker.topics[name]
		if exists {
			return fmt.Errorf("topic already exists")
		} else {
			a.broker.topics[name] = []Message[T]{}
		}
	}
	return nil
}

type ProducerRecord[T any] struct {
	topic string
	value T
}

func (p *Producer[T]) Send(record ProducerRecord[T]) error {
	if record.topic == "" {
		return fmt.Errorf("topic must not be empty")
	}
	messages, exists := p.broker.topics[record.topic]
	// default to 0 on first element
	var nextOffset int
	if len(messages) > 0 {
		lastElement := messages[len(messages)-1]
		nextOffset = lastElement.Offset + 1
	}
	if !exists {
		return fmt.Errorf("topic does not exist")
	} else {
		p.broker.topics[record.topic] = append(messages, Message[T]{
			Offset: nextOffset, Value: record.value,
		})
	}
	return nil
}

func (c *Consumer[T]) Subscribe(topics []string) {
	for _, topic := range topics {
		exists := slices.Contains(c.topicsSubscribed, topic)
		if !exists {
			c.topicsSubscribed = append(c.topicsSubscribed, topic)
		}
	}
}

// func (c *Consumer) Seek(partition TopicPartition, offset int) {}
