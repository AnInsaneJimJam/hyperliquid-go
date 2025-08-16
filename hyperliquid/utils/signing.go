package utils

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	ethmath "github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/vmihailenco/msgpack/v5"
	"golang.org/x/crypto/sha3"
)

// TIF represents Time In Force for orders
type TIF string

const (
	TIFAlo TIF = "Alo" // Add Liquidity Only
	TIFIoc TIF = "Ioc" // Immediate Or Cancel
	TIFGtc TIF = "Gtc" // Good Till Cancel
)

// TPSL represents Take Profit / Stop Loss
type TPSL string

const (
	TPSLTp TPSL = "tp" // Take Profit
	TPSLSl TPSL = "sl" // Stop Loss
)


// Grouping represents order grouping types
type Grouping string

const (
	GroupingNA           Grouping = "na"
	GroupingNormalTpsl   Grouping = "normalTpsl"
	GroupingPositionTpsl Grouping = "positionTpsl"
)

// LimitOrderType represents a limit order configuration
type LimitOrderType struct {
	TIF TIF `json:"tif"`
}

// TriggerOrderType represents a trigger order configuration
type TriggerOrderType struct {
	TriggerPx float64 `json:"triggerPx"`
	IsMarket  bool    `json:"isMarket"`
	TPSL      TPSL    `json:"tpsl"`
}

// TriggerOrderTypeWire represents a trigger order for wire format
type TriggerOrderTypeWire struct {
	TriggerPx string `json:"triggerPx"`
	IsMarket  bool   `json:"isMarket"`
	TPSL      TPSL   `json:"tpsl"`
}

// OrderType represents the type of order (limit or trigger)
type OrderType struct {
	Limit   *LimitOrderType   `json:"limit,omitempty"`
	Trigger *TriggerOrderType `json:"trigger,omitempty"`
}

// OrderTypeWire represents the wire format of order type
type OrderTypeWire struct {
	Limit   *LimitOrderType       `json:"limit,omitempty"`
	Trigger *TriggerOrderTypeWire `json:"trigger,omitempty"`
}

// Order represents a simplified order structure
type Order struct {
	Asset      int     `json:"asset"`
	IsBuy      bool    `json:"isBuy"`
	LimitPx    float64 `json:"limitPx"`
	Sz         float64 `json:"sz"`
	ReduceOnly bool    `json:"reduceOnly"`
	Cloid      *string `json:"cloid,omitempty"`
}

// OrderRequest represents a request to place an order
type OrderRequest struct {
	Coin       string     `json:"coin"`
	IsBuy      bool       `json:"is_buy"`
	Sz         float64    `json:"sz"`
	LimitPx    float64    `json:"limit_px"`
	OrderType  OrderType  `json:"order_type"`
	ReduceOnly bool       `json:"reduce_only"`
	Cloid      *string    `json:"cloid,omitempty"`
}

// OrderWire represents the wire format of an order
type OrderWire struct {
	A int            `json:"a"`      // asset
	B bool           `json:"b"`      // is_buy
	P string         `json:"p"`      // price
	S string         `json:"s"`      // size
	R bool           `json:"r"`      // reduce_only
	T OrderTypeWire  `json:"t"`      // order_type
	C *string        `json:"c,omitempty"` // cloid
}

// ModifyRequest represents a request to modify an order
type ModifyRequest struct {
	OID   interface{}  `json:"oid"` // can be int or string (cloid)
	Order OrderRequest `json:"order"`
}

// ModifyWire represents the wire format of a modify request
type ModifyWire struct {
	OID   int       `json:"oid"`
	Order OrderWire `json:"order"`
}

// CancelRequest represents a request to cancel an order
type CancelRequest struct {
	Coin string `json:"coin"`
	OID  int    `json:"oid"`
}

// CancelByCloidRequest represents a request to cancel an order by cloid
type CancelByCloidRequest struct {
	Coin  string `json:"coin"`
	Cloid string `json:"cloid"`
}

// ScheduleCancelAction represents a scheduled cancel action
type ScheduleCancelAction struct {
	Type string `json:"type"`
	Time *int64 `json:"time,omitempty"`
}

