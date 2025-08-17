package main

import (
	"fmt"
	"log"
	"time"

	"github.com/hyperliquid-go/hyperliquid-go/hyperliquid"
	"github.com/hyperliquid-go/hyperliquid-go/hyperliquid/utils"
)

func RunBasicWS() {
	// Setup clients
	address, _, _, err := Setup(utils.TestnetAPIURL, false) // Don't skip WebSocket
	if err != nil {
		log.Fatal("Setup failed:", err)
	}

	// Initialize WebSocket manager
	ws := hyperliquid.NewWebSocketManager(true) // testnet = true
	defer ws.Close()

	// Message handler function
	messageHandler := func(msg hyperliquid.WebSocketMessage) {
		fmt.Printf("Received message on channel %s: %+v\n", msg.Channel, msg.Data)
	}

	// Subscribe to different subscription types
	subscriptions := []map[string]interface{}{
		{"type": "allMids"},
		{"type": "l2Book", "coin": "ETH"},
		{"type": "trades", "coin": "PURR/USDC"},
		{"type": "userEvents", "user": address},
		{"type": "userFills", "user": address},
		{"type": "candle", "coin": "ETH", "interval": "1m"},
		{"type": "orderUpdates", "user": address},
		{"type": "userFundings", "user": address},
		{"type": "userNonFundingLedgerUpdates", "user": address},
		{"type": "webData2", "user": address},
		{"type": "bbo", "coin": "ETH"},
		{"type": "activeAssetCtx", "coin": "BTC"}, // Perp
		{"type": "activeAssetCtx", "coin": "@1"},  // Spot
		{"type": "activeAssetData", "user": address, "coin": "BTC"}, // Perp only
	}

	// Subscribe to all channels
	for _, subscription := range subscriptions {
		err := ws.Subscribe(subscription)
		if err != nil {
			log.Printf("Failed to subscribe to %+v: %v", subscription, err)
		} else {
			fmt.Printf("Subscribed to: %+v\n", subscription)
		}
	}

	// Start message processing
	go func() {
		for msg := range ws.Messages() {
			messageHandler(msg)
		}
	}()

	// Keep the program running to receive messages
	fmt.Println("WebSocket subscriptions active. Press Ctrl+C to exit.")
	
	// Run for 30 seconds to see some messages
	time.Sleep(30 * time.Second)
	fmt.Println("Shutting down...")
}
