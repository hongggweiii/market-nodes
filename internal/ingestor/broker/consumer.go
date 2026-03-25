package broker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hongggweiii/market-feed/internal/domain"
	"github.com/segmentio/kafka-go"
)

type KafkaConsumer struct {
	reader *kafka.Reader
}

func NewKafkaConsumer(brokerAddress string, topic string, groupId string) *KafkaConsumer {
	// Call factory function of kafka.Reader
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{brokerAddress},
		Topic:   topic,
		GroupID: groupId,
	})

	return &KafkaConsumer{
		reader: r,
	}
}

// Consume trade message, unmarshal and return them
func (c *KafkaConsumer) ConsumeTrade(ctx context.Context) (domain.Trade, error) {
	trade := new(domain.Trade)

	msg, err := c.reader.ReadMessage(ctx)
	if err != nil {
		return domain.Trade{}, err
	}

	err = json.Unmarshal(msg.Value, trade)
	if err != nil {
		fmt.Printf("Error while unserializing: %v", err)
		return domain.Trade{}, err
	}

	return *trade, nil
}
