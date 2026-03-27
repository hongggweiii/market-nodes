package orderbook

import (
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

	switch side {
	case "BID":
		b.bids[price] = quantity
	case "ASK":
		b.asks[price] = quantity
	}

	return nil
}

// DeleteLevel removes a price level from the order book
func (b *OrderBook) DeleteLevel(side string, price decimal.Decimal) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	switch side {
	case "BID":
		delete(b.bids, price)
	case "ASK":
		delete(b.asks, price)
	}

	return nil
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
