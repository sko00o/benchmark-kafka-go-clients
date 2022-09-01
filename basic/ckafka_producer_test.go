package basic

import (
	"fmt"
	"log"
	"strings"

	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"
)

func Example_producerCKafka() {
	var (
		brokers = []string{"127.0.0.1:29092"}
		topic   = "test"
	)

	// setup config
	config := &ckafka.ConfigMap{
		"bootstrap.servers": strings.Join(brokers, ","),
	}
	producer, err := ckafka.NewProducer(config)
	if err != nil {
		log.Fatalf("new producer: %v", err)
	}

	// handle producer errors in another goroutine
	go func() {
		for e := range producer.Events() {
			switch ev := e.(type) {
			case *ckafka.Message:
				if err := ev.TopicPartition.Error; err != nil {
					log.Printf("ERR: produce: %v", err)
				}
			default:
				log.Printf("DBG: ignore: %v", e)
				continue
			}
		}
	}()

	// send messages
	for i := 0; i < 10; i++ {
		s := fmt.Sprint("msg", i)
		msg := &ckafka.Message{
			TopicPartition: ckafka.TopicPartition{
				Topic: &topic,
			},
			Value: []byte(s),
		}

		if err := producer.Produce(msg, nil); err != nil {
			log.Printf("ERR: produce: %v", err)
		}
	}

	// shut down producer
	producer.Close()
}
