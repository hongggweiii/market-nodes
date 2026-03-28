package main

import (
	"log"
	"time"

	"github.com/hongggweiii/market-feed/internal/domain"
	"github.com/hongggweiii/market-feed/internal/exchange"
	"github.com/hongggweiii/market-feed/internal/orderbook"
)

func main() {
	engine := orderbook.NewOrderBook()

	// Channel to receive depth updates from Websocket
	updates := make(chan *domain.DepthUpdate, 100)

	snapshot, err := exchange.FetchDepthSnapshot("BTCUSDT")
	if err != nil {
		log.Fatalf("Failed to fetch order book: %v", err)
	}

	engine.Seed(snapshot)

	log.Printf("Bids: %d, Asks: %d", len(engine.GetBids()), len(engine.GetAsks()))

	go func() {
		err := exchange.StreamOrderBookDepthUpdates("BTCUSDT", updates)
		if err != nil {
			log.Fatalf("Stream stopped: %v", err)
		}
	}()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case update := <-updates: // Channel receives new data from Websocket
			engine.ProcessUpdate(update)
		case <-ticker.C: // 1 second passed
			bestBidPrice, bestBidQty, bestAskPrice, bestAskQty := engine.GetTopBook()
			log.Printf("Best Bid: %s (%s), Best Ask: %s (%s)", bestBidPrice, bestBidQty, bestAskPrice, bestAskQty)
		}
	}

}
