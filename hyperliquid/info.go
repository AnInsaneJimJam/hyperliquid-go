// Package hyperliquid - Info functionality
package hyperliquid

import (
	"fmt"
	"time"

	"github.com/hyperliquid-go/hyperliquid-go/hyperliquid/utils"
)

// Meta represents exchange metadata
type Meta struct {
	Universe []AssetInfo `json:"universe"`
}

// AssetInfo represents asset information
type AssetInfo struct {
	Name       string `json:"name"`
	SzDecimals int    `json:"szDecimals"`
}

// SpotMeta represents spot exchange metadata
type SpotMeta struct {
	Universe []SpotAssetInfo `json:"universe"`
	Tokens   []SpotTokenInfo `json:"tokens"`
}

// SpotAssetInfo represents spot asset information
type SpotAssetInfo struct {
	Name        string `json:"name"`
	Tokens      [2]int `json:"tokens"`
	Index       int    `json:"index"`
	IsCanonical bool   `json:"isCanonical"`
}

// SpotTokenInfo represents spot token information
type SpotTokenInfo struct {
	Name         string  `json:"name"`
	SzDecimals   int     `json:"szDecimals"`
	WeiDecimals  int     `json:"weiDecimals"`
	Index        int     `json:"index"`
	TokenID      string  `json:"tokenId"`
	IsCanonical  bool    `json:"isCanonical"`
	EvmContract  *string `json:"evmContract,omitempty"`
	FullName     *string `json:"fullName,omitempty"`
}

// SpotMetaAndAssetCtxs represents spot metadata and asset contexts
type SpotMetaAndAssetCtxs struct {
	Meta      SpotMeta        `json:"meta"`
	AssetCtxs []SpotAssetCtx  `json:"assetCtxs"`
}

// SpotAssetCtx represents spot asset context
type SpotAssetCtx struct {
	DayNtlVlm         string  `json:"dayNtlVlm"`
	MarkPx            string  `json:"markPx"`
	MidPx             *string `json:"midPx,omitempty"`
	PrevDayPx         string  `json:"prevDayPx"`
	CirculatingSupply string  `json:"circulatingSupply"`
	Coin              string  `json:"coin"`
}

// Info represents the Info API client
type Info struct {
	*API
	wsManager           *WebSocketManager
	coinToAsset         map[string]int
	nameToCoins         map[string]string
	assetToSzDecimals   map[int]int
}

