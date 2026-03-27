package main

import (
	"log"

	"github.com/hongggweiii/market-feed/internal/exchange"
	"github.com/hongggweiii/market-feed/internal/orderbook"
)

func main() {
	engine := orderbook.NewOrderBook()

	snapshot, err := exchange.FetchDepthSnapshot("BTCUSDT")
	if err != nil {
		log.Fatalf("Failed to fetch order book: %v", err)
	}

	engine.Seed(snapshot)

	log.Printf("Bids: %d, Asks: %d", len(engine.GetBids()), len(engine.GetAsks()))
}
