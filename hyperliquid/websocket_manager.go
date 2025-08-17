// Package hyperliquid - WebSocket manager functionality
package hyperliquid

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Subscription types
type SubscriptionType string

const (
	AllMids                        SubscriptionType = "allMids"
	L2Book                         SubscriptionType = "l2Book"
	Trades                         SubscriptionType = "trades"
	UserEvents                     SubscriptionType = "userEvents"
	UserFills                      SubscriptionType = "userFills"
	Candle                         SubscriptionType = "candle"
	OrderUpdates                   SubscriptionType = "orderUpdates"
	UserFundings                   SubscriptionType = "userFundings"
	UserNonFundingLedgerUpdates    SubscriptionType = "userNonFundingLedgerUpdates"
	WebData2                       SubscriptionType = "webData2"
	BBO                            SubscriptionType = "bbo"
	ActiveAssetCtx                 SubscriptionType = "activeAssetCtx"
	ActiveAssetData                SubscriptionType = "activeAssetData"
)

// Subscription represents a WebSocket subscription
type Subscription struct {
	Type     SubscriptionType `json:"type"`
	Coin     string           `json:"coin,omitempty"`
	User     string           `json:"user,omitempty"`
	Interval string           `json:"interval,omitempty"`
}

// WsMsg represents a WebSocket message
type WsMsg struct {
	Channel string      `json:"channel"`
	Data    interface{} `json:"data,omitempty"`
}

// ActiveSubscription represents an active subscription with callback
type ActiveSubscription struct {
	Callback       func(WsMsg)
	SubscriptionID int
}

// WebSocketManager manages WebSocket connections and subscriptions
type WebSocketManager struct {
	mu                      sync.RWMutex
	conn                    *websocket.Conn
	baseURL                 string
	subscriptionIDCounter   int
	wsReady                 bool
	queuedSubscriptions     []queuedSubscription
	activeSubscriptions     map[string][]ActiveSubscription
	ctx                     context.Context
	cancel                  context.CancelFunc
	stopCh                  chan struct{}
	pingTicker              *time.Ticker
}

type queuedSubscription struct {
	subscription Subscription
	active       ActiveSubscription
}

// NewWebSocketManager creates a new WebSocket manager
func NewWebSocketManager(baseURL string) *WebSocketManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &WebSocketManager{
		baseURL:             baseURL,
		activeSubscriptions: make(map[string][]ActiveSubscription),
		ctx:                 ctx,
		cancel:              cancel,
		stopCh:              make(chan struct{}),
	}
}

// Start starts the WebSocket connection and message handling
func (w *WebSocketManager) Start() error {
	wsURL := "ws" + w.baseURL[len("http"):] + "/ws"
	
	dialer := websocket.Dialer{
		HandshakeTimeout: 45 * time.Second,
	}
	
	conn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to WebSocket: %w", err)
	}
	
	w.mu.Lock()
	w.conn = conn
	w.mu.Unlock()
	
	// Start ping sender
	w.pingTicker = time.NewTicker(50 * time.Second)
	go w.sendPing()
	
	// Start message handler
	go w.handleMessages()
	
	return nil
}

// Stop stops the WebSocket connection and all goroutines
func (w *WebSocketManager) Stop() {
	w.cancel()
	close(w.stopCh)
	
	if w.pingTicker != nil {
		w.pingTicker.Stop()
	}
	
	w.mu.Lock()
	if w.conn != nil {
		w.conn.Close()
	}
	w.mu.Unlock()
}

// sendPing sends periodic ping messages
func (w *WebSocketManager) sendPing() {
	for {
		select {
		case <-w.ctx.Done():
			log.Println("WebSocket ping sender stopped")
			return
		case <-w.pingTicker.C:
			w.mu.RLock()
			conn := w.conn
			w.mu.RUnlock()
			
			if conn != nil {
				log.Println("WebSocket sending ping")
				pingMsg := map[string]string{"method": "ping"}
				if err := conn.WriteJSON(pingMsg); err != nil {
					log.Printf("Failed to send ping: %v", err)
				}
			}
		}
	}
}

// handleMessages handles incoming WebSocket messages
func (w *WebSocketManager) handleMessages() {
	for {
		select {
		case <-w.ctx.Done():
			return
		default:
			w.mu.RLock()
			conn := w.conn
			w.mu.RUnlock()
			
			if conn == nil {
				continue
			}
			
			var message json.RawMessage
			err := conn.ReadJSON(&message)
			if err != nil {
				log.Printf("WebSocket read error: %v", err)
				return
			}
			
			// Handle string messages
			var strMsg string
			if err := json.Unmarshal(message, &strMsg); err == nil {
				if strMsg == "Websocket connection established." {
					log.Println(strMsg)
					w.onOpen()
					continue
				}
			}
			
			// Handle JSON messages
			var wsMsg WsMsg
			if err := json.Unmarshal(message, &wsMsg); err != nil {
				log.Printf("Failed to unmarshal WebSocket message: %v", err)
				continue
			}
			
			w.onMessage(wsMsg)
		}
	}
}

