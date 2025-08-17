// Package hyperliquid - Exchange functionality
package hyperliquid

import (
	"crypto/ecdsa"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/hyperliquid-go/hyperliquid-go/hyperliquid/utils"
)

// Default max slippage for market orders (5%)
const DefaultSlippage = 0.05

// BuilderInfo represents builder information for orders
type BuilderInfo struct {
	B string `json:"b"`
	F string `json:"f"`
}

// Exchange represents the Exchange API client for trading operations
type Exchange struct {
	*API
	privateKey    *ecdsa.PrivateKey
	vaultAddress  *string
	accountAddress *string
	info          *Info
	expiresAfter  *int64
}

// NewExchange creates a new Exchange client instance
func NewExchange(privateKey *ecdsa.PrivateKey, baseURL string, meta *Meta, vaultAddress *string, accountAddress *string, spotMeta *SpotMeta, perpDexs []string, timeout time.Duration) (*Exchange, error) {
	if baseURL == "" {
		baseURL = utils.MainnetAPIURL
	}
	
	api := NewAPI(baseURL, timeout)
	info, err := NewInfo(baseURL, true, meta, spotMeta, perpDexs, timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to create info client: %w", err)
	}
	
	return &Exchange{
		API:            api,
		privateKey:    privateKey,
		vaultAddress:  vaultAddress,
		accountAddress: accountAddress,
		info:          info,
	}, nil
}

// postAction sends a signed action to the exchange
func (e *Exchange) postAction(action map[string]interface{}, signature string, nonce int64) (interface{}, error) {
	payload := map[string]interface{}{
		"action":    action,
		"nonce":     nonce,
		"signature": signature,
	}
	
	// Add vault address for certain action types
	actionType, _ := action["type"].(string)
	if actionType != "usdClassTransfer" && actionType != "sendAsset" {
		if e.vaultAddress != nil {
			payload["vaultAddress"] = *e.vaultAddress
		}
	}
	
	if e.expiresAfter != nil {
		payload["expiresAfter"] = *e.expiresAfter
	}
	
	return e.Post("/exchange", payload)
}

// slippagePrice calculates price with slippage for market orders
func (e *Exchange) slippagePrice(name string, isBuy bool, slippage float64, px *float64) (float64, error) {
	coin, exists := e.info.nameToCoins[name]
	if !exists {
		return 0, fmt.Errorf("coin not found for name: %s", name)
	}
	
	var price float64
	if px != nil {
		price = *px
	} else {
		// Get midprice
		allMids, err := e.info.AllMids("")
		if err != nil {
			return 0, fmt.Errorf("failed to get all mids: %w", err)
		}
		
		if midsMap, ok := allMids.(map[string]interface{}); ok {
			if midStr, ok := midsMap[coin].(string); ok {
				var err error
				price, err = strconv.ParseFloat(midStr, 64)
				if err != nil {
					return 0, fmt.Errorf("failed to parse mid price: %w", err)
				}
			} else {
				return 0, fmt.Errorf("mid price not found for coin: %s", coin)
			}
		} else {
			return 0, fmt.Errorf("invalid all mids response format")
		}
	}
	
	asset, exists := e.info.coinToAsset[coin]
	if !exists {
		return 0, fmt.Errorf("asset not found for coin: %s", coin)
	}
	
	// Spot assets start at 10000
	isSpot := asset >= 10000
	
	// Calculate slippage
	if isBuy {
		price *= (1 + slippage)
	} else {
		price *= (1 - slippage)
	}
	
	// Round to appropriate decimals
	szDecimals := e.info.assetToSzDecimals[asset]
	var decimals int
	if isSpot {
		decimals = 8 - szDecimals
	} else {
		decimals = 6 - szDecimals
	}
	
	// Round to 5 significant figures and appropriate decimals
	multiplier := math.Pow(10, float64(decimals))
	return math.Round(price*multiplier) / multiplier, nil
}

// SetExpiresAfter sets the expiration time for actions
func (e *Exchange) SetExpiresAfter(expiresAfter *int64) {
	e.expiresAfter = expiresAfter
}

// Order places a single order
func (e *Exchange) Order(name string, isBuy bool, sz float64, limitPx float64, orderType utils.OrderType, reduceOnly bool, cloid *string, builder *BuilderInfo) (interface{}, error) {
	orderRequest := utils.OrderRequest{
		Coin:       name,
		IsBuy:      isBuy,
		Sz:         sz,
		LimitPx:    limitPx,
		OrderType:  orderType,
		ReduceOnly: reduceOnly,
		Cloid:      cloid,
	}
	
	return e.BulkOrders([]utils.OrderRequest{orderRequest}, builder)
}

