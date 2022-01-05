package main

import (
	"flag"
	"math/rand"
	"time"

	"github.com/apex/log"
)

var (
	client      string
	mode        string
	brokers     string
	topic       string
	partition   int
	msgSize     int
	numMessages int
	value       []byte
)

const (
	SessionTimeout   = 6 * time.Second
	CommitInterval   = 2 * time.Second
	RebalanceTimeout = 60 * time.Second
	MaxWait          = 250 * time.Millisecond
	MinBytes         = 1
	MaxBytes         = 5 << 20 // 5MB

	BatchTimeout = 10 * time.Millisecond
	BatchSize    = 100000
	BatchBytes   = 1048576
)

func main() {
	flag.StringVar(&brokers, "brokers", "kafka:9092", "broker addresses")
	flag.StringVar(&topic, "topic", "test", "set topic name")
	flag.IntVar(&partition, "partition", 0, "partition")
	flag.IntVar(&msgSize, "msgSize", 64, "message size")
	flag.IntVar(&numMessages, "numMessages", 100000, "number of messages")
	flag.StringVar(&client, "client", "confluent-kafka-go", "confluent-kafka-go / sarama / kafka-go")
	flag.StringVar(&mode, "mode", "producer", "producer / consumer")
	flag.Parse()
	value = make([]byte, msgSize)
	rand.Read(value)
	switch client {
	case "confluent-kafka-go":
		if mode == "producer" {
			produceConfluentKafkaGo()
		} else {
			consumeConfluentKafkaGo()
		}
		break
	case "sarama":
		if mode == "producer" {
			produceSarama()
		} else {
			consumeSarama()
		}
		break
	case "kafka-go":
		if mode == "producer" {
			produceKafkaGo()
		} else {
			consumeKafkaGo()
		}
		break
	default:
		log.Errorf("unknown client: %s", client)
	}
}