// NewInfo creates a new Info client instance
func NewInfo(baseURL string, skipWS bool, meta *Meta, spotMeta *SpotMeta, perpDexs []string, timeout time.Duration) (*Info, error) {
	if baseURL == "" {
		baseURL = utils.MainnetAPIURL
	}
	
	api := NewAPI(baseURL, timeout)
	info := &Info{
		API:               api,
		coinToAsset:       make(map[string]int),
		nameToCoins:       make(map[string]string),
		assetToSzDecimals: make(map[int]int),
	}
	
	// Initialize WebSocket manager if not skipped
	if !skipWS {
		info.wsManager = NewWebSocketManager(baseURL)
		if err := info.wsManager.Start(); err != nil {
			return nil, fmt.Errorf("failed to start WebSocket manager: %w", err)
		}
	}
	
	// Initialize spot metadata
	if spotMeta == nil {
		var err error
		spotMeta, err = info.SpotMeta()
		if err != nil {
			return nil, fmt.Errorf("failed to get spot metadata: %w", err)
		}
	}
	
	// Process spot assets (start at 10000)
	for _, spotInfo := range spotMeta.Universe {
		asset := spotInfo.Index + 10000
		info.coinToAsset[spotInfo.Name] = asset
		info.nameToCoins[spotInfo.Name] = spotInfo.Name
		
		baseToken := spotInfo.Tokens[0]
		quoteToken := spotInfo.Tokens[1]
		baseInfo := spotMeta.Tokens[baseToken]
		quoteInfo := spotMeta.Tokens[quoteToken]
		info.assetToSzDecimals[asset] = baseInfo.SzDecimals
		
		name := fmt.Sprintf("%s/%s", baseInfo.Name, quoteInfo.Name)
		if _, exists := info.nameToCoins[name]; !exists {
			info.nameToCoins[name] = spotInfo.Name
		}
	}
	
	// Process perp dexs
	perpDexToOffset := map[string]int{"": 0}
	if perpDexs == nil {
		perpDexs = []string{""}
	} else {
		perpDexsList, err := info.PerpDexs()
		if err != nil {
			return nil, fmt.Errorf("failed to get perp dexs: %w", err)
		}
		
		if perpDexsData, ok := perpDexsList.([]interface{}); ok && len(perpDexsData) > 1 {
			for i, perpDexInterface := range perpDexsData[1:] {
				if perpDex, ok := perpDexInterface.(map[string]interface{}); ok {
					if name, ok := perpDex["name"].(string); ok {
						// Builder-deployed perp dexs start at 110000
						perpDexToOffset[name] = 110000 + i*10000
					}
				}
			}
		}
	}
	
	for _, perpDex := range perpDexs {
		offset := perpDexToOffset[perpDex]
		if perpDex == "" && meta != nil {
			info.setPerpMeta(*meta, offset)
		} else {
			freshMeta, err := info.Meta(perpDex)
			if err != nil {
				return nil, fmt.Errorf("failed to get meta for dex %s: %w", perpDex, err)
			}
			info.setPerpMeta(*freshMeta, offset)
		}
	}
	
	return info, nil
}

// setPerpMeta sets perp metadata with offset
func (i *Info) setPerpMeta(meta Meta, offset int) {
	for asset, assetInfo := range meta.Universe {
		assetID := asset + offset
		i.coinToAsset[assetInfo.Name] = assetID
		i.nameToCoins[assetInfo.Name] = assetInfo.Name
		i.assetToSzDecimals[assetID] = assetInfo.SzDecimals
	}
}

// DisconnectWebSocket disconnects the WebSocket connection
func (i *Info) DisconnectWebSocket() error {
	if i.wsManager == nil {
		return fmt.Errorf("cannot disconnect WebSocket since skip_ws was used")
	}
	i.wsManager.Stop()
	return nil
}

// UserState retrieves trading details about a user
func (i *Info) UserState(address string, dex string) (interface{}, error) {
	if dex == "" {
		dex = ""
	}
	payload := map[string]interface{}{
		"type": "clearinghouseState",
		"user": address,
		"dex":  dex,
	}
	return i.Post("/info", payload)
}

// SpotUserState retrieves spot trading details about a user
func (i *Info) SpotUserState(address string) (interface{}, error) {
	payload := map[string]interface{}{
		"type": "spotClearinghouseState",
		"user": address,
	}
	return i.Post("/info", payload)
}

// OpenOrders retrieves a user's open orders
func (i *Info) OpenOrders(address string, dex string) (interface{}, error) {
	if dex == "" {
		dex = ""
	}
	payload := map[string]interface{}{
		"type": "openOrders",
		"user": address,
		"dex":  dex,
	}
	return i.Post("/info", payload)
}

// FrontendOpenOrders retrieves a user's open orders with additional frontend info
func (i *Info) FrontendOpenOrders(address string, dex string) (interface{}, error) {
	if dex == "" {
		dex = ""
	}
	payload := map[string]interface{}{
		"type": "frontendOpenOrders",
		"user": address,
		"dex":  dex,
	}
	return i.Post("/info", payload)
}

// AllMids retrieves all mids for all actively traded coins
func (i *Info) AllMids(dex string) (interface{}, error) {
	if dex == "" {
		dex = ""
	}
	payload := map[string]interface{}{
		"type": "allMids",
		"dex":  dex,
	}
	return i.Post("/info", payload)
}

