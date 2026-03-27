package orderbook

import (
	"sync"

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
