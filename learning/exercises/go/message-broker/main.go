package main

import (
	"fmt"
)

// 1. The data model
// Think of the broker as a catalog:

// Broker keeps track of topics by name.
// Topic is divided into partitions.
// Partition is an ordered, append-only log of records.
// Record is one entry in that log.
// That gives you this shape:

// Broker
//
//	"orders" ──> Topic
//	              ├── Partition 0 ──> Record, Record, Record
//	              └── Partition 1 ──> Record, Record
type Broker struct {
	topics map[string]*Topic
}

type Consumer struct {
	broker    *Broker
	positions map[TopicPartition]int
}

type TopicPartition struct {
	Topic     string
	Partition int
}

type Topic struct {
	Partitions []*Partition `json:"partitions"`
}

type Partition struct {
	Records []Record `json:"records"`
}

type Record struct {
	Offset int `json:"offset"`
	Key    int `json:"key"`
	Value  int `json:"value"`
}

func NewBroker() *Broker {
	return &Broker{
		topics: make(map[string]*Topic),
	}
}

func NewConsumer(broker *Broker) *Consumer {
	return &Consumer{
		broker:    broker,
		positions: make(map[TopicPartition]int),
	}
}

func main() {
	broker := NewBroker()
	broker.CreateTopic("orders", 2)
	broker.Append("orders", 1, 123, 345)
	broker.Append("orders", 1, 567, 678)

	fmt.Println("hay", broker.topics)
	consumer := NewConsumer(broker)
	record1, ok1, err1 := consumer.Read("orders", 1)
	record2, ok2, err2 := consumer.Read("orders", 1)
	record3, ok3, err3 := consumer.Read("orders", 1)
	consumer2 := NewConsumer(broker)
	record4, ok4, err4 := consumer2.Read("orders", 1)
	fmt.Println(record1, ok1, err1)
	fmt.Println(record2, ok2, err2)
	fmt.Println(record3, ok3, err3)
	fmt.Println("4", record4, ok4, err4)
	// consumer :=
}

func (c *Consumer) Read(topicName string, partitionIndex int) (Record, bool, error) {
	topic, exists := c.broker.topics[topicName]
	if !exists {
		return Record{}, false, fmt.Errorf("topic does not exist")
	}

	records := topic.Partitions[partitionIndex].Records
	tp := TopicPartition{Topic: topicName, Partition: partitionIndex}
	position := c.positions[tp]

	if position >= len(records) {
		return Record{}, false, nil
	}

	record := records[position]
	c.positions[tp] = position + 1
	return record, true, nil
}

func (b *Broker) Append(topicName string, partitionIndex, key, value int) (Record, error) {
	topic, exists := b.topics[topicName]
	if !exists {
		return Record{}, fmt.Errorf("topic does not exist")
	}
	if partitionIndex < 0 {
		return Record{}, fmt.Errorf("partition index out of range")
	}
	if partitionIndex >= len(b.topics[topicName].Partitions) {
		return Record{}, fmt.Errorf("partition index out of range")
	}
	partition := topic.Partitions[partitionIndex]

	offset := len(partition.Records)

	newRecord := Record{Offset: offset, Value: value, Key: key}
	partition.Records = append(partition.Records, newRecord)

	return newRecord, nil
}

func (b *Broker) CreateTopic(name string, partitionCount int) error {
	if name == "" {
		return fmt.Errorf("name must not be empty")
	}
	_, exists := b.topics[name]
	if exists {
		return fmt.Errorf("topic already exists")
	}
	var partitions = []*Partition{}
	var num int
	for num < partitionCount {
		partitions = append(partitions, &Partition{})
		num++
	}
	b.topics[name] = &Topic{Partitions: partitions}

	// proof that it worked correctly
	// fmt.Println(len(b.topics[name].Partitions))
	return nil
}
