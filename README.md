# MarketNodes

Market Nodes is a comprehensive backend engineering project which serves as a learning laboratory for exploring complex distributed systems, high-throughput data pipelines, and advanced microservice architecture. This project aims to tackle concurrency management, strict state synchronisation, event-driven messaging and high speed inter-service communication.

## The Core Microservices

1. **The Ingestor**
2. **The Order Book**
3. **The Arbitrage Detector**

---

## Microservice 1: Market Feed Ingestor

A high-throughput data pipeline built in Go. This service streams real-time cryptocurrency trade data from the Binance WebSocket API, buffers the events through Apache Kafka, and utilises an in-memory batching engine to write financial data into a ClickHouse OLAP database.

## Architecture Overview

1. **Producer:** Connects to the Binance WebSocket (`BTCUSDT` by default) and streams raw trades.
2. **Message Broker:** Pushes trades to a Kafka topic (`crypto.trades.raw`) to decouple ingestion from storage and handle sudden traffic spikes.
3. **Batching Consumer:** Reads from Kafka and accumulates trades in memory.
4. **Storage:** Flushes batches to ClickHouse every 2 seconds or 1,000 trades

## Setup & Configuration

**1. Clone the repository**
```bash
git clone [https://github.com/yourusername/market-nodes.git](https://github.com/yourusername/market-nodes.git)
cd market-nodes
```
**2. Configure Env Variables**
Create a .env file from .env.example
```bash
cp .env.example .env
```
Ensure your .env has the following values to run locally:
```bash
RUN_MIGRATIONS=true
KAFKA_BROKER=localhost:9092
CLICKHOUSE_ADDR=localhost:9000
```
**3. Install Dependencies**
```bash
go mod tidy
```

## Running the Infrastructure
Start local Docker Compose:
```bash
docker compose up -d
```

## Database Migrations
By setting `RUN_MIGRATIONS=true`, app will execute database schema SQL files to build Clickhouse tables. 

If you want to run it manually using CLI:
```bash
migrate -path db/migrations -database "clickhouse://default:@localhost:9000/default" up
```

## Running Application
```bash
go run cmd/ingestor/main.go
```

## Accessing ClickHouse terminal
```bash
docker exec -it clickhouse-db clickhouse-client
```

---

## Microservice 2: Order Book Engine

A highly performant, concurrency-safe, real-time order book replica built in Go. This service connects to the Binance WebSocket and REST APIs to maintain a mathematically accurate, millisecond-level state of the market depth for BTCUSDT.

## Architecture Overview

1. **Websocket Buffer:** Spawns a background Goroutine to connect to Binance's Websocket, buffering live depth trades.
2. **REST Snapshot Seeding:** Fetch a massive depth snapshot via HTTP REST, seeding the initial `bids` and `asks` maps
3. **Sequence Sync Engine:** Stitches the stream to the snapshot. It drops obsolete packets, identifies the exact overlap point, and enforces strict sequence continuity.
4. **Concurrency-Safe Memory:** Utilises Go maps protected by `sync.RWMutex` to safely process thousands of bid and ask insertions/deletions per second without race conditions.

## Setup & Configuration

**1. Clone the repository**
```bash
git clone [https://github.com/yourusername/market-nodes.git](https://github.com/yourusername/market-nodes.git)
cd market-nodes
```
**2. Install Dependencies**
```bash
go mod tidy
```

## Running Application
```bash
go run cmd/orderbook/main.go
```

---

## Microservice 3: Arbitrage Detector

Functions as a gRPC client which continuously polls multiple exchange nodes simultaneously, calculates the exact cross-exchange spread, and logs profitable arbitrage pathways. 

## Architecture Overview

1. **Concurrent gRPC Clients:** Utilizes Protobuf clients to establish persistent HTTP/2 connections to the local Order Book servers. Allows for highly optimized, binary communication rather than relying on slow HTTP/JSON requests.
2. **Synchronized Polling (`sync.WaitGroup`):** To eliminate sequential latency (where one exchange's price becomes outdated while waiting for the other)

## Setup & Configuration

**1. Clone the repository**
```bash
git clone [https://github.com/yourusername/market-nodes.git](https://github.com/yourusername/market-nodes.git)
cd market-nodes
```
**2. Install Dependencies**
```bash
go mod tidy
```

## Running Application
Run OrderBook first before running Arbitrage, since gRPC server is located in OrderBook
```bash
go run cmd/orderbook/main.go
go run cmd/arbitrage/main.go
```