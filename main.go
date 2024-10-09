package main

import (
	"database/sql"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"os"
	//"net/http"
	_ "github.com/lib/pq"
	"sync"
	"time"
)

// Upgrade HTTP connection to WebSocket connection
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Database connection pool
var db *sql.DB

// Data write channel
var dataChannel = make(chan string, 1000) // Buffered channel to reduce write blocking

// Initialize database connection
func initDB() {
	var err error
	// Get the database IP address from the environment variable DB_HOST, default to 127.0.0.1 if not set
	host := os.Getenv("DB_HOST")

	// Establish a connection to the PostgreSQL database using the specified host IP
	db, err = sql.Open("postgres", fmt.Sprintf("host=%s port=5432 user=postgres dbname=oxk_data sslmode=disable password=123456789", host))
	if err != nil {
		log.Fatal("Failed to connect to the database:", err)
	}
	// Set maximum connection numbers
	db.SetMaxOpenConns(50)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(time.Hour)
}

// WebSocket connection handling
func handleConnections() {
	// Connect to OKX WebSocket server
	conn, _, err := websocket.DefaultDialer.Dial("wss://ws.okx.com:8443/ws/v5/public", nil)
	if err != nil {
		log.Println("Failed to connect to WebSocket:", err)
		return
	}
	defer conn.Close()

	// Build subscribe message
	subscribeMsg := `{
        "op": "subscribe",
        "args": [
            {
                "channel": "tickers",
                "instId": "BTC-USDT"
            }
        ]
    }`

	// Send subscribe message
	err = conn.WriteMessage(websocket.TextMessage, []byte(subscribeMsg))
	if err != nil {
		log.Fatalf("Failed to send subscription message: %v", err)
		return
	}

	// Continuously receive WebSocket messages
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading WebSocket message:", err)
			break
		}

		// Send message to channel
		dataChannel <- string(msg)
	}
}

// Write data to the database
func writeToDatabase(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case msg := <-dataChannel:
			// Batch write to the database, or single write (choose based on actual situation)
			_, err := db.Exec("INSERT INTO oxk_pepe_spot (message) VALUES ($1)", msg)
			if err != nil {
				log.Println("Failed to insert data into database:", err)
			}
		}
	}
}

func main() {
	// Initialize database
	initDB()
	defer db.Close()

	// Create wait group
	var wg sync.WaitGroup

	// Start WebSocket connection handling
	go handleConnections()

	// Start multiple goroutines for database writing
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go writeToDatabase(&wg)
	}

	// Start HTTP server (if needed)
	log.Println("Server has started")
	select {} // Keep main goroutine running

	// Wait for all database goroutines to complete
	wg.Wait()
}
