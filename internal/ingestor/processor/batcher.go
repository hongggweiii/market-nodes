package processor

import (
	"context"
	"fmt"
	"time"

	"github.com/hongggweiii/market-feed/internal/domain"
	"github.com/hongggweiii/market-feed/internal/ingestor/broker"
	"github.com/hongggweiii/market-feed/internal/ingestor/database"
)

func StartBatchingEngine(ctx context.Context, consumer *broker.KafkaConsumer, repo *database.ClickHouseRepo) error {
	// Channel to pass trades between threads
	tradeChan := make(chan domain.Trade, 1000)

	go func() {
		for {
			trade, err := consumer.ConsumeTrade(ctx)
			if err != nil {
				fmt.Printf("Error while consuming trade: %v", err)
				continue
			}
			tradeChan <- trade // Send trade into channel
		}
	}() // Immediately executes, anonymous func

	var batch []domain.Trade
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		// Wait on multiple channels. Proceeds with case whose channel operation is ready first
		select {
		case trade := <-tradeChan: // Append trade to batch
			batch = append(batch, trade)
			if len(batch) >= 1000 {
				err := repo.InsertTrades(ctx, batch)
				if err != nil {
					fmt.Printf("Error while inserting batch to ClickHouse: %v", err)
				}
				batch = nil
			}
		case <-ticker.C: // Insert existing batch to database every 2s
			if len(batch) > 0 {
				err := repo.InsertTrades(ctx, batch)
				if err != nil {
					fmt.Printf("Error while inserting batch to ClickHouse: %v", err)
				}
				batch = nil
			}
		}
	}
}
