package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"strconv"
	"sync"
	"time"

	"github.com/hyperliquid-go/hyperliquid-go/hyperliquid"
	"github.com/hyperliquid-go/hyperliquid-go/hyperliquid/utils"
)

// Configuration constants
const (
	// How far from the best bid and offer this strategy ideally places orders (0.3%)
	DEPTH = 0.003
	
	// How far from the target price a resting order can deviate before cancellation (50% of depth)
	ALLOWABLE_DEVIATION = 0.5
	
	// Maximum absolute position value the strategy can accumulate
	MAX_POSITION = 1.0
	
	// The coin to add liquidity on
	COIN = "ETH"
	COIN_ASSET_ID = 1 // ETH asset ID
	
	// Polling interval in seconds
	POLL_INTERVAL = 10
	
	// Maximum time to wait for an in-flight order before treating it as cancelled (ms)
	ORDER_TIMEOUT = 10000
	
	// Time to keep recently cancelled orders before cleanup (ms)
	CANCEL_CLEANUP_TIME = 30000
)

// Order states
type ProvideState struct {
	Type string  // "in_flight_order", "resting", "cancelled"
	Time int64   // timestamp for in_flight_order
	Px   float64 // price for resting order
	Oid  int     // order ID for resting order
}

type BasicAdder struct {
	address    string
	info       *hyperliquid.Info
	exchange   *hyperliquid.Exchange
	ws         *hyperliquid.WebSocketManager
	ctx        context.Context
	cancel     context.CancelFunc
	
	// State tracking
	position    *float64
	provideState map[string]*ProvideState // "A" for ask, "B" for bid
	recentlyCancelled map[int]int64
	
	// Synchronization
	mu sync.RWMutex
}

func NewBasicAdder(address string, info *hyperliquid.Info, exchange *hyperliquid.Exchange) *BasicAdder {
	ctx, cancel := context.WithCancel(context.Background())
	
	adder := &BasicAdder{
		address:    address,
		info:       info,
		exchange:   exchange,
		ctx:        ctx,
		cancel:     cancel,
		provideState: map[string]*ProvideState{
			"A": {Type: "cancelled"},
			"B": {Type: "cancelled"},
		},
		recentlyCancelled: make(map[int]int64),
	}
	
	// Initialize WebSocket
	adder.ws = hyperliquid.NewWebSocketManager(utils.TestnetAPIURL)
	
	return adder
}

func (ba *BasicAdder) Start() error {
	// Start WebSocket connection
	if err := ba.ws.Start(); err != nil {
		return fmt.Errorf("failed to start WebSocket: %v", err)
	}
	
	// Subscribe to order book updates
	ba.ws.Subscribe(hyperliquid.Subscription{
		Type: hyperliquid.L2Book,
		Coin: COIN,
	}, ba.handleL2BookMessage)
	
	// Subscribe to user events
	ba.ws.Subscribe(hyperliquid.Subscription{
		Type: hyperliquid.UserEvents,
		User: ba.address,
	}, ba.handleUserEventsMessage)
	
	// No need for separate message processing since callbacks handle messages directly
	
	// Start polling thread
	go ba.poll()
	
	return nil
}

func (ba *BasicAdder) Stop() {
	ba.cancel()
	if ba.ws != nil {
		ba.ws.Stop()
	}
}

// handleL2BookMessage handles order book updates
func (ba *BasicAdder) handleL2BookMessage(msg hyperliquid.WsMsg) {
	log.Printf("Received L2 book update: %+v", msg)
	ba.onBookUpdate(msg)
}

// handleUserEventsMessage handles user events
func (ba *BasicAdder) handleUserEventsMessage(msg hyperliquid.WsMsg) {
	log.Printf("Received user event: %+v", msg)
	ba.onUserEvents(msg)
}

func (ba *BasicAdder) handleMessage(msg hyperliquid.WsMsg) {
	switch msg.Channel {
	case "l2Book":
		ba.onBookUpdate(msg)
	case "userEvents":
		ba.onUserEvents(msg)
	}
}

func (ba *BasicAdder) onBookUpdate(msg hyperliquid.WsMsg) {
	// Parse L2 book data
	bookData, ok := msg.Data.(map[string]interface{})
	if !ok {
		log.Printf("Invalid book data format")
		return
	}
	
	coin, ok := bookData["coin"].(string)
	if !ok || coin != COIN {
		return
	}
	
	levels, ok := bookData["levels"].([]interface{})
	if !ok || len(levels) < 2 {
		return
	}
	
	// Handle both sides
	ba.handleOrderPlacement("B", levels) // Bid
	ba.handleOrderPlacement("A", levels) // Ask
}

