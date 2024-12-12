package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	_ "github.com/lib/pq"
)

var (
	db          *sql.DB
	dataChannel = make(chan string, 1000) // Buffered channel
)

func initDB() {
	var err error
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "127.0.0.1"
	}

	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = "123456789"
	}

	db, err = sql.Open("postgres", fmt.Sprintf("host=%s port=5432 user=postgres dbname=postgres sslmode=disable password=%s", host, password))
	if err != nil {
		log.Fatal("Failed to connect to the database:", err)
	}
	db.SetMaxOpenConns(50)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(time.Hour)
}

func handleConnections(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			log.Println("WebSocket connection handler exiting...")
			return
		default:
		}

		// Connect to WebSocket
		conn, _, err := websocket.DefaultDialer.Dial("wss://ws.okx.com:8443/ws/v5/public", nil)
		if err != nil {
			log.Println("WebSocket connection failed, retrying:", err)
			time.Sleep(5 * time.Second)
			continue
		}

		log.Println("WebSocket connected")
		subscribeMsg := `{
			"op": "subscribe",
			"args": [
				{
					"channel": "tickers",
					"instId": "BTC-USDT"
				}
			]
		}`
		err = conn.WriteMessage(websocket.TextMessage, []byte(subscribeMsg))
		if err != nil {
			log.Println("Failed to send subscription message:", err)
			conn.Close()
			continue
		}

		// Listen for messages
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Println("WebSocket disconnected:", err)
				conn.Close()
				break
			}
			select {
			case dataChannel <- string(msg):
			case <-ctx.Done():
				conn.Close()
				return
			}
		}
	}
}

func writeToDatabase(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			log.Println("Database writer exiting...")
			return
		case msg := <-dataChannel:
			_, err := db.Exec("INSERT INTO oxk_pepe_spot (message) VALUES ($1)", msg)
			if err != nil {
				log.Println("Failed to write to database:", err)
			}
		}
	}
}

func main() {
	initDB()
	defer db.Close()

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())

	// Handle system signals for graceful shutdown
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalChan
		log.Println("Shutting down...")
		cancel()
	}()

	// Start WebSocket handler
	wg.Add(1)
	go handleConnections(ctx, &wg)

	// Start database writers
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go writeToDatabase(ctx, &wg)
	}

	log.Println("Server has started")
	wg.Wait()
	log.Println("Server stopped")
}
