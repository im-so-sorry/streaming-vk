package main

import (
	"encoding/json"
	"github.com/im-so-sorry/streaming-vk/pkg/brokers/kafka"
	"github.com/im-so-sorry/streaming-vk/pkg/vk"
	confluentkafka "github.com/confluentinc/confluent-kafka-go/kafka"
	"log"
)

type CommandContent struct {
	CommandType string `json:"command_type"`
	Rule vk.Rule
}

type MessageMeta struct {

}

type CommandMessage struct {
	Content CommandContent `json:"content"`
	Meta MessageMeta `json:"meta"`
}

type CommandResponse struct {
	
}

type ResponseMessage struct {
	Meta MessageMeta `json:"meta"`
}

func main() {
	consumer := kafka.Consumer{}
	producer := kafka.Producer{}

	resultTopic := "vk.command.result"

	producer.Initialize("localhost:9093", 0)
	consumer.Initialize("localhost:9093", "vk.command", []string{"vk.command"}, 0)

	client := vk.Client{}

	client.Initialize("da776f3bda776f3bda776f3b83da1fbe7cdda77da776f3b861a30cb5acd2a8da94f406b")

	auth, _ := client.GetServerUrl()

	messageChan := make(chan *confluentkafka.Message)


	go func() {
		var cmd  CommandMessage
		for message := range messageChan {

			if err := json.Unmarshal(message.Value, &cmd); err != nil {
				log.Fatal("unmarshal response json failed:", err)
			}

			switch cmd.Content.CommandType {
			case "list":
				if rules, err := client.GetRules(auth); err != nil {
					json.Marshal(rules)
					producer.Produce())
				}
			case "add":
				client.AddRule(auth, cmd.Content.Rule)
			case "remove":
				client.RemoveRule(auth, cmd.Content.Rule.Tag)
			default:
				log.Println("Unknown command type:", cmd.Content.CommandType)
			}

		}
	}()

	consumer.Listen(messageChan)

}