// UserFills retrieves a given user's fills
func (i *Info) UserFills(address string) (interface{}, error) {
	payload := map[string]interface{}{
		"type": "userFills",
		"user": address,
	}
	return i.Post("/info", payload)
}

// UserFillsByTime retrieves a given user's fills by time
func (i *Info) UserFillsByTime(address string, startTime int64, endTime *int64) (interface{}, error) {
	payload := map[string]interface{}{
		"type":      "userFillsByTime",
		"user":      address,
		"startTime": startTime,
	}
	if endTime != nil {
		payload["endTime"] = *endTime
	}
	return i.Post("/info", payload)
}

// Meta retrieves exchange perp metadata
func (i *Info) Meta(dex string) (*Meta, error) {
	if dex == "" {
		dex = ""
	}
	payload := map[string]interface{}{
		"type": "meta",
		"dex":  dex,
	}
	result, err := i.Post("/info", payload)
	if err != nil {
		return nil, err
	}
	
	// Convert interface{} to Meta struct
	var meta Meta
	if resultMap, ok := result.(map[string]interface{}); ok {
		if universe, ok := resultMap["universe"].([]interface{}); ok {
			for _, assetInterface := range universe {
				if assetMap, ok := assetInterface.(map[string]interface{}); ok {
					assetInfo := AssetInfo{}
					if name, ok := assetMap["name"].(string); ok {
						assetInfo.Name = name
					}
					if szDecimals, ok := assetMap["szDecimals"].(float64); ok {
						assetInfo.SzDecimals = int(szDecimals)
					}
					meta.Universe = append(meta.Universe, assetInfo)
				}
			}
		}
	}
	
	return &meta, nil
}

// MetaAndAssetCtxs retrieves exchange MetaAndAssetCtxs
func (i *Info) MetaAndAssetCtxs() (interface{}, error) {
	payload := map[string]interface{}{
		"type": "metaAndAssetCtxs",
	}
	return i.Post("/info", payload)
}

// PerpDexs retrieves perp dexs
func (i *Info) PerpDexs() (interface{}, error) {
	payload := map[string]interface{}{
		"type": "perpDexs",
	}
	return i.Post("/info", payload)
}

// SpotMeta retrieves exchange spot metadata
func (i *Info) SpotMeta() (*SpotMeta, error) {
	payload := map[string]interface{}{
		"type": "spotMeta",
	}
	result, err := i.Post("/info", payload)
	if err != nil {
		return nil, err
	}
	
	// Convert interface{} to SpotMeta struct
	var spotMeta SpotMeta
	if resultMap, ok := result.(map[string]interface{}); ok {
		// Parse universe
		if universe, ok := resultMap["universe"].([]interface{}); ok {
			for _, assetInterface := range universe {
				if assetMap, ok := assetInterface.(map[string]interface{}); ok {
					assetInfo := SpotAssetInfo{}
					if name, ok := assetMap["name"].(string); ok {
						assetInfo.Name = name
					}
					if index, ok := assetMap["index"].(float64); ok {
						assetInfo.Index = int(index)
					}
					if isCanonical, ok := assetMap["isCanonical"].(bool); ok {
						assetInfo.IsCanonical = isCanonical
					}
					if tokens, ok := assetMap["tokens"].([]interface{}); ok && len(tokens) == 2 {
						if token0, ok := tokens[0].(float64); ok {
							assetInfo.Tokens[0] = int(token0)
						}
						if token1, ok := tokens[1].(float64); ok {
							assetInfo.Tokens[1] = int(token1)
						}
					}
					spotMeta.Universe = append(spotMeta.Universe, assetInfo)
				}
			}
		}
		
		// Parse tokens
		if tokens, ok := resultMap["tokens"].([]interface{}); ok {
			for _, tokenInterface := range tokens {
				if tokenMap, ok := tokenInterface.(map[string]interface{}); ok {
					tokenInfo := SpotTokenInfo{}
					if name, ok := tokenMap["name"].(string); ok {
						tokenInfo.Name = name
					}
					if szDecimals, ok := tokenMap["szDecimals"].(float64); ok {
						tokenInfo.SzDecimals = int(szDecimals)
					}
					if weiDecimals, ok := tokenMap["weiDecimals"].(float64); ok {
						tokenInfo.WeiDecimals = int(weiDecimals)
					}
					if index, ok := tokenMap["index"].(float64); ok {
						tokenInfo.Index = int(index)
					}
					if tokenID, ok := tokenMap["tokenId"].(string); ok {
						tokenInfo.TokenID = tokenID
					}
					if isCanonical, ok := tokenMap["isCanonical"].(bool); ok {
						tokenInfo.IsCanonical = isCanonical
					}
					if evmContract, ok := tokenMap["evmContract"].(string); ok {
						tokenInfo.EvmContract = &evmContract
					}
					if fullName, ok := tokenMap["fullName"].(string); ok {
						tokenInfo.FullName = &fullName
					}
					spotMeta.Tokens = append(spotMeta.Tokens, tokenInfo)
				}
			}
		}
	}
	
	return &spotMeta, nil
}

