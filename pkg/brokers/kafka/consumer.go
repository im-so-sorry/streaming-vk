package kafka

import (
	"fmt"
	confluentkafka "github.com/confluentinc/confluent-kafka-go/kafka"
)

type Consumer struct {
	Broker  string
	Group   string
	Topics  []string
	timeout int
}

func (c *Consumer) Initialize(Broker string, Group string, Topics []string, Timeout int) {
	c.Broker = Broker
	c.Group = Group
	c.Topics = Topics
	c.timeout = Timeout
}

func (c *Consumer) GetConsumer() (*confluentkafka.Consumer, error) {
	consumer, err := confluentkafka.NewConsumer(&confluentkafka.ConfigMap{
		"bootstrap.servers": c.Broker,
		"group.id":          c.Group,
		"auto.offset.reset": "earliest",
	})

	if err != nil {
		return consumer, err
	}

	consumer.SubscribeTopics(c.Topics, nil)

	return consumer, err
}


func (c *Consumer) Listen(messageChan chan *confluentkafka.Message) error {
	consumer, err := c.GetConsumer()

	if err != nil {
		panic(err)
	}

	for {
		msg, err := consumer.ReadMessage(-1)
		if err == nil {
			fmt.Printf("Message on %s: %s\n", msg.TopicPartition, string(msg.Value))
			messageChan <- msg
		} else {
			// The client will automatically try to recover from all errors.
			fmt.Printf("Consumer error: %v (%v)\n", err, msg)
		}
	}

	consumer.Close()

	return nil
}