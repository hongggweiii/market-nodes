package main

import (
	"log"

	"github.com/hongggweiii/market-nodes/internal/arbitrage"
)

func main() {
	detector, err := arbitrage.NewDetector("localhost:50001", "localhost:50002")
	if err != nil {
		log.Fatalf("Failed to create detector: %v", err)
	}

	detector.Start("BTCUSDT")
}