// EIP712 type definitions for various signing operations
var (
	USDSendSignTypes = []apitypes.Type{
		{Name: "hyperliquidChain", Type: "string"},
		{Name: "destination", Type: "string"},
		{Name: "amount", Type: "string"},
		{Name: "time", Type: "uint64"},
	}

	SpotTransferSignTypes = []apitypes.Type{
		{Name: "hyperliquidChain", Type: "string"},
		{Name: "destination", Type: "string"},
		{Name: "token", Type: "string"},
		{Name: "amount", Type: "string"},
		{Name: "time", Type: "uint64"},
	}

	WithdrawSignTypes = []apitypes.Type{
		{Name: "hyperliquidChain", Type: "string"},
		{Name: "destination", Type: "string"},
		{Name: "amount", Type: "string"},
		{Name: "time", Type: "uint64"},
	}

	USDClassTransferSignTypes = []apitypes.Type{
		{Name: "hyperliquidChain", Type: "string"},
		{Name: "amount", Type: "string"},
		{Name: "toPerp", Type: "bool"},
		{Name: "nonce", Type: "uint64"},
	}

	SendAssetSignTypes = []apitypes.Type{
		{Name: "hyperliquidChain", Type: "string"},
		{Name: "destination", Type: "string"},
		{Name: "sourceDex", Type: "string"},
		{Name: "destinationDex", Type: "string"},
		{Name: "token", Type: "string"},
		{Name: "amount", Type: "string"},
		{Name: "fromSubAccount", Type: "string"},
		{Name: "nonce", Type: "uint64"},
	}

	TokenDelegateTypes = []apitypes.Type{
		{Name: "hyperliquidChain", Type: "string"},
		{Name: "validator", Type: "address"},
		{Name: "wei", Type: "uint64"},
		{Name: "isUndelegate", Type: "bool"},
		{Name: "nonce", Type: "uint64"},
	}

	ConvertToMultiSigUserSignTypes = []apitypes.Type{
		{Name: "hyperliquidChain", Type: "string"},
		{Name: "signers", Type: "string"},
		{Name: "nonce", Type: "uint64"},
	}

	MultiSigEnvelopeSignTypes = []apitypes.Type{
		{Name: "hyperliquidChain", Type: "string"},
		{Name: "multiSigActionHash", Type: "bytes32"},
		{Name: "nonce", Type: "uint64"},
	}
)

// Signature represents an Ethereum signature
type Signature struct {
	R string `json:"r"`
	S string `json:"s"`
	V uint8  `json:"v"`
}

// PhantomAgent represents a phantom agent for L1 actions
type PhantomAgent struct {
	Source       string `json:"source"`
	ConnectionID string `json:"connectionId"`
}

// FloatToWire converts a float to wire format string with proper precision
func FloatToWire(x float64) (string, error) {
	rounded := fmt.Sprintf("%.8f", x)
	parsedRounded, err := strconv.ParseFloat(rounded, 64)
	if err != nil {
		return "", err
	}
	
	if math.Abs(parsedRounded-x) >= 1e-12 {
		return "", fmt.Errorf("float_to_wire causes rounding: %f", x)
	}
	
	if rounded == "-0.00000000" {
		rounded = "0.00000000"
	}
	
	// Remove trailing zeros and decimal point if not needed
	trimmed := strings.TrimRight(rounded, "0")
	trimmed = strings.TrimRight(trimmed, ".")
	if trimmed == "" {
		trimmed = "0"
	}
	
	return trimmed, nil
}

// FloatToIntForHashing converts float to int for hashing with 8 decimal places
func FloatToIntForHashing(x float64) (int64, error) {
	return FloatToInt(x, 8)
}

// FloatToUSDInt converts float to int for USD with 6 decimal places
func FloatToUSDInt(x float64) (int64, error) {
	return FloatToInt(x, 6)
}

// FloatToInt converts float to int with specified decimal places
func FloatToInt(x float64, power int) (int64, error) {
	withDecimals := x * math.Pow(10, float64(power))
	rounded := math.Round(withDecimals)
	
	if math.Abs(rounded-withDecimals) >= 1e-3 {
		return 0, fmt.Errorf("float_to_int causes rounding: %f", x)
	}
	
	return int64(rounded), nil
}

