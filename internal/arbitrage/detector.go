// Provides functionality for detecting arbitrage opportunities
package arbitrage

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	orderbookpb "github.com/hongggweiii/market-nodes/api/proto"
	"github.com/shopspring/decimal"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Detector monitors order books across multiple exchanges to identify arbitrage opportunities
type Detector struct {
	binanceClient orderbookpb.OrderBookServiceClient
	bybitClient   orderbookpb.OrderBookServiceClient
}

func NewDetector(binance, bybit string) (*Detector, error) {
	binanceConn, err := grpc.NewClient(binance, grpc.WithTransportCredentials(insecure.NewCredentials())) // Insecure for local development
	if err != nil {
		return nil, err
	}

	bybitConn, err := grpc.NewClient(bybit, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &Detector{
		binanceClient: orderbookpb.NewOrderBookServiceClient(binanceConn),
		bybitClient:   orderbookpb.NewOrderBookServiceClient(bybitConn),
	}, nil
}

// Start begins monitoring a trading symbol
// Checks price spreads between Binance and Bybit at 100ms intervals continuously
func (d *Detector) Start(symbol string) {
	log.Printf("Starting arbitrage detector for %s...\n", symbol)

	// Check spreads every 100 milliseconds
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		err := d.checkSpread(symbol, &sync.WaitGroup{})
		if err != nil {
			log.Printf("Error checking spread for %s: %v", symbol, err)
		}
	}
}

// checkSpread fetches the best bid/ask prices from both exchanges and identifies arbitrage opportunities
func (d *Detector) checkSpread(symbol string, wg *sync.WaitGroup) error {
	// Create a context with timeout for gRPC calls to prevent hanging indefinitely
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var binanceResp, bybitResp *orderbookpb.GetTopBookResponse
	var binanceErr, bybitErr error

	wg.Add(2)

	// Fetch Binance order book concurrently
	go func() {
		defer wg.Done()
		binanceResp, binanceErr = d.binanceClient.GetOrderBook(ctx, &orderbookpb.GetTopBookRequest{Symbol: symbol})
		if binanceErr != nil {
			log.Printf("Error fetching Binance order book: %v", binanceErr)
			return
		}
	}()

	// Fetch Bybit order book concurrently
	go func() {
		defer wg.Done()
		bybitResp, bybitErr = d.bybitClient.GetOrderBook(ctx, &orderbookpb.GetTopBookRequest{Symbol: symbol})
		if bybitErr != nil {
			log.Printf("Error fetching Bybit order book: %v", bybitErr)
			return
		}
	}()

	// Wait for both goroutines to complete
	wg.Wait()

	if binanceErr != nil || bybitErr != nil {
		return errors.New("Failed to fetch order book data from one or both exchanges")
	}

	binanceBid, _ := decimal.NewFromString(binanceResp.GetBestBidPrice())
	binanceAsk, _ := decimal.NewFromString(binanceResp.GetBestAskPrice())
	bybitBid, _ := decimal.NewFromString(bybitResp.GetBestBidPrice())
	bybitAsk, _ := decimal.NewFromString(bybitResp.GetBestAskPrice())

	var foundProfit bool
	// Opportunity 1: Binance bid > Bybit ask (long on Bybit, short on Binance)
	if binanceBid.GreaterThan(bybitAsk) {
		log.Printf("Profit Opportunity: Buy on Bybit at %s and sell on Binance at %s", bybitResp.BestAskPrice, binanceResp.BestBidPrice)
		foundProfit = true
	}

	// Opportunity 2: Bybit bid > Binance ask (long on Binance, short on Bybit)
	if bybitBid.GreaterThan(binanceAsk) {
		log.Printf("Profit Opportunity: Buy on Binance at %s and sell on Bybit at %s", binanceResp.BestAskPrice, bybitResp.BestBidPrice)
		foundProfit = true
	}

	// No opportunity found
	if !foundProfit {
		log.Printf("No arbitrage opportunity. Binance Bid: %s, Ask: %s | Bybit Bid: %s, Ask: %s", binanceBid, binanceAsk, bybitBid, bybitAsk)
	}

	if ctx.Err() != nil {
		log.Println("Timeout fetching prices from order book services")
		return ctx.Err()
	}
	return nil
}
