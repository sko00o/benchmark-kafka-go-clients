package basic

import (
	"fmt"
	"log"

	"github.com/Shopify/sarama"
)

func Example_producerSyncSarama() {
	var (
		brokers = []string{"127.0.0.1:29092"}
		topic   = "test"
	)

	// setup config
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		log.Fatalf("new producer: %v", err)
	}

	// send messages
	for i := 0; i < 10; i++ {
		s := fmt.Sprint("msg", i)
		msg := &sarama.ProducerMessage{
			Topic: topic,
			Value: sarama.StringEncoder(s),
		}

		p, o, err := producer.SendMessage(msg)
		if err != nil {
			log.Printf("ERR: send partation %d offset %d: %v", p, o, err)
		}
	}

	// shut down producer
	if err := producer.Close(); err != nil {
		log.Printf("ERR: close producer: %v", err)
	}
}

func Example_producerAsyncSarama() {
	var (
		brokers = []string{"localhost:29092"}
		topic   = "test"
	)

	// setup config
	config := sarama.NewConfig()
	producer, err := sarama.NewAsyncProducer(brokers, config)
	if err != nil {
		log.Fatalf("new producer: %v", err)
	}

	// handle producer errors in another goroutine
	go func() {
		for err := range producer.Errors() {
			log.Printf("ERR: produce: %v", err)
		}
	}()

	// send messages
	for i := 0; i < 10; i++ {
		s := fmt.Sprint("msg", i)
		msg := &sarama.ProducerMessage{
			Topic: topic,
			Value: sarama.StringEncoder(s),
		}
		producer.Input() <- msg
	}

	// shut down producer, it will flush all pending messages
	if err := producer.Close(); err != nil {
		log.Printf("ERR: close producer: %v", err)
	}
}
