package basic

import (
	"log"
	"strings"

	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"
)

func Example_consumerCKafka() {
	var (
		brokers = []string{"127.0.0.1:29092"}
		topics  = []string{"test"}
		group   = "ckafka"
	)

	config := &ckafka.ConfigMap{
		"bootstrap.servers": strings.Join(brokers, ","),
		"group.id":          group,
	}
	consumer, err := ckafka.NewConsumer(config)
	if err != nil {
		log.Fatalf("new consumer: %v", err)
	}
	defer func() { _ = consumer.Close() }()

	if err := consumer.SubscribeTopics(topics, nil); err != nil {
		log.Fatalf("subscribe topics: %v", err)
	}

	for {
		msg, err := consumer.ReadMessage(-1)
		if err != nil {
			if e, ok := err.(ckafka.Error); ok {
				log.Printf("ERR: consume: %v", e)
			}
			continue
		}

		log.Printf("msg:%s topic:%q partition:%d offset:%d",
			msg.Value,
			*msg.TopicPartition.Topic,
			msg.TopicPartition.Partition,
			msg.TopicPartition.Offset,
		)
	}
}
