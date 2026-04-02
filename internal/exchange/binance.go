package exchange

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/hongggweiii/market-feed/internal/domain"
	"github.com/hongggweiii/market-feed/internal/ingestor/broker"
	"github.com/shopspring/decimal"
)

type BinanceClient struct{}

type binanceDepthSnapshotDTO struct {
	LastUpdateID int64               `json:"lastUpdateId"`
	Bids         [][]decimal.Decimal `json:"bids"`
	Asks         [][]decimal.Decimal `json:"asks"`
}

type binanceTradeDTO struct {
	EventType     string          `json:"e"`
	EventTime     int64           `json:"E"`
	Symbol        string          `json:"s"`
	TradeID       int64           `json:"t"`
	Price         decimal.Decimal `json:"p"`
	Quantity      decimal.Decimal `json:"q"`
	TradeTime     int64           `json:"T"`
	IsMarketMaker bool            `json:"m"`
}

type binanceDepthUpdateDTO struct {
	EventType     string              `json:"e"`
	EventTime     int64               `json:"E"`
	Symbol        string              `json:"s"`
	FirstUpdateID int64               `json:"U"`
	FinalUpdateID int64               `json:"u"`
	Bids          [][]decimal.Decimal `json:"b"`
	Asks          [][]decimal.Decimal `json:"a"`
}

// StreamBinanceTrades streams trades for a given symbol and publishes them to the provided Kafka producer
func (c *BinanceClient) StreamBinanceTrades(symbol string, broker *broker.KafkaProducer) error {
	baseUrl := "wss://stream.binance.com:9443/ws"
	lowercaseSymbol := strings.ToLower(symbol)
	websocketUrl := fmt.Sprintf("%s/%s@trade", baseUrl, lowercaseSymbol)

	// Initialise Websocket connection
	conn, _, err := websocket.DefaultDialer.Dial(websocketUrl, nil)
	if err != nil {
		fmt.Printf("Error while initialising Websocket connection: %v", err)
		return err
	}
	defer conn.Close()

	// Infinite loop for Websocket
	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error while reading message:", err)
			break
		}

		dto := new(binanceTradeDTO)
		err = json.Unmarshal(p, dto)
		if err != nil {
			fmt.Println("Error while unserializing:", err)
			continue
		}

		trade := &domain.Trade{
			EventType:     dto.EventType,
			EventTime:     dto.EventTime,
			Symbol:        dto.Symbol,
			TradeID:       dto.TradeID,
			Price:         dto.Price,
			Quantity:      dto.Quantity,
			TradeTime:     dto.TradeTime,
			IsMarketMaker: dto.IsMarketMaker,
		}

		err = broker.PublishTrade(*trade)
		if err != nil {
			fmt.Printf("Error while publishing trade: %v", err)
		}

		fmt.Println("Published trade:", trade)
	}

	return nil
}

// FetchDepthSnapshot fetches the current order book snapshot for a given symbol
func (c *BinanceClient) FetchDepthSnapshot(symbol string) (*domain.DepthSnapshot, error) {
	const limit = 1000
	baseUrl := "https://api.binance.com"
	lowercaseSymbol := strings.ToUpper(symbol)
	restUrl := fmt.Sprintf("%s/api/v3/depth?symbol=%s&limit=%d", baseUrl, lowercaseSymbol, limit)

	resp, err := http.Get(restUrl)
	if err != nil {
		fmt.Println("Error fetching depth:", err)
		return nil, err
	}
	defer resp.Body.Close() // Prevent resource leaks

	dto := new(binanceDepthSnapshotDTO)
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
func (c *BinanceClient) StreamOrderBookDepthUpdates(symbol string, updates chan<- *domain.DepthUpdate) error {
	baseUrl := "wss://stream.binance.com:9443/ws"
	lowercaseSymbol := strings.ToLower(symbol)
	websocketUrl := fmt.Sprintf("%s/%s@depth", baseUrl, lowercaseSymbol)

	// Initialise Websocket connection
	conn, _, err := websocket.DefaultDialer.Dial(websocketUrl, nil)
	if err != nil {
		fmt.Printf("Error while initialising Websocket connection: %v", err)
		return err
	}
	defer conn.Close()

	fmt.Printf("[Binance] Connected and subscribed to %s", symbol)

	// Infinite loop for Websocket
	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error while reading message:", err)
			break
		}

		dto := new(binanceDepthUpdateDTO)
		err = json.Unmarshal(p, dto)
		if err != nil {
			fmt.Println("Error while unserializing:", err)
			continue
		}

		update := &domain.DepthUpdate{
			EventType:     dto.EventType,
			EventTime:     dto.EventTime,
			Symbol:        dto.Symbol,
			FirstUpdateID: dto.FirstUpdateID,
			FinalUpdateID: dto.FinalUpdateID,
			Bids:          make(map[string]decimal.Decimal),
			Asks:          make(map[string]decimal.Decimal),
		}

		// Populate the Bids and Asks maps
		for _, bid := range dto.Bids {
			priceStr := bid[0].String() // Convert the price decimal to a string key
			quantity := bid[1]
			update.Bids[priceStr] = quantity
		}

		for _, ask := range dto.Asks {
			priceStr := ask[0].String()
			quantity := ask[1]
			update.Asks[priceStr] = quantity
		}

		updates <- update
	}

	return nil
}