// GetTimestampMs returns current timestamp in milliseconds
func GetTimestampMs() int64 {
	return time.Now().UnixMilli()
}

// OrderTypeToWire converts OrderType to wire format
func OrderTypeToWire(orderType OrderType) (OrderTypeWire, error) {
	if orderType.Limit != nil {
		return OrderTypeWire{Limit: orderType.Limit}, nil
	} else if orderType.Trigger != nil {
		triggerPxWire, err := FloatToWire(orderType.Trigger.TriggerPx)
		if err != nil {
			return OrderTypeWire{}, err
		}
		return OrderTypeWire{
			Trigger: &TriggerOrderTypeWire{
				IsMarket:  orderType.Trigger.IsMarket,
				TriggerPx: triggerPxWire,
				TPSL:      orderType.Trigger.TPSL,
			},
		}, nil
	}
	return OrderTypeWire{}, fmt.Errorf("invalid order type")
}

// AddressToBytes converts hex address to bytes
func AddressToBytes(address string) ([]byte, error) {
	address = strings.TrimPrefix(address, "0x")
	return hex.DecodeString(address)
}

// ActionHash computes the hash of an action for L1 signing
func ActionHash(action interface{}, vaultAddress *string, nonce uint64, expiresAfter *uint64) ([]byte, error) {
	data, err := msgpack.Marshal(action)
	if err != nil {
		return nil, err
	}
	
	// Add nonce (8 bytes, big endian)
	nonceBytes := make([]byte, 8)
	for i := 7; i >= 0; i-- {
		nonceBytes[i] = byte(nonce & 0xff)
		nonce >>= 8
	}
	data = append(data, nonceBytes...)
	
	// Add vault address
	if vaultAddress == nil {
		data = append(data, 0x00)
	} else {
		data = append(data, 0x01)
		vaultBytes, err := AddressToBytes(*vaultAddress)
		if err != nil {
			return nil, err
		}
		data = append(data, vaultBytes...)
	}
	
	// Add expires after if present
	if expiresAfter != nil {
		data = append(data, 0x00)
		expiresBytes := make([]byte, 8)
		expires := *expiresAfter
		for i := 7; i >= 0; i-- {
			expiresBytes[i] = byte(expires & 0xff)
			expires >>= 8
		}
		data = append(data, expiresBytes...)
	}
	
	hash := sha3.NewLegacyKeccak256()
	hash.Write(data)
	return hash.Sum(nil), nil
}

// ConstructPhantomAgent creates a phantom agent for L1 actions
func ConstructPhantomAgent(hash []byte, isMainnet bool) PhantomAgent {
	source := "b" // testnet
	if isMainnet {
		source = "a" // mainnet
	}
	return PhantomAgent{
		Source:       source,
		ConnectionID: hexutil.Encode(hash),
	}
}

// L1Payload creates the EIP712 payload for L1 actions
func L1Payload(phantomAgent PhantomAgent) apitypes.TypedData {
	return apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": []apitypes.Type{
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
			"Agent": []apitypes.Type{
				{Name: "source", Type: "string"},
				{Name: "connectionId", Type: "bytes32"},
			},
		},
		PrimaryType: "Agent",
		Domain: apitypes.TypedDataDomain{
			Name:              "Exchange",
			Version:           "1",
			ChainId:           (*ethmath.HexOrDecimal256)(big.NewInt(1337)),
			VerifyingContract: "0x0000000000000000000000000000000000000000",
		},
		Message: apitypes.TypedDataMessage{
			"source":       phantomAgent.Source,
			"connectionId": phantomAgent.ConnectionID,
		},
	}
}