func (ba *BasicAdder) handleOrderPlacement(side string, levels []interface{}) {
	ba.mu.Lock()
	defer ba.mu.Unlock()
	
	// Get the appropriate level (0 for bid, 1 for ask)
	levelIndex := 0
	if side == "A" {
		levelIndex = 1
	}
	
	if len(levels) <= levelIndex {
		return
	}
	
	level, ok := levels[levelIndex].([]interface{})
	if !ok || len(level) < 2 {
		return
	}
	
	pxStr, ok := level[0].(string)
	if !ok {
		return
	}
	
	bookPrice := parseFloat(pxStr)
	if bookPrice <= 0 {
		return
	}
	
	idealDistance := bookPrice * DEPTH
	sideMultiplier := 1.0
	if side == "B" {
		sideMultiplier = -1.0
	}
	idealPrice := bookPrice + (idealDistance * sideMultiplier)
	
	provideState := ba.provideState[side]
	
	switch provideState.Type {
	case "resting":
		ba.maybeCancelOrder(side, provideState, idealPrice, idealDistance)
	case "in_flight_order":
		ba.checkInFlightOrder(side, provideState)
	}
	
	if provideState.Type == "cancelled" {
		ba.placeNewOrder(side, idealPrice)
	}
}

func (ba *BasicAdder) maybeCancelOrder(side string, provideState *ProvideState, idealPrice, idealDistance float64) {
	distance := math.Abs(idealPrice - provideState.Px)
	if distance > ALLOWABLE_DEVIATION*idealDistance {
		fmt.Printf("Cancelling order due to deviation: oid:%d, side:%s, ideal_price:%.2f\n",
			provideState.Oid, side, idealPrice)

		cancelResult, err := ba.exchange.Cancel(COIN, provideState.Oid)
		if err != nil {
			log.Printf("Failed to cancel order %d for side %s: %v", provideState.Oid, side, err)
		} else if result, ok := cancelResult.(map[string]interface{}); ok && result["status"] == "ok" {
			ba.recentlyCancelled[provideState.Oid] = time.Now().UnixMilli()
			ba.provideState[side] = &ProvideState{Type: "cancelled"}
		}
	}
}

func (ba *BasicAdder) checkInFlightOrder(side string, provideState *ProvideState) {
	if time.Now().UnixMilli()-provideState.Time > ORDER_TIMEOUT {
		fmt.Println("Order is still in flight after timeout, treating as cancelled.")
		ba.provideState[side] = &ProvideState{Type: "cancelled"}
	}
}

func (ba *BasicAdder) placeNewOrder(side string, idealPrice float64) {
	// Check position limits
	if ba.position != nil && math.Abs(*ba.position) >= MAX_POSITION {
		return
	}

	isBuy := side == "B"

	orderType := utils.OrderType{
		Limit: &utils.LimitOrderType{},
	}

	fmt.Printf("Placing %s order at %.2f\n", side, idealPrice)

	size := 0.1
	price := idealPrice
	orderResult, err := ba.exchange.Order(COIN, isBuy, size, price, orderType, false, nil, nil)
	if err != nil {
		log.Printf("Failed to place order for side %s: %v", side, err)
		return
	}

	if result, ok := orderResult.(map[string]interface{}); ok && result["status"] == "ok" {
		// Order placed successfully
		log.Printf("Order placed successfully for side %s", side)
	} else {
		ba.provideState[side] = &ProvideState{
			Type: "in_flight_order",
			Time: time.Now().UnixMilli(),
		}
	}
}

func (ba *BasicAdder) onUserEvents(msg hyperliquid.WsMsg) {
	// Handle user events (fills, etc.)
	fmt.Printf("User event: %+v\n", msg.Data)
}

func (ba *BasicAdder) poll() {
	ticker := time.NewTicker(POLL_INTERVAL * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ba.ctx.Done():
			return
		case <-ticker.C:
			ba.updatePosition()
			ba.cleanupCancelledOrders()
		}
	}
}

func (ba *BasicAdder) updatePosition() {
	userState, err := ba.info.UserState(ba.address, "")
	if err != nil {
		log.Printf("Failed to get user state: %v", err)
		return
	}
	
	ba.mu.Lock()
	defer ba.mu.Unlock()
	
	// Parse user state response
	if stateMap, ok := userState.(map[string]interface{}); ok {
		if assetPositions, ok := stateMap["assetPositions"].([]interface{}); ok {
			for _, pos := range assetPositions {
				if posMap, ok := pos.(map[string]interface{}); ok {
					if position, ok := posMap["position"].(map[string]interface{}); ok {
						if coin, ok := position["coin"].(string); ok && coin == COIN {
							if sizeStr, ok := position["szi"].(string); ok {
								if size, err := strconv.ParseFloat(sizeStr, 64); err == nil {
									ba.position = &size
									return
								}
							}
						}
					}
				}
			}
		}
	}
}

func (ba *BasicAdder) cleanupCancelledOrders() {
	ba.mu.Lock()
	defer ba.mu.Unlock()
	
	now := time.Now().UnixMilli()
	for oid, cancelTime := range ba.recentlyCancelled {
		if now-cancelTime > CANCEL_CLEANUP_TIME {
			delete(ba.recentlyCancelled, oid)
		}
	}
}

func parseFloat(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}

func RunBasicAdding() {
	// Setup clients
	address, info, exchange, err := Setup(utils.TestnetAPIURL, false)
	if err != nil {
		log.Fatal("Setup failed:", err)
	}
	
	// Create and start the basic adder
	adder := NewBasicAdder(address, info, exchange)
	
	err = adder.Start()
	if err != nil {
		log.Fatal("Failed to start adder:", err)
	}
	
	fmt.Printf("Basic adding strategy started for %s on %s\n", COIN, address)
	fmt.Println("Press Ctrl+C to stop...")
	
	// Keep running
	select {}
}
