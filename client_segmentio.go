package main

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/apex/log"
	kafka "github.com/segmentio/kafka-go"
)

// ==== segmentio/kafka-go ====

func consumeKafkaGo() {
	logger := log.WithFields(log.Fields{
		"client": "kafka-go",
		"mode":   "consumer",
	})

	group, _ := newUUID()
	consumer := kafka.NewReader(kafka.ReaderConfig{
		Brokers:          strings.Split(brokers, ","),
		GroupID:          group,
		GroupTopics:      []string{topic},
		SessionTimeout:   SessionTimeout,
		CommitInterval:   CommitInterval,
		RebalanceTimeout: RebalanceTimeout,
		StartOffset:      kafka.FirstOffset,
		MaxWait:          MaxWait,
		MinBytes:         MinBytes,
		MaxBytes:         MaxBytes,
	})
	// NOTE: for fast quit after benchmark
	//defer func() {
	//	if err := consumer.Close(); err != nil {
	//		logger.WithError(err).Error("close consumer")
	//	}
	//}()

	msgCount := 0
	var start = time.Now()
	done := make(chan bool)
	go func() {
		for {
			if _, err := consumer.ReadMessage(context.Background()); err != nil {
				log.WithError(err).Errorf("consume")
				done <- false
				return
			}
			msgCount++
			if msgCount >= numMessages {
				done <- true
			}
		}
	}()
	<-done
	elapsed := time.Since(start)
	logger.Infof("msg/s: %.2f", float64(numMessages)/elapsed.Seconds())
}

func produceKafkaGo() {
	logger := log.WithFields(log.Fields{
		"client": "kafka-go",
		"mode":   "producer",
	})

	producer := &batchProducer{
		Writer: &kafka.Writer{
			Addr:         kafka.TCP(brokers),
			Topic:        topic,
			Balancer:     &kafka.Hash{},
			BatchTimeout: BatchTimeout,
			BatchSize:    BatchSize,
			BatchBytes:   BatchBytes,
			RequiredAcks: kafka.RequireNone,
			Async:        false,
		},
		logger: logger,
		sema:   make(chan struct{}, 10000),
		wg:     new(sync.WaitGroup),
	}
	var start = time.Now()
	for j := 0; j < numMessages; j++ {
		msg := kafka.Message{
			Value: value,
		}

		if err := producer.WriteMessages(context.Background(), msg); err != nil {
			logger.WithError(err).Error("produce")
		}
	}
	// flush all pending messages
	func() {
		if err := producer.Close(); err != nil {
			log.WithError(err).Errorf("close producer")
		}
	}()

	elapsed := time.Since(start)
	logger.Infof("msg/s: %.2f", float64(numMessages)/elapsed.Seconds())
}

type batchProducer struct {
	*kafka.Writer
	logger *log.Entry

	// limit max goroutine number
	sema chan struct{}
	// wait for all goroutine
	wg *sync.WaitGroup
}

func (bp *batchProducer) WriteMessages(ctx context.Context, msg kafka.Message) error {
	bp.wg.Add(1)
	bp.sema <- struct{}{}
	go func() {
		defer func() {
			<-bp.sema
			bp.wg.Done()
		}()

		if err := bp.Writer.WriteMessages(ctx, msg); err != nil {
			bp.logger.WithError(err).Error("produce")
		}
	}()
	return nil
}

func (bp *batchProducer) Close() error {
	bp.wg.Wait()
	close(bp.sema)
	return bp.Writer.Close()
}