// UserSignedPayload creates the EIP712 payload for user-signed actions
func UserSignedPayload(primaryType string, payloadTypes []apitypes.Type, action map[string]interface{}) (apitypes.TypedData, error) {
	chainIDStr, ok := action["signatureChainId"].(string)
	if !ok {
		return apitypes.TypedData{}, fmt.Errorf("signatureChainId not found or not string")
	}
	
	chainID, err := strconv.ParseInt(chainIDStr, 0, 64)
	if err != nil {
		return apitypes.TypedData{}, err
	}
	
	types := apitypes.Types{
		"EIP712Domain": []apitypes.Type{
			{Name: "name", Type: "string"},
			{Name: "version", Type: "string"},
			{Name: "chainId", Type: "uint256"},
			{Name: "verifyingContract", Type: "address"},
		},
		primaryType: payloadTypes,
	}
	
	message := make(apitypes.TypedDataMessage)
	for k, v := range action {
		message[k] = v
	}
	
	return apitypes.TypedData{
		Types:       types,
		PrimaryType: primaryType,
		Domain: apitypes.TypedDataDomain{
			Name:              "HyperliquidSignTransaction",
			Version:           "1",
			ChainId:           (*ethmath.HexOrDecimal256)(big.NewInt(chainID)),
			VerifyingContract: "0x0000000000000000000000000000000000000000",
		},
		Message: message,
	}, nil
}

// SignInner performs the actual EIP712 signing
func SignInner(privateKey *ecdsa.PrivateKey, data apitypes.TypedData) (*Signature, error) {
	domainSeparator, err := data.HashStruct("EIP712Domain", data.Domain.Map())
	if err != nil {
		return nil, err
	}
	
	typedDataHash, err := data.HashStruct(data.PrimaryType, data.Message)
	if err != nil {
		return nil, err
	}
	
	// EIP712 signing: keccak256("\x19\x01" + domainSeparator + structHash)
	rawData := append([]byte("\x19\x01"), append(domainSeparator, typedDataHash...)...)
	hash := crypto.Keccak256Hash(rawData)
	
	signature, err := crypto.Sign(hash.Bytes(), privateKey)
	if err != nil {
		return nil, err
	}
	
	r := hexutil.Encode(signature[:32])
	s := hexutil.Encode(signature[32:64])
	v := signature[64] + 27
	
	return &Signature{
		R: r,
		S: s,
		V: v,
	}, nil
}

// SignL1Action signs an L1 action
func SignL1Action(privateKey *ecdsa.PrivateKey, action interface{}, activePool *string, nonce uint64, expiresAfter *uint64, isMainnet bool) (*Signature, error) {
	hash, err := ActionHash(action, activePool, nonce, expiresAfter)
	if err != nil {
		return nil, err
	}
	
	phantomAgent := ConstructPhantomAgent(hash, isMainnet)
	data := L1Payload(phantomAgent)
	
	return SignInner(privateKey, data)
}

// SignUserSignedAction signs a user-signed action
func SignUserSignedAction(privateKey *ecdsa.PrivateKey, action map[string]interface{}, payloadTypes []apitypes.Type, primaryType string, isMainnet bool) (*Signature, error) {
	// Set signature chain ID and hyperliquid chain
	action["signatureChainId"] = "0x66eee"
	if isMainnet {
		action["hyperliquidChain"] = "Mainnet"
	} else {
		action["hyperliquidChain"] = "Testnet"
	}
	
	data, err := UserSignedPayload(primaryType, payloadTypes, action)
	if err != nil {
		return nil, err
	}
	
	return SignInner(privateKey, data)
}

// OrderRequestToOrderWire converts an OrderRequest to wire format
func OrderRequestToOrderWire(order OrderRequest, asset int) (*OrderWire, error) {
	limitPxWire, err := FloatToWire(order.LimitPx)
	if err != nil {
		return nil, err
	}
	
	szWire, err := FloatToWire(order.Sz)
	if err != nil {
		return nil, err
	}
	
	orderTypeWire, err := OrderTypeToWire(order.OrderType)
	if err != nil {
		return nil, err
	}
	
	orderWire := &OrderWire{
		A: asset,
		B: order.IsBuy,
		P: limitPxWire,
		S: szWire,
		R: order.ReduceOnly,
		T: orderTypeWire,
	}
	
	if order.Cloid != nil {
		orderWire.C = order.Cloid
	}
	
	return orderWire, nil
}

