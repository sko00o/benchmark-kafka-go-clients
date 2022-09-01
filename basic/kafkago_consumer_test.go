package basic

import (
	"context"
	"log"
	"time"

	kafkago "github.com/segmentio/kafka-go"
)

func Example_consumerKafkaGo() {
	var (
		brokers = []string{"127.0.0.1:29092"}
		topics  = []string{"test"}
		group   = "kafkago"
	)

	config := kafkago.ReaderConfig{
		Brokers:     brokers,
		GroupID:     group,
		GroupTopics: topics,

		CommitInterval: 2 * time.Second,
	}
	consumer := kafkago.NewReader(config)
	defer func() { _ = consumer.Close() }()

	ctx := context.Background()
	for {
		msg, err := consumer.ReadMessage(ctx)
		if err != nil {
			log.Printf("ERR: consume: %v", err)
			continue
		}

		log.Printf("msg:%s topic:%q partition:%d offset:%d",
			msg.Value,
			msg.Topic,
			msg.Partition,
			msg.Offset,
		)
	}
}
