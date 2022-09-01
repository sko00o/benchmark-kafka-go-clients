package benchmark

import (
	"context"
	"sync"
	"testing"

	"github.com/Shopify/sarama"
	ckafkago "github.com/confluentinc/confluent-kafka-go/kafka"
	kafkago "github.com/segmentio/kafka-go"
)

var m = []byte("this is benchmark for three mainstream kafka client")

func BenchmarkSaramaAsync(b *testing.B) {
	b.ReportAllocs()
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	producer, err := sarama.NewAsyncProducer([]string{"localhost:29092"}, config)
	if err != nil {
		panic(err)
	}

	message := &sarama.ProducerMessage{Topic: "test", Value: sarama.ByteEncoder(m)}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		producer.Input() <- message
	}
}

func BenchmarkSaramaAsyncInParallel(b *testing.B) {
	b.ReportAllocs()
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	producer, err := sarama.NewAsyncProducer([]string{"localhost:29092"}, config)
	if err != nil {
		panic(err)
	}

	message := &sarama.ProducerMessage{Topic: "test", Value: sarama.ByteEncoder(m)}

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			producer.Input() <- message
		}
	})
}

func BenchmarkKafkaGoAsync(b *testing.B) {
	b.ReportAllocs()
	w := &kafkago.Writer{
		Addr:         kafkago.TCP("localhost:29092"),
		Topic:        "test",
		Balancer:     &kafkago.LeastBytes{},
		Async:        true,
		RequiredAcks: kafkago.RequireAll,
	}

	c := context.Background()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		w.WriteMessages(c, kafkago.Message{Value: []byte(m)})
	}
}

func BenchmarkKafkaGoAsyncInParallel(b *testing.B) {
	b.ReportAllocs()
	w := &kafkago.Writer{
		Addr:         kafkago.TCP("localhost:29092"),
		Topic:        "test",
		Balancer:     &kafkago.LeastBytes{},
		Async:        true,
		RequiredAcks: kafkago.RequireAll,
	}

	c := context.Background()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w.WriteMessages(c, kafkago.Message{Value: []byte(m)})
		}
	})
}

func BenchmarkConfluentKafkaGoAsync(b *testing.B) {
	b.ReportAllocs()

	topic := "test"
	config := &ckafkago.ConfigMap{
		"bootstrap.servers": "localhost:29092",
		"acks":              "all",
	}
	p, err := ckafkago.NewProducer(config)
	if err != nil {
		panic(err)
	}

	go func() {
		for _ = range p.Events() {
		}
	}()

	key := []byte("")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.Produce(&ckafkago.Message{
			TopicPartition: ckafkago.TopicPartition{Topic: &topic, Partition: ckafkago.PartitionAny},
			Key:            key,
			Value:          m,
		}, nil)
	}
}

func BenchmarkConfluentKafkaGoAsyncInParallel(b *testing.B) {
	b.ReportAllocs()

	topic := "test"
	config := &ckafkago.ConfigMap{
		"bootstrap.servers": "localhost:29092",
		"acks":              "all",
	}
	p, err := ckafkago.NewProducer(config)
	if err != nil {
		panic(err)
	}

	go func() {
		for range p.Events() {
		}
	}()

	var mu sync.Mutex
	key := []byte("")
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mu.Lock()
			p.Produce(&ckafkago.Message{
				TopicPartition: ckafkago.TopicPartition{Topic: &topic, Partition: ckafkago.PartitionAny},
				Key:            key,
				Value:          m,
			}, nil)
			mu.Unlock()
		}
	})
}
