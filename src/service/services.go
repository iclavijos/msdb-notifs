package service

import (
	"dao"
	"encoding/json"
	"fmt"
	"log"
	"model"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func ProcessNotification(producer *kafka.Producer, sessionData *model.SessionData, minutes int) {

	log.Printf("Processing notification of %v minutes to  %s-%s session start\n", minutes, sessionData.EventName, sessionData.SessionName)
	users := dao.GetUsersSuscribedToSeries(sessionData.SeriesId, minutes)

	for _, user := range users {
		//Send notification to user
		log.Printf("Sending notification to user %s", user.Email)

		// Produce messages to topic (asynchronously)
		topic := "eventNotifications"
		// Optional delivery channel, if not specified the Producer object's
		// .Events channel is used.
		deliveryChan := make(chan kafka.Event)

		msg, _ := json.Marshal(sessionData)
		producer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
			Value:          []byte(msg),
		}, deliveryChan)

		e := <-deliveryChan
		m := e.(*kafka.Message)

		if m.TopicPartition.Error != nil {
			fmt.Printf("Delivery failed: %v\n", m.TopicPartition.Error)
		} else {
			fmt.Printf("Delivered message to topic %s [%d] at offset %v\n",
				*m.TopicPartition.Topic, m.TopicPartition.Partition, m.TopicPartition.Offset)
		}

		close(deliveryChan)

	}
}

func ProcessEventEditionEvents() {

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
		"group.id":          "eventEditionsGroup",
		"auto.offset.reset": "earliest",
	})

	if err != nil {
		panic(err)
	}

	c.SubscribeTopics([]string{"eventEditionTopic"}, nil)

	for {
		msg, err := c.ReadMessage(-1)
		if err == nil {
			fmt.Printf("Message on %s: %s\n", msg.TopicPartition, string(msg.Value))
		} else {
			fmt.Printf("Consumer error: %v (%v)\n", err, msg)
			break
		}
	}

	c.Close()
}
