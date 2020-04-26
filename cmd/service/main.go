package main

import (
	"encoding/json"
	confluentkafka "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/im-so-sorry/streaming-vk/pkg/brokers/kafka"
	"github.com/im-so-sorry/streaming-vk/pkg/vk"
	"github.com/joho/godotenv"
	"log"
	"os"
)

type CommandContent struct {
	CommandType string `json:"command_type"`
	Rule        vk.Rule
}

type MessageMeta struct {
}

type CommandMessage struct {
	Content CommandContent `json:"content"`
	Meta    MessageMeta    `json:"meta"`
}

type CommandResponse struct {
}

type ResponseMessage struct {
	Meta MessageMeta `json:"meta"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	vkToken := os.Getenv("VK_TOKEN")
	kafkaHost := os.Getenv("KAFKA_HOST")
	kafkaProduceTopic := os.Getenv("KAFKA_COMMAND_RESULT_TOPIC")
	kafkaCommandTopic := os.Getenv("KAFKA_COMMAND_TOPIC")

	consumer := kafka.Consumer{}
	producer := kafka.Producer{}

	producer.Initialize(kafkaHost, 0)
	consumer.Initialize(kafkaHost, "vk_command", []string{kafkaCommandTopic}, 0)

	client := vk.Client{}

	client.Initialize(vkToken)

	auth, _ := client.GetServerUrl()

	messageChan := make(chan *confluentkafka.Message)

	go func() {
		var cmd CommandMessage
		for message := range messageChan {

			if err := json.Unmarshal(message.Value, &cmd); err != nil {
				log.Fatal("unmarshal response json failed:", err)
			}

			switch cmd.Content.CommandType {
			case "list":
				if rules, err := client.GetRules(auth); err == nil {
					msg, _ := json.Marshal(rules)
					producer.Produce(msg, kafkaProduceTopic)
				} else {
					log.Fatal(err)
				}
			case "add":
				response, err := client.AddRule(auth, cmd.Content.Rule)

				if err != nil {
					log.Println(err)
					continue
				}
				msg, _ := json.Marshal(response)
				producer.Produce(msg, kafkaProduceTopic)

			case "remove":
				response, err := client.RemoveRule(auth, cmd.Content.Rule.Tag)

				if err != nil {
					log.Println(err)
					continue
				}
				msg, _ := json.Marshal(response)
				producer.Produce(msg, kafkaProduceTopic)
			default:
				log.Println("Unknown command type:", cmd.Content.CommandType)
			}

		}
	}()

	consumer.Listen(messageChan)

}
