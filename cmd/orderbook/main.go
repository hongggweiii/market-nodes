package main

import (
	"log"

	"github.com/hongggweiii/market-feed/internal/exchange"
)

func main() {

	snapshot, err := exchange.FetchDepthSnapshot("BTCUSDT")
	if err != nil {
		log.Fatalf("Failed to fetch order book: %v", err)
	}

	log.Printf("Snapshot Success! LastUpdateID: %d", snapshot.LastUpdateID)
	log.Printf("Bids: %d, Asks: %d", len(snapshot.Bids), len(snapshot.Asks))

	if len(snapshot.Bids) > 0 {
		log.Printf("First Bid: %v", snapshot.Bids[0])
	}
	if len(snapshot.Asks) > 0 {
		log.Printf("First Ask: %v", snapshot.Asks[0])
	}
}
