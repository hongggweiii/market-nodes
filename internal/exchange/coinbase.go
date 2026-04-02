package exchange

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hongggweiii/market-feed/internal/domain"
	"github.com/shopspring/decimal"
)

type CoinbaseClient struct{}

type coinbaseDepthSnapshotDTO struct {
	LastUpdateID int64               `json:"lastUpdateId"`
	Bids         [][]decimal.Decimal `json:"bids"`
	Asks         [][]decimal.Decimal `json:"asks"`
}

type coinbaseWSSubscribeDTO struct {
	Type       string   `json:"type"`
	ProductIDs []string `json:"product_ids"`
	Channels   []string `json:"channels"`
}

type coinbaseDepthUpdateDTO struct {
	Type    string     `json:"type"`
	Changes [][]string `json:"changes"` // ["buy" or "sell", price, size]
}

// FetchDepthSnapshot fetches the current order book snapshot for a given symbol
func (c *CoinbaseClient) FetchDepthSnapshot(symbol string) (*domain.DepthSnapshot, error) {
	const limit = 1000
	baseUrl := "https://api.pro.coinbase.com/products"
	restUrl := fmt.Sprintf("%s/%s/book?level=2", baseUrl, symbol)

	resp, err := http.Get(restUrl)
	if err != nil {
		fmt.Println("Error fetching depth:", err)
		return nil, err
	}
	defer resp.Body.Close() // Prevent resource leaks

	dto := new(coinbaseDepthSnapshotDTO)
	if resp.StatusCode == http.StatusOK {
		// io.ReadAll() take sup lots of memory
		if err := json.NewDecoder(resp.Body).Decode(dto); err != nil {
			return nil, fmt.Errorf("Failed to decode response: %w", err)
		}
	} else {
		return nil, fmt.Errorf("Request failed with status: %d", resp.StatusCode)
	}

	// Map DTO to Domain model
	snapshot := &domain.DepthSnapshot{
		LastUpdateID: dto.LastUpdateID,
		Bids:         dto.Bids,
		Asks:         dto.Asks,
	}

	fmt.Println("Successful fetch!")
	return snapshot, nil
}

// StreamOrderBookDepthUpdates streams depth updates for a given symbol and sends them to the channel
func (c *CoinbaseClient) StreamOrderBookDepthUpdates(symbol string, updates chan<- *domain.DepthUpdate) error {
	wsUrl := "wss://ws-feed.exchange.coinbase.com"

	// Initialise Websocket connection
	conn, _, err := websocket.DefaultDialer.Dial(wsUrl, nil)
	if err != nil {
		fmt.Printf("Error while initialising Websocket connection: %v", err)
		return err
	}
	defer conn.Close()

	// Send subscription message to Websocket
	subMsg := coinbaseWSSubscribeDTO{
		Type:       "subscribe",
		ProductIDs: []string{symbol},
		Channels:   []string{"level2"},
	}
	if err := conn.WriteJSON(subMsg); err != nil {
		fmt.Printf("Error while subscribing to channel: %v", err)
		return err
	}

	fmt.Printf("[Coinbase] Connected and subscribed to %s", symbol)

	// Infinite loop for Websocket
	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error while reading message:", err)
			break
		}

		dto := new(coinbaseDepthUpdateDTO)
		err = json.Unmarshal(p, dto)
		if err != nil {
			fmt.Println("Error while unserializing:", err)
			continue
		}

		// Parse the depth update changes
		bids := make(map[string]decimal.Decimal)
		asks := make(map[string]decimal.Decimal)

		for _, change := range dto.Changes {
			if len(change) < 3 {
				continue
			}
			side := change[0]    // "buy" or "sell"
			price := change[1]   // price as string
			sizeStr := change[2] // size as string

			size, err := decimal.NewFromString(sizeStr)
			if err != nil {
				continue
			}

			switch side {
			case "buy":
				bids[price] = size
			case "sell":
				asks[price] = size
			}
		}

		update := &domain.DepthUpdate{
			EventType:     "depth",                // Coinbase doesn't send explicit event type
			EventTime:     time.Now().UnixMilli(), // Coinbase WS doesn't send time, so we attach it
			Symbol:        symbol,
			FirstUpdateID: 0, // Coinbase doesn't use sequence IDs
			FinalUpdateID: 0,
			Bids:          bids,
			Asks:          asks,
		}

		updates <- update
	}

	return nil
}