// BulkOrders places multiple orders in a single transaction
func (e *Exchange) BulkOrders(orderRequests []utils.OrderRequest, builder *BuilderInfo) (interface{}, error) {
	orderWires := make([]utils.OrderWire, len(orderRequests))
	
	for i, order := range orderRequests {
		asset, err := e.info.NameToAsset(order.Coin)
		if err != nil {
			return nil, fmt.Errorf("failed to get asset for coin %s: %w", order.Coin, err)
		}
		
		orderWire, err := utils.OrderRequestToOrderWire(order, asset)
		if err != nil {
			return nil, fmt.Errorf("failed to convert order to wire format: %w", err)
		}
		orderWires[i] = *orderWire
	}
	
	timestamp := utils.GetTimestampMs()
	
	var builderStr *string
	if builder != nil {
		builderStr = &builder.B
	}
	orderAction := utils.OrderWiresToOrderAction(orderWires, builderStr)
	
	isMainnet := e.GetBaseURL() == utils.MainnetAPIURL
	
	var expiresAfterUint *uint64
	if e.expiresAfter != nil {
		uint64Val := uint64(*e.expiresAfter)
		expiresAfterUint = &uint64Val
	}
	
	signature, err := utils.SignL1Action(e.privateKey, orderAction, e.vaultAddress, uint64(timestamp), expiresAfterUint, isMainnet)
	if err != nil {
		return nil, fmt.Errorf("failed to sign order action: %w", err)
	}
	
	return e.postAction(orderAction, signature.R+signature.S+fmt.Sprintf("%02x", signature.V), timestamp)
}

// MarketOpen places a market order to open a position
func (e *Exchange) MarketOpen(name string, isBuy bool, sz float64, px *float64, slippage float64, cloid *string, builder *BuilderInfo) (interface{}, error) {
	if slippage == 0 {
		slippage = DefaultSlippage
	}
	
	// Get aggressive market price
	price, err := e.slippagePrice(name, isBuy, slippage, px)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate slippage price: %w", err)
	}
	
	// Market order is an aggressive limit order IoC
	orderType := utils.OrderType{
		Limit: &utils.LimitOrderType{
			TIF: utils.TIFIoc,
		},
	}
	
	return e.Order(name, isBuy, sz, price, orderType, false, cloid, builder)
}

// MarketClose places a market order to close a position
func (e *Exchange) MarketClose(coin string, sz *float64, px *float64, slippage float64, cloid *string, builder *BuilderInfo) (interface{}, error) {
	if slippage == 0 {
		slippage = DefaultSlippage
	}
	
	address := crypto.PubkeyToAddress(e.privateKey.PublicKey).Hex()
	if e.accountAddress != nil {
		address = *e.accountAddress
	}
	if e.vaultAddress != nil {
		address = *e.vaultAddress
	}
	
	userState, err := e.info.UserState(address, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get user state: %w", err)
	}
	
	if userStateMap, ok := userState.(map[string]interface{}); ok {
		if assetPositions, ok := userStateMap["assetPositions"].([]interface{}); ok {
			for _, positionInterface := range assetPositions {
				if positionMap, ok := positionInterface.(map[string]interface{}); ok {
					if position, ok := positionMap["position"].(map[string]interface{}); ok {
						if positionCoin, ok := position["coin"].(string); ok && positionCoin == coin {
							if sziStr, ok := position["szi"].(string); ok {
								szi, err := strconv.ParseFloat(sziStr, 64)
								if err != nil {
									return nil, fmt.Errorf("failed to parse szi: %w", err)
								}
								
								size := sz
								if size == nil {
									absSize := math.Abs(szi)
									size = &absSize
								}
								
								isBuy := szi < 0
								
								// Get aggressive market price
								price, err := e.slippagePrice(coin, isBuy, slippage, px)
								if err != nil {
									return nil, fmt.Errorf("failed to calculate slippage price: %w", err)
								}
								
								// Market order is an aggressive limit order IoC
								orderType := utils.OrderType{
									Limit: &utils.LimitOrderType{
										TIF: utils.TIFIoc,
									},
								}
								
								return e.Order(coin, isBuy, *size, price, orderType, true, cloid, builder)
							}
						}
					}
				}
			}
		}
	}
	
	return nil, fmt.Errorf("position not found for coin: %s", coin)
}

