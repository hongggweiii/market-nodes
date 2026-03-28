package orderbook

import (
	"maps"
	"slices"
	"sync"

	"github.com/hongggweiii/market-feed/internal/domain"
	"github.com/shopspring/decimal"
)

type OrderBook struct {
	mu   sync.RWMutex // Allow multiple reads but only 1 writes
	bids map[decimal.Decimal]decimal.Decimal
	asks map[decimal.Decimal]decimal.Decimal
}

func NewOrderBook() *OrderBook {
	return &OrderBook{
		bids: make(map[decimal.Decimal]decimal.Decimal),
		asks: make(map[decimal.Decimal]decimal.Decimal),
	}
}

// UpdateLevel updates or adds a price level in the order book
func (b *OrderBook) UpdateLevel(side string, price decimal.Decimal, quantity decimal.Decimal) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.updateLevelUnsafe(side, price, quantity)
}

// DeleteLevel removes a price level from the order book
func (b *OrderBook) DeleteLevel(side string, price decimal.Decimal) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.deleteLevelUnsafe(side, price)
}

// Seed initialises the order book with a snapshot of the current market depth
func (b *OrderBook) Seed(snapshot *domain.DepthSnapshot) {
	// Writing to the maps, so we need a write lock
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, bid := range snapshot.Bids {
		price := bid[0]
		quantity := bid[1]
		b.bids[price] = quantity
	}

	for _, ask := range snapshot.Asks {
		price := ask[0]
		quantity := ask[1]
		b.asks[price] = quantity
	}
}

// ProcessUpdate applies a depth update to the order book, updating or deleting levels
func (b *OrderBook) ProcessUpdate(update *domain.DepthUpdate) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, bid := range update.Bids {
		price := bid[0]
		quantity := bid[1]
		if quantity.IsZero() {
			b.deleteLevelUnsafe("BID", price)
		} else {
			b.updateLevelUnsafe("BID", price, quantity)
		}
	}

	for _, ask := range update.Asks {
		price := ask[0]
		quantity := ask[1]
		if quantity.IsZero() {
			b.deleteLevelUnsafe("ASK", price)
		} else {
			b.updateLevelUnsafe("ASK", price, quantity)
		}
	}
}

// Unsafe versions of the update and delete functions that assume the caller has already acquired the necessary locks
func (b *OrderBook) updateLevelUnsafe(side string, price decimal.Decimal, quantity decimal.Decimal) error {
	switch side {
	case "BID":
		b.bids[price] = quantity
	case "ASK":
		b.asks[price] = quantity
	}

	return nil
}

func (b *OrderBook) deleteLevelUnsafe(side string, price decimal.Decimal) error {
	switch side {
	case "BID":
		delete(b.bids, price)
	case "ASK":
		delete(b.asks, price)
	}

	return nil
}

// getTopBook returns the best bid and ask price and quantity for the current order book state
func (b *OrderBook) GetTopBook() (decimal.Decimal, decimal.Decimal, decimal.Decimal, decimal.Decimal) {
	bids := b.GetBids()
	asks := b.GetAsks()

	bidPrices := slices.Collect(maps.Keys(bids))
	askPrices := slices.Collect(maps.Keys(asks))

	var bestBidPrice, bestBidQty, bestAskPrice, bestAskQty decimal.Decimal

	if len(bidPrices) > 0 {
		slices.SortFunc(bidPrices, func(a, b decimal.Decimal) int {
			return b.Cmp(a)
		})
		bestBidPrice = bidPrices[0]
		bestBidQty = bids[bestBidPrice]
	}

	if len(askPrices) > 0 {
		slices.SortFunc(askPrices, func(a, b decimal.Decimal) int {
			return a.Cmp(b)
		})
		bestAskPrice = askPrices[0]
		bestAskQty = asks[bestAskPrice]
	}

	return bestBidPrice, bestBidQty, bestAskPrice, bestAskQty
}

// GetBids and GetAsks return copies of the current order book state to prevent external modification
func (b *OrderBook) GetBids() map[decimal.Decimal]decimal.Decimal {
	b.mu.RLock()
	defer b.mu.RUnlock()

	// Return a copy to prevent external modification
	copy := make(map[decimal.Decimal]decimal.Decimal)
	for k, v := range b.bids {
		copy[k] = v
	}
	return copy
}

func (b *OrderBook) GetAsks() map[decimal.Decimal]decimal.Decimal {
	b.mu.RLock()
	defer b.mu.RUnlock()

	// Return a copy to prevent external modification
	copy := make(map[decimal.Decimal]decimal.Decimal)
	for k, v := range b.asks {
		copy[k] = v
	}
	return copy
}
