package kafka

import (
	"errors"
	"fmt"
	confluentkafka "github.com/confluentinc/confluent-kafka-go/kafka"
	"sync"
)

type Producer struct {
	Broker   string
	timeout  int
	producer *confluentkafka.Producer
	wg       sync.WaitGroup
	quit     chan bool
}

func (p *Producer) getProducer() (*confluentkafka.Producer, error) {
	producer, err := confluentkafka.NewProducer(&confluentkafka.ConfigMap{
		"bootstrap.servers": p.Broker,
	})
	return producer, err
}

func (p *Producer) Initialize(Broker string, Timeout int) {
	p.Broker = Broker
	p.timeout = Timeout
	var err error
	p.producer, err = p.getProducer()
	if err != nil {
		panic(err)
	}
	p.quit = make(chan bool)
	go p.handleEvents()
}

func (p *Producer) handleEvents() error {

	if p.producer == nil {
		return errors.New("producer not found")
	}

	eventChan := p.producer.Events()
	for {
		select {
		case e := <-eventChan:
			switch ev := e.(type) {
			case *confluentkafka.Message:
				if ev.TopicPartition.Error != nil {
					fmt.Printf("Delivery failed: %v\n", ev.TopicPartition)
				} else {
					p.wg.Done()
				}
			}
		case <-p.quit:
			fmt.Println("quit")
			p.wg.Wait()
			return nil
		}
	}
}

func (p *Producer) Produce(msg []byte, topic string) error {
	if p.producer == nil {
		var err error
		p.producer, err = p.getProducer()
		if err != nil {
			panic(err)
		}
	}
	p.wg.Add(1)
	p.producer.Produce(&confluentkafka.Message{
		TopicPartition: confluentkafka.TopicPartition{Topic: &topic, Partition: confluentkafka.PartitionAny},
		Value:          msg,
	}, nil)

	return nil
}