// onOpen handles WebSocket connection open event
func (w *WebSocketManager) onOpen() {
	log.Println("WebSocket connection opened")
	w.mu.Lock()
	w.wsReady = true
	
	// Process queued subscriptions
	for _, queued := range w.queuedSubscriptions {
		w.subscribeInternal(queued.subscription, queued.active.Callback, queued.active.SubscriptionID)
	}
	w.queuedSubscriptions = nil
	w.mu.Unlock()
}

// onMessage handles incoming WebSocket messages
func (w *WebSocketManager) onMessage(wsMsg WsMsg) {
	log.Printf("Received message: %+v", wsMsg)
	
	identifier := w.wsMsgToIdentifier(wsMsg)
	if identifier == "pong" {
		log.Println("WebSocket received pong")
		return
	}
	
	if identifier == "" {
		log.Println("WebSocket not handling empty message")
		return
	}
	
	w.mu.RLock()
	activeSubscriptions := w.activeSubscriptions[identifier]
	w.mu.RUnlock()
	
	if len(activeSubscriptions) == 0 {
		log.Printf("WebSocket message from unexpected subscription: %s, identifier: %s", wsMsg.Channel, identifier)
	} else {
		for _, activeSub := range activeSubscriptions {
			activeSub.Callback(wsMsg)
		}
	}
}

// Subscribe subscribes to a WebSocket channel
func (w *WebSocketManager) Subscribe(subscription Subscription, callback func(WsMsg)) int {
	w.mu.Lock()
	defer w.mu.Unlock()
	
	w.subscriptionIDCounter++
	subscriptionID := w.subscriptionIDCounter
	
	if !w.wsReady {
		log.Println("Enqueueing subscription")
		w.queuedSubscriptions = append(w.queuedSubscriptions, queuedSubscription{
			subscription: subscription,
			active:       ActiveSubscription{Callback: callback, SubscriptionID: subscriptionID},
		})
	} else {
		w.subscribeInternal(subscription, callback, subscriptionID)
	}
	
	return subscriptionID
}

// subscribeInternal handles the actual subscription logic
func (w *WebSocketManager) subscribeInternal(subscription Subscription, callback func(WsMsg), subscriptionID int) {
	log.Println("Subscribing")
	identifier := w.subscriptionToIdentifier(subscription)
	
	// Check for single subscription constraints
	if identifier == "userEvents" || identifier == "orderUpdates" {
		if len(w.activeSubscriptions[identifier]) != 0 {
			log.Printf("Cannot subscribe to %s multiple times", identifier)
			return
		}
	}
	
	w.activeSubscriptions[identifier] = append(w.activeSubscriptions[identifier], ActiveSubscription{
		Callback:       callback,
		SubscriptionID: subscriptionID,
	})
	
	subMsg := map[string]interface{}{
		"method":      "subscribe",
		"subscription": subscription,
	}
	
	if w.conn != nil {
		if err := w.conn.WriteJSON(subMsg); err != nil {
			log.Printf("Failed to send subscription: %v", err)
		}
	}
}

// Unsubscribe unsubscribes from a WebSocket channel
func (w *WebSocketManager) Unsubscribe(subscription Subscription, subscriptionID int) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	
	if !w.wsReady {
		log.Println("Cannot unsubscribe before WebSocket connected")
		return false
	}
	
	identifier := w.subscriptionToIdentifier(subscription)
	activeSubscriptions := w.activeSubscriptions[identifier]
	
	newActiveSubscriptions := make([]ActiveSubscription, 0)
	for _, sub := range activeSubscriptions {
		if sub.SubscriptionID != subscriptionID {
			newActiveSubscriptions = append(newActiveSubscriptions, sub)
		}
	}
	
	if len(newActiveSubscriptions) == 0 {
		unsubMsg := map[string]interface{}{
			"method":      "unsubscribe",
			"subscription": subscription,
		}
		if w.conn != nil {
			if err := w.conn.WriteJSON(unsubMsg); err != nil {
				log.Printf("Failed to send unsubscription: %v", err)
			}
		}
	}
	
	w.activeSubscriptions[identifier] = newActiveSubscriptions
	return len(activeSubscriptions) != len(newActiveSubscriptions)
}

