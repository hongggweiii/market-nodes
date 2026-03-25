package exchange

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/hongggweiii/market-feed/internal/domain"
	"github.com/hongggweiii/market-feed/internal/ingestor/broker"
)

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

		trade := new(domain.Trade)
		err = json.Unmarshal(p, trade)
		if err != nil {
			fmt.Println("Error while unserializing:", err)
			continue
		}

		err = broker.PublishTrade(*trade)
		if err != nil {
			fmt.Printf("Error while publishing trade: %v", err)
		}

		fmt.Println("Published trade:", trade)
	}

	return nil
}
