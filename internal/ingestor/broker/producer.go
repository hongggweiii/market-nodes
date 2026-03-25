package broker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hongggweiii/market-feed/internal/domain"
	"github.com/segmentio/kafka-go"
)

type KafkaProducer struct {
	writer *kafka.Writer
}

func NewKafkaProducer(brokerAddress string, topic string) *KafkaProducer {
	return &KafkaProducer{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(brokerAddress),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{}, // Good default for routing messages
		},
	}
}

func (p *KafkaProducer) PublishTrade(trade domain.Trade) error {
	ctx := context.Background()

	rawJSON, err := json.Marshal(trade)
	if err != nil {
		fmt.Printf("Error while serializing: %v", err)
		return err
	}

	msg := kafka.Message{
		Value: rawJSON,
	}

	// Write parsed trades to Kafka
	err = p.writer.WriteMessages(ctx, msg)
	if err != nil {
		fmt.Printf("Failed to write message: %v", err)
		return err
	}

	return nil
}
