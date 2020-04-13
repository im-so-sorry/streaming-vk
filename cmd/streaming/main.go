package main

import (
	"github.com/im-so-sorry/streaming-vk/pkg/brokers/kafka"
	"github.com/im-so-sorry/streaming-vk/pkg/vk"
	"log"
)

func main() {
	client := vk.Client{}

	client.Initialize("da776f3bda776f3bda776f3b83da1fbe7cdda77da776f3b861a30cb5acd2a8da94f406b")

	v, _ := client.GetServerUrl()

	log.Println("host:", v.Endpoint, "key:", v.Key)

	res, _ := client.AddRule(v, vk.Rule{"vk", "vk"})
	log.Println(res)
	rules, _ := client.GetRules(v)

	log.Println(rules)

	messageChan := make(chan []byte)

	producer := kafka.Producer{}

	producer.Initialize("localhost:9093", 0)

	go func() {
		for message := range messageChan {
			err := producer.Produce(message, "vk_stream")

			if err != nil {
				panic(err)
			}
		}
	}()

	client.Stream(v, 1, messageChan)
}
