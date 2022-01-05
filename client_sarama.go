package main

import (
	"context"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Shopify/sarama"
	"github.com/apex/log"
)

// ==== Shopify/sarama ====

var KafkaVersion = sarama.V2_1_1_0

func consumeSarama() {
	logger := log.WithFields(log.Fields{
		"client": "sarama",
		"mode":   "consumer",
	})

	config := sarama.NewConfig()
	config.Version = KafkaVersion
	config.Consumer.Group.Session.Timeout = SessionTimeout
	config.Consumer.Group.Rebalance.Timeout = RebalanceTimeout
	config.Consumer.Offsets.AutoCommit.Interval = CommitInterval
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.MaxWaitTime = MaxWait
	config.Consumer.Fetch.Min = MinBytes
	config.Consumer.Fetch.Max = MaxBytes
	if err := config.Validate(); err != nil {
		log.WithError(err).Fatal("validate config")
	}

	group, _ := newUUID()
	client, err := sarama.NewConsumerGroup(strings.Split(brokers, ","), group, config)
	if err != nil {
		logger.WithError(err).Fatal("new consumer")
	}
	// NOTE: for fast quit after benchmark
	//defer func() {
	//	if err := client.Close(); err != nil {
	//		logger.WithError(err).Error("close consumer")
	//	}
	//}()

	consumer := &Consumer{
		ready: make(chan bool),
		done:  make(chan bool),
	}
	ctx := context.Background()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			// `Consume` should be called inside an infinite loop, when a
			// server-side rebalance happens, the consumer session will need to be
			// recreated to get the new claims
			if err := client.Consume(ctx, []string{topic}, consumer); err != nil {
				logger.WithError(err).Fatal("consume loop")
			}
			if ctx.Err() != nil {
				return
			}
			consumer.ready = make(chan bool)
		}
	}()
	<-consumer.ready // wait until the consumer has been set up
	<-consumer.done
	elapsed := time.Since(consumer.start)
	logger.Infof("msg/s: %.2f", float64(numMessages)/elapsed.Seconds())
}

// Consumer represents a Sarama consumer group consumer
type Consumer struct {
	ready    chan bool
	msgCount uint32
	done     chan bool
	start    time.Time
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (consumer *Consumer) Setup(sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	close(consumer.ready)
	consumer.start = time.Now()
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (consumer *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (consumer *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// NOTE:
	// Do not move the code below to a goroutine.
	// The `ConsumeClaim` itself is called within a goroutine, see:
	// https://github.com/Shopify/sarama/blob/main/consumer_group.go#L27-L29
	for message := range claim.Messages() {
		if atomic.AddUint32(&consumer.msgCount, 1) >= uint32(numMessages) {
			consumer.done <- true
		}
		session.MarkMessage(message, "")
	}

	return nil
}

func produceSarama() {
	logger := log.WithFields(log.Fields{
		"client": "sarama",
		"mode":   "producer",
	})

	config := sarama.NewConfig()
	config.Version = KafkaVersion
	config.Producer.Flush.Frequency = BatchTimeout
	config.Producer.Flush.MaxMessages = BatchSize
	config.Producer.Flush.Bytes = BatchBytes
	config.Producer.RequiredAcks = sarama.NoResponse
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = false
	if err := config.Validate(); err != nil {
		log.WithError(err).Fatal("validate config")
	}

	producer, err := sarama.NewAsyncProducer(strings.Split(brokers, ","), config)
	if err != nil {
		logger.WithError(err).Fatal("new producer")
	}
	defer func() {
		if err := producer.Close(); err != nil {
			log.WithError(err).Errorf("close producer")
		}
	}()

	var msgCount int
	go func() {
		for {
			select {
			case <-producer.Successes():
				msgCount++
			}
		}
	}()

	var start = time.Now()
	for j := 0; j < numMessages; j++ {
		msg := &sarama.ProducerMessage{
			Topic:     topic,
			Partition: int32(partition),
			Value:     sarama.ByteEncoder(value),
		}

		producer.Input() <- msg
	}
	// flush all pending messages
	func() {
		if err := producer.Close(); err != nil {
			log.WithError(err).Errorf("close producer")
		}
	}()
	elapsed := time.Since(start)
	logger.Infof("msg/s: %.2f", float64(msgCount)/elapsed.Seconds())
}
