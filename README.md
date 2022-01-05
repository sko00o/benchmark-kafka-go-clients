# Benchmark for Kafka clients in Go

## how to run benchmark

Pull up docker environment for test.

`make compose-up`

Open stats monitor on worker.

`docker stats kafka-benchmark-worker-1`

Stop test environment.

`make compose-down`

Run test

```shell
docker compose exec worker \
  ./kafka-benchmark -numMessages 1000000 -msgSize 512
docker compose exec worker \
  ./kafka-benchmark -numMessages 1000000 -msgSize 512 -mode consumer
docker compose exec worker \
  ./kafka-benchmark -numMessages 1000000 -msgSize 512 -client kafka-go
docker compose exec worker \
  ./kafka-benchmark -numMessages 1000000 -msgSize 512 -client kafka-go -mode consumer
docker compose exec worker \
  ./kafka-benchmark -numMessages 1000000 -msgSize 512 -client sarama
docker compose exec worker \
  ./kafka-benchmark -numMessages 1000000 -msgSize 512 -client sarama -mode consumer
```

## test environment

docker compose:
* CPUs: 4
* Memory: 5.00 GB

## test report

### test case 1

* numMessages = 10,000,000
* msgSize = 64

| client    | producer (msg/s) | consumer (msg/s) |
|-----------|------------------|------------------|
| confluent | 28553            | 62435            |
| kafka-go  | 481932           | 50983            |
| sarama    | 37493            | 1044443          |

| client    | producer MEM Usage (MiB) | consumer MEM Usage (MiB) |
|-----------|--------------------------|--------------------------|
| confluent | 47                       | 47                       |
| kafka-go  | 200                      | 9                        |
| sarama    | 10                       | 18                       |

### test case 2

* numMessages = 1,000,000
* msgSize = 512

| client    | producer (msg/s) | consumer (msg/s) |
|-----------|------------------|------------------|
| confluent | 26446            | 60610            |
| kafka-go  | 18073            | 44491            |
| sarama    | 32256            | 313556           |

| client    | producer MEM Usage (MiB) | consumer MEM Usage (MiB) |
|-----------|--------------------------|--------------------------|
| confluent | 47                       | 67                       |
| kafka-go  | 977                      | 10                       |
| sarama    | 10                       | 14                       |

## conclusion

* consumer:
    1. sarama
    2. confluent
    3. kafka-go
* producer:
    1. sarama
    2. kafka-go (memory use too much, slow on big message)
    3. confluent

## reference

* [kafka go client performance testing](https://gist.github.com/mhowlett/e9491aad29817aeda6003c3404874b35)
* [librdkafka configuration properties](https://github.com/edenhill/librdkafka/tree/master/CONFIGURATION.md)
* [sarama/examples/consumergroup](https://github.com/Shopify/sarama/blob/main/examples/consumergroup/main.go)
* [sarama/tool/kafka-producer-performance](https://github.com/Shopify/sarama/blob/main/tools/kafka-producer-performance/main.go)
