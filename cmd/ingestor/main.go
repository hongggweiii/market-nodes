package main

import (
	"context"
	"log"

	"github.com/hongggweiii/market-feed/internal/broker"
	"github.com/hongggweiii/market-feed/internal/database"
	"github.com/hongggweiii/market-feed/internal/exchange"
	"github.com/hongggweiii/market-feed/internal/processor"
)

func main() {
	const BrokerAddress = "localhost:9092"
	const KafkaTopic = "crypto.trades.raw"
	const ClickHouseAddress = "localhost:9000"

	ctx := context.Background()
	consumer := broker.NewKafkaConsumer(BrokerAddress, KafkaTopic, "trades-clickhouse-ingestor")
	producer := broker.NewKafkaProducer(BrokerAddress, KafkaTopic)
	repo := database.NewClickHouseRepo(ClickHouseAddress)

	err := broker.PrepareKafkaTopic(BrokerAddress, KafkaTopic)
	if err != nil {
		log.Fatalf("Failed to create Kafka topic: %v", err)
	}

	// Websocket connections are blocking
	go func() {
		err := exchange.StreamBinanceTrades("BTCUSDT", producer)
		if err != nil {
			log.Fatalf("Stream stopped: %v", err)
		}
	}()

	err = processor.StartBatchingEngine(ctx, consumer, repo)
	if err != nil {
		log.Fatalf("Failed to start batching engine: %v", err)
	}
}
