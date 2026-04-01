package main

import (
	"context"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	// Blank import to run init() in the background for src and dest file paths
	_ "github.com/golang-migrate/migrate/v4/database/clickhouse"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/hongggweiii/market-feed/internal/exchange"
	"github.com/hongggweiii/market-feed/internal/ingestor/broker"
	"github.com/hongggweiii/market-feed/internal/ingestor/database"
	"github.com/hongggweiii/market-feed/internal/ingestor/processor"
	"github.com/joho/godotenv"
)

type TradeStreamer interface {
	StreamBinanceTrades(symbol string, broker *broker.KafkaProducer) error
}

func RunIngestor(ctx context.Context, api TradeStreamer, producer *broker.KafkaProducer, consumer *broker.KafkaConsumer, repo *database.ClickHouseRepo, symbol string) {
	// Websocket connections are blocking, so we run it in a separate goroutine
	go func() {
		err := api.StreamBinanceTrades(symbol, producer)
		if err != nil {
			log.Fatalf("Stream stopped: %v", err)
		}
	}()

	// Starts the batching engine and blocks the main thread
	err := processor.StartBatchingEngine(ctx, consumer, repo)
	if err != nil {
		log.Fatalf("Failed to start batching engine: %v", err)
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		// Not fatal, since production variables not set in Docker/OS
		log.Println("Error loading .env file, using system env")
	}

	var BrokerAddress = os.Getenv("KAFKA_BROKER")
	const KafkaTopic = "crypto.trades.raw"
	var ClickHouseAddress = os.Getenv("CLICKHOUSE_ADDR")

	// Saves resources and prevents deadlocks
	if os.Getenv("RUN_MIGRATIONS") == "true" {
		log.Println("Running embedded database migrations...")

		m, err := migrate.New(
			"file://db/migrations",                         // Source: SQL migration files
			"clickhouse://default:@localhost:9000/default", // Destination: Database connection string
		)
		if err != nil {
			log.Fatalf("Migration setup failed: %v", err)
		}

		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("An error occurred while syncing the database: %v", err)
		}
		log.Println("Database successfully migrated!")
	}

	err = broker.PrepareKafkaTopic(BrokerAddress, KafkaTopic)
	if err != nil {
		log.Fatalf("Failed to create Kafka topic: %v", err)
	}

	ctx := context.Background()
	consumer := broker.NewKafkaConsumer(BrokerAddress, KafkaTopic, "trades-clickhouse-ingestor")
	producer := broker.NewKafkaProducer(BrokerAddress, KafkaTopic)
	repo := database.NewClickHouseRepo(ClickHouseAddress)
	binanceAPI := &exchange.BinanceClient{}

	log.Println("Booting Ingestor...")
	RunIngestor(ctx, binanceAPI, producer, consumer, repo, "BTCUSDT")
}