// SpotMetaAndAssetCtxs retrieves exchange spot asset contexts
func (i *Info) SpotMetaAndAssetCtxs() (interface{}, error) {
	payload := map[string]interface{}{
		"type": "spotMetaAndAssetCtxs",
	}
	return i.Post("/info", payload)
}

// FundingHistory retrieves funding history for a given coin
func (i *Info) FundingHistory(name string, startTime int64, endTime *int64) (interface{}, error) {
	coin, exists := i.nameToCoins[name]
	if !exists {
		return nil, fmt.Errorf("coin not found for name: %s", name)
	}
	
	payload := map[string]interface{}{
		"type":      "fundingHistory",
		"coin":      coin,
		"startTime": startTime,
	}
	if endTime != nil {
		payload["endTime"] = *endTime
	}
	return i.Post("/info", payload)
}

// UserFundingHistory retrieves a user's funding history
func (i *Info) UserFundingHistory(user string, startTime int64, endTime *int64) (interface{}, error) {
	payload := map[string]interface{}{
		"type":      "userFunding",
		"user":      user,
		"startTime": startTime,
	}
	if endTime != nil {
		payload["endTime"] = *endTime
	}
	return i.Post("/info", payload)
}

// L2Snapshot retrieves L2 snapshot for a given coin
func (i *Info) L2Snapshot(name string) (interface{}, error) {
	coin, exists := i.nameToCoins[name]
	if !exists {
		return nil, fmt.Errorf("coin not found for name: %s", name)
	}
	
	payload := map[string]interface{}{
		"type": "l2Book",
		"coin": coin,
	}
	return i.Post("/info", payload)
}

// CandlesSnapshot retrieves candles snapshot for a given coin
func (i *Info) CandlesSnapshot(name string, interval string, startTime int64, endTime int64) (interface{}, error) {
	coin, exists := i.nameToCoins[name]
	if !exists {
		return nil, fmt.Errorf("coin not found for name: %s", name)
	}
	
	req := map[string]interface{}{
		"coin":      coin,
		"interval":  interval,
		"startTime": startTime,
		"endTime":   endTime,
	}
	
	payload := map[string]interface{}{
		"type": "candleSnapshot",
		"req":  req,
	}
	return i.Post("/info", payload)
}

// UserFees retrieves the volume of trading activity associated with a user
func (i *Info) UserFees(address string) (interface{}, error) {
	payload := map[string]interface{}{
		"type": "userFees",
		"user": address,
	}
	return i.Post("/info", payload)
}

// UserStakingSummary retrieves the staking summary associated with a user
func (i *Info) UserStakingSummary(address string) (interface{}, error) {
	payload := map[string]interface{}{
		"type": "delegatorSummary",
		"user": address,
	}
	return i.Post("/info", payload)
}

