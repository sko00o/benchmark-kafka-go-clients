package basic

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/Shopify/sarama"
	"github.com/Shopify/sarama/mocks"
)

type KafkaWriter struct {
	producer       sarama.AsyncProducer
	topic          string
	fallbackWriter io.WriteCloser
}

func (w *KafkaWriter) Write(b []byte) (n int, err error) {
	msg := &sarama.ProducerMessage{
		Topic: w.topic,
		Value: sarama.ByteEncoder(b),
	}

	select {
	case w.producer.Input() <- msg:
	default:
		// if producer block on input, write to fallbackWriter
		return w.fallbackWriter.Write(b)
	}
	return len(b), nil
}

func (w *KafkaWriter) Close() error {
	w.producer.AsyncClose()
	return w.fallbackWriter.Close()
}

func NewKafkaAsyncWriter(producer sarama.AsyncProducer, topic string, fallbackWriter io.WriteCloser) io.WriteCloser {
	w := &KafkaWriter{
		producer:       producer,
		topic:          topic,
		fallbackWriter: fallbackWriter,
	}

	go func() {
		// 将 producer 发送失败的消息写入 fallbackWriter
		for e := range producer.Errors() {
			val, err := e.Msg.Value.Encode()
			if err != nil {
				continue
			}

			_, _ = fallbackWriter.Write(val)
		}
	}()

	return w
}

type noCloser struct{ io.Writer }

func (n noCloser) Close() error { return nil }

func TestWriteFailWithSarama(t *testing.T) {
	config := sarama.NewConfig()
	p := mocks.NewAsyncProducer(t, config)

	buf := make([]byte, 0, 256)
	fallbackW := bytes.NewBuffer(buf)
	w := NewKafkaAsyncWriter(p, "test", noCloser{fallbackW})

	// 指定前两条请求失败
	p.ExpectInputAndFail(errors.New("produce error"))
	p.ExpectInputAndFail(errors.New("produce error"))

	w.Write([]byte("demo1"))
	w.Write([]byte("demo2"))
	w.Close()

	s := string(fallbackW.Bytes())
	if !strings.Contains(s, "demo1") {
		t.Errorf("want true, actual false")
	}
	if !strings.Contains(s, "demo2") {
		t.Errorf("want true, actual false")
	}

	if err := p.Close(); err != nil {
		t.Error(err)
	}
}
