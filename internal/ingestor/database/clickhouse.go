package database

import (
	"context"
	"fmt"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/hongggweiii/market-feed/internal/domain"
	"github.com/shopspring/decimal"
)

type ClickHouseRepo struct {
	conn clickhouse.Conn
}

func NewClickHouseRepo(brokerAddress string) *ClickHouseRepo {
	// Initialise a connection
	c, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{brokerAddress},
		Auth: clickhouse.Auth{
			Database: "default",
			Username: "default",
			Password: "",
		},
	})

	if err = c.Ping(context.Background()); err != nil {
		fmt.Printf("Error connecting to ClickHouse database: %v", err)
	}

	return &ClickHouseRepo{
		conn: c,
	}
}

func (repo *ClickHouseRepo) InsertTrades(ctx context.Context, trades []domain.Trade) error {
	batch, err := repo.conn.PrepareBatch(ctx, "INSERT INTO trades (event_type, event_time, symbol, trade_id, price, quantity, trade_time, is_market_maker)")
	if err != nil {
		fmt.Printf("Error inserting trades batch into database: %v", err)
		return err
	}

	fmt.Printf("Flushing batch of size %d to ClickHouse...", len(trades))

	for i := range trades {
		// Prevent floating point errors
		priceDec, _ := decimal.NewFromString(trades[i].Price)
		qtyDec, _ := decimal.NewFromString(trades[i].Quantity)

		err := batch.Append(
			trades[i].EventType,
			trades[i].EventTime,
			trades[i].Symbol,
			trades[i].TradeID,
			priceDec,
			qtyDec,
			trades[i].TradeTime,
			trades[i].IsMarketMaker,
		)
		if err != nil {
			fmt.Printf("Error appending trade to batch: %v", err)
			continue
		}
	}

	// Send to ClickHouse after all trades are flushed to batch
	return batch.Send()
}
