package kafka

import (
	_ "github.com/confluentinc/confluent-kafka-go/kafka"
)

type Consumer struct {
	Broker string
	Group string
	Topics []string
	timeout int
}

func (c *Consumer) Initialize(Broker string, Group string, Topics []string, Timeout int) {
	c.Broker = Broker
	c.Group = Group
	c.Topics = Topics
	c.timeout = Timeout
}


func (c *Consumer) GetConsumer() (Consumer, error) {

}

func (c *Consumer) listen() {

}