// subscriptionToIdentifier converts a subscription to an identifier string
func (w *WebSocketManager) subscriptionToIdentifier(subscription Subscription) string {
	switch subscription.Type {
	case AllMids:
		return "allMids"
	case L2Book:
		return fmt.Sprintf("l2Book:%s", strings.ToLower(subscription.Coin))
	case Trades:
		return fmt.Sprintf("trades:%s", strings.ToLower(subscription.Coin))
	case UserEvents:
		return "userEvents"
	case UserFills:
		return fmt.Sprintf("userFills:%s", strings.ToLower(subscription.User))
	case Candle:
		return fmt.Sprintf("candle:%s,%s", strings.ToLower(subscription.Coin), subscription.Interval)
	case OrderUpdates:
		return "orderUpdates"
	case UserFundings:
		return fmt.Sprintf("userFundings:%s", strings.ToLower(subscription.User))
	case UserNonFundingLedgerUpdates:
		return fmt.Sprintf("userNonFundingLedgerUpdates:%s", strings.ToLower(subscription.User))
	case WebData2:
		return fmt.Sprintf("webData2:%s", strings.ToLower(subscription.User))
	case BBO:
		return fmt.Sprintf("bbo:%s", strings.ToLower(subscription.Coin))
	case ActiveAssetCtx:
		return fmt.Sprintf("activeAssetCtx:%s", strings.ToLower(subscription.Coin))
	case ActiveAssetData:
		return fmt.Sprintf("activeAssetData:%s,%s", strings.ToLower(subscription.Coin), strings.ToLower(subscription.User))
	default:
		return ""
	}
}

// wsMsgToIdentifier converts a WebSocket message to an identifier string
func (w *WebSocketManager) wsMsgToIdentifier(wsMsg WsMsg) string {
	switch wsMsg.Channel {
	case "pong":
		return "pong"
	case "allMids":
		return "allMids"
	case "l2Book":
		if data, ok := wsMsg.Data.(map[string]interface{}); ok {
			if coin, ok := data["coin"].(string); ok {
				return fmt.Sprintf("l2Book:%s", strings.ToLower(coin))
			}
		}
	case "trades":
		if trades, ok := wsMsg.Data.([]interface{}); ok && len(trades) > 0 {
			if trade, ok := trades[0].(map[string]interface{}); ok {
				if coin, ok := trade["coin"].(string); ok {
					return fmt.Sprintf("trades:%s", strings.ToLower(coin))
				}
			}
		}
	case "user":
		return "userEvents"
	case "userFills":
		if data, ok := wsMsg.Data.(map[string]interface{}); ok {
			if user, ok := data["user"].(string); ok {
				return fmt.Sprintf("userFills:%s", strings.ToLower(user))
			}
		}
	case "candle":
		if data, ok := wsMsg.Data.(map[string]interface{}); ok {
			if s, ok := data["s"].(string); ok {
				if i, ok := data["i"].(string); ok {
					return fmt.Sprintf("candle:%s,%s", strings.ToLower(s), i)
				}
			}
		}
	case "orderUpdates":
		return "orderUpdates"
	case "userFundings":
		if data, ok := wsMsg.Data.(map[string]interface{}); ok {
			if user, ok := data["user"].(string); ok {
				return fmt.Sprintf("userFundings:%s", strings.ToLower(user))
			}
		}
	case "userNonFundingLedgerUpdates":
		if data, ok := wsMsg.Data.(map[string]interface{}); ok {
			if user, ok := data["user"].(string); ok {
				return fmt.Sprintf("userNonFundingLedgerUpdates:%s", strings.ToLower(user))
			}
		}
	case "webData2":
		if data, ok := wsMsg.Data.(map[string]interface{}); ok {
			if user, ok := data["user"].(string); ok {
				return fmt.Sprintf("webData2:%s", strings.ToLower(user))
			}
		}
	case "bbo":
		if data, ok := wsMsg.Data.(map[string]interface{}); ok {
			if coin, ok := data["coin"].(string); ok {
				return fmt.Sprintf("bbo:%s", strings.ToLower(coin))
			}
		}
	case "activeAssetCtx", "activeSpotAssetCtx":
		if data, ok := wsMsg.Data.(map[string]interface{}); ok {
			if coin, ok := data["coin"].(string); ok {
				return fmt.Sprintf("activeAssetCtx:%s", strings.ToLower(coin))
			}
		}
	case "activeAssetData":
		if data, ok := wsMsg.Data.(map[string]interface{}); ok {
			if coin, ok := data["coin"].(string); ok {
				if user, ok := data["user"].(string); ok {
					return fmt.Sprintf("activeAssetData:%s,%s", strings.ToLower(coin), strings.ToLower(user))
				}
			}
		}
	}
	return ""
}
