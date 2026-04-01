package main

import (
	"log"
	"time"

	"github.com/hongggweiii/market-feed/internal/domain"
	"github.com/hongggweiii/market-feed/internal/exchange"
	"github.com/hongggweiii/market-feed/internal/orderbook"
)

func RunOrderBook(api orderbook.DepthProvider, symbol string) {
	engine := orderbook.NewOrderBook()

	// Channel to receive depth updates from Websocket
	updates := make(chan *domain.DepthUpdate, 1000)

	go func() {
		err := api.StreamOrderBookDepthUpdates(symbol, updates)
		if err != nil {
			log.Fatalf("Stream stopped: %v", err)
		}
	}()

	log.Println("Websocket started...")

	snapshot, err := api.FetchDepthSnapshot(symbol)
	if err != nil {
		log.Fatalf("Failed to fetch order book: %v", err)
	}

	engine.Seed(snapshot)
	log.Printf("Engine seeded! Bids: %d, Asks: %d", len(engine.GetBids()), len(engine.GetAsks()))

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case update := <-updates: // Channel receives new data from Websocket
			err := engine.ProcessUpdate(update)
			if err != nil {
				log.Fatalf("Fatal sync error: %v", err)
			}
		case <-ticker.C: // 1 second passed
			bestBidPrice, bestBidQty, bestAskPrice, bestAskQty := engine.GetTopBook()
			log.Printf("Best Bid: %s (%s), Best Ask: %s (%s)", bestBidPrice, bestBidQty, bestAskPrice, bestAskQty)
		}
	}
}

func main() {
	binanceAPI := &exchange.BinanceClient{}

	log.Println("Starting order book...")
	RunOrderBook(binanceAPI, "BTCUSDT")

}