// Cancel cancels a single order
func (e *Exchange) Cancel(name string, oid int) (interface{}, error) {
	cancelRequest := utils.CancelRequest{
		Coin: name,
		OID:  oid,
	}
	return e.BulkCancel([]utils.CancelRequest{cancelRequest})
}

// BulkCancel cancels multiple orders
func (e *Exchange) BulkCancel(cancelRequests []utils.CancelRequest) (interface{}, error) {
	timestamp := utils.GetTimestampMs()
	cancels := make([]map[string]interface{}, len(cancelRequests))
	
	for i, cancel := range cancelRequests {
		asset, err := e.info.NameToAsset(cancel.Coin)
		if err != nil {
			return nil, fmt.Errorf("failed to get asset for coin %s: %w", cancel.Coin, err)
		}
		
		cancels[i] = map[string]interface{}{
			"a": asset,
			"o": cancel.OID,
		}
	}
	
	cancelAction := map[string]interface{}{
		"type":    "cancel",
		"cancels": cancels,
	}
	
	isMainnet := e.GetBaseURL() == utils.MainnetAPIURL
	
	var expiresAfterUint *uint64
	if e.expiresAfter != nil {
		uint64Val := uint64(*e.expiresAfter)
		expiresAfterUint = &uint64Val
	}
	
	signature, err := utils.SignL1Action(e.privateKey, cancelAction, e.vaultAddress, uint64(timestamp), expiresAfterUint, isMainnet)
	if err != nil {
		return nil, fmt.Errorf("failed to sign cancel action: %w", err)
	}
	
	return e.postAction(cancelAction, signature.R+signature.S+fmt.Sprintf("%02x", signature.V), timestamp)
}

// UpdateLeverage updates leverage for a specific asset
func (e *Exchange) UpdateLeverage(leverage int, name string, isCross bool) (interface{}, error) {
	timestamp := utils.GetTimestampMs()
	asset, err := e.info.NameToAsset(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get asset for name %s: %w", name, err)
	}
	
	updateAction := map[string]interface{}{
		"type":     "updateLeverage",
		"asset":    asset,
		"isCross":  isCross,
		"leverage": leverage,
	}
	
	isMainnet := e.GetBaseURL() == utils.MainnetAPIURL
	
	var expiresAfterUint *uint64
	if e.expiresAfter != nil {
		uint64Val := uint64(*e.expiresAfter)
		expiresAfterUint = &uint64Val
	}
	
	signature, err := utils.SignL1Action(e.privateKey, updateAction, e.vaultAddress, uint64(timestamp), expiresAfterUint, isMainnet)
	if err != nil {
		return nil, fmt.Errorf("failed to sign update leverage action: %w", err)
	}
	
	return e.postAction(updateAction, signature.R+signature.S+fmt.Sprintf("%02x", signature.V), timestamp)
}

// UsdClassTransfer transfers USD between perp and spot
func (e *Exchange) UsdClassTransfer(amount float64, toPerp bool) (interface{}, error) {
	timestamp := utils.GetTimestampMs()
	strAmount := fmt.Sprintf("%.6f", amount)
	
	if e.vaultAddress != nil {
		strAmount += fmt.Sprintf(" subaccount:%s", *e.vaultAddress)
	}
	
	action := map[string]interface{}{
		"type":   "usdClassTransfer",
		"amount": strAmount,
		"toPerp": toPerp,
		"nonce":  timestamp,
	}
	
	isMainnet := e.GetBaseURL() == utils.MainnetAPIURL
	
	signature, err := utils.SignUSDClassTransferAction(e.privateKey, action, isMainnet)
	if err != nil {
		return nil, fmt.Errorf("failed to sign USD class transfer action: %w", err)
	}
	
	return e.postAction(action, signature.R+signature.S+fmt.Sprintf("%02x", signature.V), timestamp)
}

// UsdTransfer transfers USD to another address
func (e *Exchange) UsdTransfer(amount float64, destination string) (interface{}, error) {
	timestamp := utils.GetTimestampMs()
	action := map[string]interface{}{
		"destination": destination,
		"amount":      fmt.Sprintf("%.6f", amount),
		"time":        timestamp,
		"type":        "usdSend",
	}
	
	isMainnet := e.GetBaseURL() == utils.MainnetAPIURL
	
	signature, err := utils.SignUSDTransferAction(e.privateKey, action, isMainnet)
	if err != nil {
		return nil, fmt.Errorf("failed to sign USD transfer action: %w", err)
	}
	
	return e.postAction(action, signature.R+signature.S+fmt.Sprintf("%02x", signature.V), timestamp)
}
