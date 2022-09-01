package basic

import (
	"context"
	"fmt"
	"log"

	kafkago "github.com/segmentio/kafka-go"
)

func Example_producerKafkaGo() {
	var (
		brokers = []string{"127.0.0.1:29092"}
		topic   = "test"
	)

	// setup config
	producer := kafkago.Writer{
		Addr:  kafkago.TCP(brokers...),
		Async: true,
	}

	// send messages
	ctx := context.Background()
	for i := 0; i < 10; i++ {
		s := fmt.Sprint("msg", i)
		msg := kafkago.Message{
			Topic: topic,
			Value: []byte(s),
		}

		if err := producer.WriteMessages(ctx, msg); err != nil {
			log.Printf("ERR: produce: %v", err)
		}
	}

	// shut down producer
	if err := producer.Close(); err != nil {
		log.Printf("ERR: close producer: %v", err)
	}
}
