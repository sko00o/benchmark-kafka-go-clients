package basic

import (
	"context"
	"log"

	"github.com/Shopify/sarama"
)

type consumeHandler struct{}

func (consumeHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (consumeHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h consumeHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		log.Printf("msg:%s topic:%q partition:%d offset:%d",
			msg.Value, msg.Topic, msg.Partition, msg.Offset)
		sess.MarkMessage(msg, "")
	}
	return nil
}

func Example_consumerSarama() {
	var (
		brokers = []string{"127.0.0.1:29092"}
		topics  = []string{"test"}
		group   = "sarama"
	)

	config := sarama.NewConfig()
	consumerGroup, err := sarama.NewConsumerGroup(brokers, group, config)
	if err != nil {
		log.Fatalf("new consumer: %v", err)
	}
	defer func() { _ = consumerGroup.Close() }()

	ctx := context.Background()
	for {
		handler := consumeHandler{}

		// `Consume` should be called inside an infinite loop, when a
		// server-side rebalance happens, the consumer session will need to be
		// recreated to get the new claims
		err := consumerGroup.Consume(ctx, topics, handler)
		if err != nil {
			log.Fatalf("consume: %v", err)
		}
	}
}
