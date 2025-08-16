package utils

import (
	"fmt"
	"strconv"
	"strings"
)

// Side represents trading side (Ask or Bid)
type Side string

const (
	SideAsk Side = "A" // Ask
	SideBid Side = "B" // Bid
)

var Sides = []Side{SideAsk, SideBid}

// AssetInfo represents basic asset information
type AssetInfo struct {
	Name       string `json:"name"`
	SzDecimals int    `json:"szDecimals"`
}

// Meta contains universe of assets
type Meta struct {
	Universe []AssetInfo `json:"universe"`
}

// SpotAssetInfo represents spot asset information
type SpotAssetInfo struct {
	Name        string `json:"name"`
	Tokens      []int  `json:"tokens"`
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

// SpotMeta contains spot asset and token information
type SpotMeta struct {
	Universe []SpotAssetInfo `json:"universe"`
	Tokens   []SpotTokenInfo `json:"tokens"`
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

// SpotMetaAndAssetCtxs combines spot meta and asset contexts
type SpotMetaAndAssetCtxs struct {
	Meta     SpotMeta
	AssetCtx []SpotAssetCtx
}

// Subscription types
type SubscriptionType string

const (
	SubTypeAllMids                        SubscriptionType = "allMids"
	SubTypeBbo                           SubscriptionType = "bbo"
	SubTypeL2Book                        SubscriptionType = "l2Book"
	SubTypeTrades                        SubscriptionType = "trades"
	SubTypeUserEvents                    SubscriptionType = "userEvents"
	SubTypeUserFills                     SubscriptionType = "userFills"
	SubTypeCandle                        SubscriptionType = "candle"
	SubTypeOrderUpdates                  SubscriptionType = "orderUpdates"
	SubTypeUserFundings                  SubscriptionType = "userFundings"
	SubTypeUserNonFundingLedgerUpdates   SubscriptionType = "userNonFundingLedgerUpdates"
	SubTypeWebData2                      SubscriptionType = "webData2"
	SubTypeActiveAssetCtx                SubscriptionType = "activeAssetCtx"
	SubTypeActiveAssetData               SubscriptionType = "activeAssetData"
)

// Base subscription interface
type Subscription interface {
	GetType() SubscriptionType
}

// AllMidsSubscription for all mids
type AllMidsSubscription struct {
	Type SubscriptionType `json:"type"`
}

func (s AllMidsSubscription) GetType() SubscriptionType { return s.Type }

// BboSubscription for best bid/offer
type BboSubscription struct {
	Type SubscriptionType `json:"type"`
	Coin string           `json:"coin"`
}

func (s BboSubscription) GetType() SubscriptionType { return s.Type }

// L2BookSubscription for level 2 order book
type L2BookSubscription struct {
	Type SubscriptionType `json:"type"`
	Coin string           `json:"coin"`
}

func (s L2BookSubscription) GetType() SubscriptionType { return s.Type }

// TradesSubscription for trades
type TradesSubscription struct {
	Type SubscriptionType `json:"type"`
	Coin string           `json:"coin"`
}

func (s TradesSubscription) GetType() SubscriptionType { return s.Type }

// UserEventsSubscription for user events
type UserEventsSubscription struct {
	Type SubscriptionType `json:"type"`
	User string           `json:"user"`
}

func (s UserEventsSubscription) GetType() SubscriptionType { return s.Type }

// UserFillsSubscription for user fills
type UserFillsSubscription struct {
	Type SubscriptionType `json:"type"`
	User string           `json:"user"`
}

func (s UserFillsSubscription) GetType() SubscriptionType { return s.Type }

// CandleSubscription for candles
type CandleSubscription struct {
	Type     SubscriptionType `json:"type"`
	Coin     string           `json:"coin"`
	Interval string           `json:"interval"`
}

func (s CandleSubscription) GetType() SubscriptionType { return s.Type }

// OrderUpdatesSubscription for order updates
type OrderUpdatesSubscription struct {
	Type SubscriptionType `json:"type"`
	User string           `json:"user"`
}

func (s OrderUpdatesSubscription) GetType() SubscriptionType { return s.Type }

// UserFundingsSubscription for user fundings
type UserFundingsSubscription struct {
	Type SubscriptionType `json:"type"`
	User string           `json:"user"`
}

func (s UserFundingsSubscription) GetType() SubscriptionType { return s.Type }

// UserNonFundingLedgerUpdatesSubscription for user non-funding ledger updates
type UserNonFundingLedgerUpdatesSubscription struct {
	Type SubscriptionType `json:"type"`
	User string           `json:"user"`
}

func (s UserNonFundingLedgerUpdatesSubscription) GetType() SubscriptionType { return s.Type }

// WebData2Subscription for web data 2
type WebData2Subscription struct {
	Type SubscriptionType `json:"type"`
	User string           `json:"user"`
}

func (s WebData2Subscription) GetType() SubscriptionType { return s.Type }

// ActiveAssetCtxSubscription for active asset context
type ActiveAssetCtxSubscription struct {
	Type SubscriptionType `json:"type"`
	Coin string           `json:"coin"`
}

func (s ActiveAssetCtxSubscription) GetType() SubscriptionType { return s.Type }

// ActiveAssetDataSubscription for active asset data
type ActiveAssetDataSubscription struct {
	Type SubscriptionType `json:"type"`
	User string           `json:"user"`
	Coin string           `json:"coin"`
}

func (s ActiveAssetDataSubscription) GetType() SubscriptionType { return s.Type }

// WebSocket message data types

// AllMidsData contains mid prices for all assets
type AllMidsData struct {
	Mids map[string]string `json:"mids"`
}

// AllMidsMsg is the message for all mids
type AllMidsMsg struct {
	Channel string      `json:"channel"`
	Data    AllMidsData `json:"data"`
}

// L2Level represents a level in the order book
type L2Level struct {
	Px string `json:"px"` // Price
	Sz string `json:"sz"` // Size
	N  int    `json:"n"`  // Number of orders
}

// L2BookData contains level 2 order book data
type L2BookData struct {
	Coin   string      `json:"coin"`
	Levels [2][]L2Level `json:"levels"` // [bids, asks]
	Time   int64       `json:"time"`
}

// L2BookMsg is the message for level 2 order book
type L2BookMsg struct {
	Channel string     `json:"channel"`
	Data    L2BookData `json:"data"`
}

// BboData contains best bid/offer data
type BboData struct {
	Coin string     `json:"coin"`
	Time int64      `json:"time"`
	Bbo  [2]*L2Level `json:"bbo"` // [bid, ask]
}

// BboMsg is the message for best bid/offer
type BboMsg struct {
	Channel string  `json:"channel"`
	Data    BboData `json:"data"`
}

// PongMsg is the pong response message
type PongMsg struct {
	Channel string `json:"channel"`
}

// Trade represents a trade
type Trade struct {
	Coin string `json:"coin"`
	Side Side   `json:"side"`
	Px   string `json:"px"`   // Price
	Sz   int    `json:"sz"`   // Size
	Hash string `json:"hash"`
	Time int64  `json:"time"`
}

// Leverage types
type LeverageType string

const (
	LeverageTypeCross    LeverageType = "cross"
	LeverageTypeIsolated LeverageType = "isolated"
)

// CrossLeverage represents cross leverage
type CrossLeverage struct {
	Type  LeverageType `json:"type"`
	Value int          `json:"value"`
}

// IsolatedLeverage represents isolated leverage
type IsolatedLeverage struct {
	Type   LeverageType `json:"type"`
	Value  int          `json:"value"`
	RawUsd string       `json:"rawUsd"`
}

// Leverage interface
type Leverage interface {
	GetType() LeverageType
}

func (l CrossLeverage) GetType() LeverageType    { return l.Type }
func (l IsolatedLeverage) GetType() LeverageType { return l.Type }

// TradesMsg is the message for trades
type TradesMsg struct {
	Channel string  `json:"channel"`
	Data    []Trade `json:"data"`
}

// PerpAssetCtx represents perpetual asset context
type PerpAssetCtx struct {
	Funding      string    `json:"funding"`
	OpenInterest string    `json:"openInterest"`
	PrevDayPx    string    `json:"prevDayPx"`
	DayNtlVlm    string    `json:"dayNtlVlm"`
	Premium      string    `json:"premium"`
	OraclePx     string    `json:"oraclePx"`
	MarkPx       string    `json:"markPx"`
	MidPx        *string   `json:"midPx,omitempty"`
	ImpactPxs    *[2]string `json:"impactPxs,omitempty"`
	DayBaseVlm   string    `json:"dayBaseVlm"`
}

// ActiveAssetCtx represents active asset context
type ActiveAssetCtx struct {
	Coin string       `json:"coin"`
	Ctx  PerpAssetCtx `json:"ctx"`
}

// ActiveSpotAssetCtx represents active spot asset context
type ActiveSpotAssetCtx struct {
	Coin string       `json:"coin"`
	Ctx  SpotAssetCtx `json:"ctx"`
}

// ActiveAssetCtxMsg is the message for active asset context
type ActiveAssetCtxMsg struct {
	Channel string         `json:"channel"`
	Data    ActiveAssetCtx `json:"data"`
}

// ActiveSpotAssetCtxMsg is the message for active spot asset context
type ActiveSpotAssetCtxMsg struct {
	Channel string             `json:"channel"`
	Data    ActiveSpotAssetCtx `json:"data"`
}

// ActiveAssetData represents active asset data
type ActiveAssetData struct {
	User              string     `json:"user"`
	Coin              string     `json:"coin"`
	Leverage          Leverage   `json:"leverage"`
	MaxTradeSzs       [2]string  `json:"maxTradeSzs"`
	AvailableToTrade  [2]string  `json:"availableToTrade"`
	MarkPx            string     `json:"markPx"`
}

// ActiveAssetDataMsg is the message for active asset data
type ActiveAssetDataMsg struct {
	Channel string          `json:"channel"`
	Data    ActiveAssetData `json:"data"`
}

// Fill represents a trade fill
type Fill struct {
	Coin          string `json:"coin"`
	Px            string `json:"px"`
	Sz            string `json:"sz"`
	Side          Side   `json:"side"`
	Time          int64  `json:"time"`
	StartPosition string `json:"startPosition"`
	Dir           string `json:"dir"`
	ClosedPnl     string `json:"closedPnl"`
	Hash          string `json:"hash"`
	Oid           int    `json:"oid"`
	Crossed       bool   `json:"crossed"`
	Fee           string `json:"fee"`
	Tid           int    `json:"tid"`
	FeeToken      string `json:"feeToken"`
}

// UserEventsData contains user event data
type UserEventsData struct {
	Fills []Fill `json:"fills,omitempty"`
}

// UserEventsMsg is the message for user events
type UserEventsMsg struct {
	Channel string         `json:"channel"`
	Data    UserEventsData `json:"data"`
}

// UserFillsData contains user fills data
type UserFillsData struct {
	User       string `json:"user"`
	IsSnapshot bool   `json:"isSnapshot"`
	Fills      []Fill `json:"fills"`
}

// UserFillsMsg is the message for user fills
type UserFillsMsg struct {
	Channel string        `json:"channel"`
	Data    UserFillsData `json:"data"`
}

// OtherWsMsg represents other WebSocket messages
type OtherWsMsg struct {
	Channel string      `json:"channel"`
	Data    interface{} `json:"data,omitempty"`
}

// WsMsg represents any WebSocket message
type WsMsg interface {
	GetChannel() string
}

// Implement GetChannel for all message types
func (m AllMidsMsg) GetChannel() string           { return m.Channel }
func (m BboMsg) GetChannel() string               { return m.Channel }
func (m L2BookMsg) GetChannel() string            { return m.Channel }
func (m TradesMsg) GetChannel() string            { return m.Channel }
func (m UserEventsMsg) GetChannel() string        { return m.Channel }
func (m PongMsg) GetChannel() string              { return m.Channel }
func (m UserFillsMsg) GetChannel() string         { return m.Channel }
func (m OtherWsMsg) GetChannel() string           { return m.Channel }
func (m ActiveAssetCtxMsg) GetChannel() string    { return m.Channel }
func (m ActiveSpotAssetCtxMsg) GetChannel() string { return m.Channel }
func (m ActiveAssetDataMsg) GetChannel() string   { return m.Channel }

// BuilderInfo represents builder information
type BuilderInfo struct {
	B string `json:"b"` // Public address of the builder
	F int    `json:"f"` // Fee in tenths of basis points
}

// PerpDexSchemaInput represents perpetual DEX schema input
type PerpDexSchemaInput struct {
	FullName         string  `json:"fullName"`
	CollateralToken  int     `json:"collateralToken"`
	OracleUpdater    *string `json:"oracleUpdater,omitempty"`
}

// Cloid represents a client order ID
type Cloid struct {
	rawCloid string
}

// NewCloid creates a new Cloid from a hex string
func NewCloid(rawCloid string) (*Cloid, error) {
	c := &Cloid{rawCloid: rawCloid}
	if err := c.validate(); err != nil {
		return nil, err
	}
	return c, nil
}

// NewCloidFromInt creates a new Cloid from an integer
func NewCloidFromInt(cloid int) *Cloid {
	return &Cloid{rawCloid: fmt.Sprintf("%#034x", cloid)}
}

// NewCloidFromStr creates a new Cloid from a string
func NewCloidFromStr(cloid string) (*Cloid, error) {
	return NewCloid(cloid)
}

// validate checks if the cloid is valid
func (c *Cloid) validate() error {
	if !strings.HasPrefix(c.rawCloid, "0x") {
		return fmt.Errorf("cloid is not a hex string")
	}
	if len(c.rawCloid[2:]) != 32 {
		return fmt.Errorf("cloid is not 16 bytes")
	}
	return nil
}

// String returns the string representation of the cloid
func (c *Cloid) String() string {
	return c.rawCloid
}

// ToRaw returns the raw cloid string
func (c *Cloid) ToRaw() string {
	return c.rawCloid
}

// ToInt converts the cloid to an integer
func (c *Cloid) ToInt() (int64, error) {
	return strconv.ParseInt(c.rawCloid, 0, 64)
}