// UserStakingDelegations retrieves the user's staking delegations
func (i *Info) UserStakingDelegations(address string) (interface{}, error) {
	payload := map[string]interface{}{
		"type": "delegations",
		"user": address,
	}
	return i.Post("/info", payload)
}

// UserStakingRewards retrieves the historic staking rewards associated with a user
func (i *Info) UserStakingRewards(address string) (interface{}, error) {
	payload := map[string]interface{}{
		"type": "delegatorRewards",
		"user": address,
	}
	return i.Post("/info", payload)
}

// QueryOrderByOID queries order by order ID
func (i *Info) QueryOrderByOID(user string, oid int) (interface{}, error) {
	payload := map[string]interface{}{
		"type": "orderStatus",
		"user": user,
		"oid":  oid,
	}
	return i.Post("/info", payload)
}

// QueryOrderByCloid queries order by client order ID
func (i *Info) QueryOrderByCloid(user string, cloid string) (interface{}, error) {
	payload := map[string]interface{}{
		"type": "orderStatus",
		"user": user,
		"oid":  cloid,
	}
	return i.Post("/info", payload)
}

// QueryReferralState queries referral state
func (i *Info) QueryReferralState(user string) (interface{}, error) {
	payload := map[string]interface{}{
		"type": "referral",
		"user": user,
	}
	return i.Post("/info", payload)
}

// QuerySubAccounts queries sub accounts
func (i *Info) QuerySubAccounts(user string) (interface{}, error) {
	payload := map[string]interface{}{
		"type": "subAccounts",
		"user": user,
	}
	return i.Post("/info", payload)
}

// QueryUserToMultiSigSigners queries user to multi-sig signers
func (i *Info) QueryUserToMultiSigSigners(multiSigUser string) (interface{}, error) {
	payload := map[string]interface{}{
		"type": "userToMultiSigSigners",
		"user": multiSigUser,
	}
	return i.Post("/info", payload)
}

// QueryPerpDeployAuctionStatus queries perp deploy auction status
func (i *Info) QueryPerpDeployAuctionStatus() (interface{}, error) {
	payload := map[string]interface{}{
		"type": "perpDeployAuctionStatus",
	}
	return i.Post("/info", payload)
}

// remapCoinSubscription remaps coin in subscription
func (i *Info) remapCoinSubscription(subscription *Subscription) {
	if subscription.Type == L2Book || subscription.Type == Trades || subscription.Type == Candle ||
		subscription.Type == BBO || subscription.Type == ActiveAssetCtx {
		if coin, exists := i.nameToCoins[subscription.Coin]; exists {
			subscription.Coin = coin
		}
	}
}

// Subscribe subscribes to a WebSocket channel
func (i *Info) Subscribe(subscription Subscription, callback func(WsMsg)) (int, error) {
	i.remapCoinSubscription(&subscription)
	if i.wsManager == nil {
		return 0, fmt.Errorf("cannot subscribe since skip_ws was used")
	}
	return i.wsManager.Subscribe(subscription, callback), nil
}

// Unsubscribe unsubscribes from a WebSocket channel
func (i *Info) Unsubscribe(subscription Subscription, subscriptionID int) (bool, error) {
	i.remapCoinSubscription(&subscription)
	if i.wsManager == nil {
		return false, fmt.Errorf("cannot unsubscribe since skip_ws was used")
	}
	return i.wsManager.Unsubscribe(subscription, subscriptionID), nil
}

// NameToAsset converts name to asset ID
func (i *Info) NameToAsset(name string) (int, error) {
	if coin, exists := i.nameToCoins[name]; exists {
		if asset, exists := i.coinToAsset[coin]; exists {
			return asset, nil
		}
	}
	return 0, fmt.Errorf("asset not found for name: %s", name)
}
