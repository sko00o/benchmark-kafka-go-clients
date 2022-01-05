package main

import (
	"time"

	"github.com/apex/log"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

// ==== confluentinc/confluent-kafka-go ====

func consumeConfluentKafkaGo() {
	logger := log.WithFields(log.Fields{
		"client": "confluent-kafka-go",
		"mode":   "consumer",
	})
	group, _ := newUUID()
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":       brokers,
		"group.id":                group,
		"session.timeout.ms":      int(SessionTimeout / time.Millisecond),
		"auto.commit.interval.ms": int(CommitInterval / time.Millisecond),
		"auto.offset.reset":       "earliest",
		"fetch.wait.max.ms":       int(MaxWait / time.Millisecond),
		"fetch.min.bytes":         MinBytes,
		"fetch.max.bytes":         MaxBytes,
	})
	if err != nil {
		logger.WithError(err).Fatal("new consumer")
	}
	// NOTE: for fast quit after benchmark
	//defer func() {
	//	if err := consumer.Close(); err != nil {
	//		logger.WithError(err).Error("close consumer")
	//	}
	//}()

	if err := consumer.SubscribeTopics([]string{topic}, nil); err != nil {
		logger.WithError(err).Fatal("subscribe topics")
	}

	var start = time.Now()
	var msgCount = 0
	for msgCount < numMessages {
		_, err := consumer.ReadMessage(-1)
		if err != nil {
			if e, ok := err.(kafka.Error); ok {
				logger.WithError(err).WithField("code", e.Code()).Error("consume loop")
				if e.Code() == kafka.ErrAllBrokersDown {
					break
				}
			}
			continue
		}

		msgCount++
	}
	elapsed := time.Since(start)
	logger.Infof("msg/s: %.2f", float64(msgCount)/elapsed.Seconds())
}

func produceConfluentKafkaGo() {
	logger := log.WithFields(log.Fields{
		"client": "confluent-kafka-go",
		"mode":   "producer",
	})
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers":                     brokers,
		"linger.ms":                             int(BatchTimeout / time.Millisecond),
		"max.in.flight.requests.per.connection": BatchSize,
		"batch.size":                            BatchBytes,
		"acks":                                  "0",
	})
	if err != nil {
		logger.WithError(err).Fatal("new producer")
	}
	defer producer.Close()

	done := make(chan bool)
	var msgCount int
	go func() {
		for e := range producer.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if err := ev.TopicPartition.Error; err != nil {
					logger.WithError(err).Error("delivery")
				}
			default:
				logger.Debugf("ignore: %v", e)
				continue
			}

			msgCount++
			if msgCount >= numMessages {
				done <- true
				break
			}
		}
	}()

	var start = time.Now()
	for j := 0; j < numMessages; j++ {
		msg := &kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     &topic,
				Partition: kafka.PartitionAny,
			},
			Value: value,
		}

		if err := producer.Produce(msg, nil); err != nil {
			logger.WithError(err).Error("produce")
		}
	}
	<-done
	producer.Flush(int(15 * time.Second / time.Millisecond))
	elapsed := time.Since(start)
	logger.Infof("msg/s: %.2f", float64(msgCount)/elapsed.Seconds())
}
