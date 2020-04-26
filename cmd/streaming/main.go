package main

import (
	"github.com/im-so-sorry/streaming-vk/pkg/brokers/kafka"
	"github.com/im-so-sorry/streaming-vk/pkg/vk"
	"github.com/joho/godotenv"
	"log"
	"os"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	vkToken := os.Getenv("VK_TOKEN")
	kafkaHost := os.Getenv("KAFKA_HOST")
	kafkaProduceTopic := os.Getenv("KAFKA_PRODUCE_TOPIC")

	client := vk.Client{}

	client.Initialize(vkToken)

	v, _ := client.GetServerUrl()

	log.Println("host:", v.Endpoint, "key:", v.Key)

	rules, _ := client.GetRules(v)

	log.Println(rules)

	messageChan := make(chan []byte)

	producer := kafka.Producer{}

	producer.Initialize(kafkaHost, 0)

	go func() {
		for message := range messageChan {
			log.Println("Sending message....")
			log.Println(kafkaProduceTopic)
			err := producer.Produce(message, kafkaProduceTopic)

			if err != nil {
				log.Fatal(err)
			} else {
				log.Println("Message send...")
			}

		}
	}()

	client.Stream(v, 1, messageChan)
}
