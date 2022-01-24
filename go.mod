module github.com/sko00o/benchmark-kafka-go-clients

go 1.16

require (
	github.com/Shopify/sarama v1.30.1
	github.com/apex/log v1.9.0
	github.com/confluentinc/confluent-kafka-go v1.8.2
	github.com/segmentio/kafka-go v0.4.27
)

replace github.com/segmentio/kafka-go v0.4.27 => github.com/sko00o/kafka-go v0.4.26-0.20220119091327-4f9c8294575b
