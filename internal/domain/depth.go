package domain

import "github.com/shopspring/decimal"

// REST
type DepthSnapshot struct {
	LastUpdateID int64               // Last update ID
	Bids         [][]decimal.Decimal // Bids
	Asks         [][]decimal.Decimal // Asks

}

// Websocket
type DepthUpdate struct {
	EventType     string                     // Event type
	EventTime     int64                      // Event time
	Symbol        string                     // Symbol
	FirstUpdateID int64                      // 1st update ID
	FinalUpdateID int64                      // 2nd update ID
	Bids          map[string]decimal.Decimal // Bids
	Asks          map[string]decimal.Decimal // Asks
}