// OrderWiresToOrderAction converts order wires to an order action
func OrderWiresToOrderAction(orderWires []OrderWire, builder *string) map[string]interface{} {
	action := map[string]interface{}{
		"type":     "order",
		"orders":   orderWires,
		"grouping": "na",
	}
	
	if builder != nil {
		action["builder"] = *builder
	}
	
	return action
}

// Specific signing functions for different action types

// SignUSDTransferAction signs a USD transfer action
func SignUSDTransferAction(privateKey *ecdsa.PrivateKey, action map[string]interface{}, isMainnet bool) (*Signature, error) {
	return SignUserSignedAction(privateKey, action, USDSendSignTypes, "HyperliquidTransaction:UsdSend", isMainnet)
}

// SignSpotTransferAction signs a spot transfer action
func SignSpotTransferAction(privateKey *ecdsa.PrivateKey, action map[string]interface{}, isMainnet bool) (*Signature, error) {
	return SignUserSignedAction(privateKey, action, SpotTransferSignTypes, "HyperliquidTransaction:SpotSend", isMainnet)
}

// SignWithdrawFromBridgeAction signs a withdraw from bridge action
func SignWithdrawFromBridgeAction(privateKey *ecdsa.PrivateKey, action map[string]interface{}, isMainnet bool) (*Signature, error) {
	return SignUserSignedAction(privateKey, action, WithdrawSignTypes, "HyperliquidTransaction:Withdraw", isMainnet)
}

// SignUSDClassTransferAction signs a USD class transfer action
func SignUSDClassTransferAction(privateKey *ecdsa.PrivateKey, action map[string]interface{}, isMainnet bool) (*Signature, error) {
	return SignUserSignedAction(privateKey, action, USDClassTransferSignTypes, "HyperliquidTransaction:UsdClassTransfer", isMainnet)
}

// SignSendAssetAction signs a send asset action
func SignSendAssetAction(privateKey *ecdsa.PrivateKey, action map[string]interface{}, isMainnet bool) (*Signature, error) {
	return SignUserSignedAction(privateKey, action, SendAssetSignTypes, "HyperliquidTransaction:SendAsset", isMainnet)
}

// SignConvertToMultiSigUserAction signs a convert to multi-sig user action
func SignConvertToMultiSigUserAction(privateKey *ecdsa.PrivateKey, action map[string]interface{}, isMainnet bool) (*Signature, error) {
	return SignUserSignedAction(privateKey, action, ConvertToMultiSigUserSignTypes, "HyperliquidTransaction:ConvertToMultiSigUser", isMainnet)
}

// SignAgent signs an agent approval action
func SignAgent(privateKey *ecdsa.PrivateKey, action map[string]interface{}, isMainnet bool) (*Signature, error) {
	agentSignTypes := []apitypes.Type{
		{Name: "hyperliquidChain", Type: "string"},
		{Name: "agentAddress", Type: "address"},
		{Name: "agentName", Type: "string"},
		{Name: "nonce", Type: "uint64"},
	}
	return SignUserSignedAction(privateKey, action, agentSignTypes, "HyperliquidTransaction:ApproveAgent", isMainnet)
}

// SignApproveBuilderFee signs an approve builder fee action
func SignApproveBuilderFee(privateKey *ecdsa.PrivateKey, action map[string]interface{}, isMainnet bool) (*Signature, error) {
	builderFeeSignTypes := []apitypes.Type{
		{Name: "hyperliquidChain", Type: "string"},
		{Name: "maxFeeRate", Type: "string"},
		{Name: "builder", Type: "address"},
		{Name: "nonce", Type: "uint64"},
	}
	return SignUserSignedAction(privateKey, action, builderFeeSignTypes, "HyperliquidTransaction:ApproveBuilderFee", isMainnet)
}

// SignTokenDelegateAction signs a token delegate action
func SignTokenDelegateAction(privateKey *ecdsa.PrivateKey, action map[string]interface{}, isMainnet bool) (*Signature, error) {
	return SignUserSignedAction(privateKey, action, TokenDelegateTypes, "HyperliquidTransaction:TokenDelegate", isMainnet)
}
