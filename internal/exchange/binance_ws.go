package exchange

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/hongggweiii/market-feed/internal/domain"
	"github.com/hongggweiii/market-feed/internal/ingestor/broker"
	"github.com/shopspring/decimal"
)

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
func StreamBinanceTrades(symbol string, broker *broker.KafkaProducer) error {
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

// StreamOrderBookDepthUpdates streams depth updates for a given symbol and sends them to the channel
func StreamOrderBookDepthUpdates(symbol string, updates chan<- *domain.DepthUpdate) error {
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
			Bids:          dto.Bids,
			Asks:          dto.Asks,
		}

		updates <- update
	}

	return nil
}
