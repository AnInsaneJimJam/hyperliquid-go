// Package tests - Signing functionality tests
package tests

import (
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/hyperliquid-go/hyperliquid-go/hyperliquid/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFloatToWire(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected string
		hasError bool
	}{
		{"Simple decimal", 1.5, "1.5", false},
		{"Zero", 0.0, "0", false},
		{"Multiple decimals", 123.456789, "123.456789", false},
		{"Small number", 0.00000001, "0.00000001", false},
		{"Large number", 1000000.0, "1000000", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := utils.FloatToWire(tt.input)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestFloatToInt(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		power    int
		expected int64
		hasError bool
	}{
		{"Basic conversion", 1.5, 1, 15, false},
		{"USD conversion", 100.123456, 6, 100123456, false},
		{"Hash conversion", 0.12345678, 8, 12345678, false},
		{"Zero", 0.0, 8, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := utils.FloatToInt(tt.input, tt.power)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestOrderTypeToWire(t *testing.T) {
	// Test limit order
	limitOrder := utils.OrderType{
		Limit: &utils.LimitOrderType{
			TIF: utils.TIFGtc,
		},
	}
	
	wireOrder, err := utils.OrderTypeToWire(limitOrder)
	require.NoError(t, err)
	assert.NotNil(t, wireOrder.Limit)
	assert.Equal(t, utils.TIFGtc, wireOrder.Limit.TIF)

	// Test trigger order
	triggerOrder := utils.OrderType{
		Trigger: &utils.TriggerOrderType{
			TriggerPx: 100.5,
			IsMarket:  true,
			TPSL:      utils.TPSLTp,
		},
	}
	
	wireTrigger, err := utils.OrderTypeToWire(triggerOrder)
	require.NoError(t, err)
	assert.NotNil(t, wireTrigger.Trigger)
	assert.Equal(t, "100.5", wireTrigger.Trigger.TriggerPx)
	assert.True(t, wireTrigger.Trigger.IsMarket)
	assert.Equal(t, utils.TPSLTp, wireTrigger.Trigger.TPSL)
}

func TestOrderRequestToOrderWire(t *testing.T) {
	cloid := "test-cloid"
	orderRequest := utils.OrderRequest{
		Coin:       "BTC",
		IsBuy:      true,
		Sz:         1.5,
		LimitPx:    50000.0,
		OrderType: utils.OrderType{
			Limit: &utils.LimitOrderType{
				TIF: utils.TIFGtc,
			},
		},
		ReduceOnly: false,
		Cloid:      &cloid,
	}

	orderWire, err := utils.OrderRequestToOrderWire(orderRequest, 0)
	require.NoError(t, err)
	
	assert.Equal(t, 0, orderWire.A)
	assert.True(t, orderWire.B)
	assert.Equal(t, "50000", orderWire.P)
	assert.Equal(t, "1.5", orderWire.S)
	assert.False(t, orderWire.R)
	assert.NotNil(t, orderWire.C)
	assert.Equal(t, "test-cloid", *orderWire.C)
}

func TestConstructPhantomAgent(t *testing.T) {
	hash := []byte{0x01, 0x02, 0x03, 0x04}
	
	// Test mainnet
	agentMainnet := utils.ConstructPhantomAgent(hash, true)
	assert.Equal(t, "a", agentMainnet.Source)
	assert.Contains(t, agentMainnet.ConnectionID, "0x")
	
	// Test testnet
	agentTestnet := utils.ConstructPhantomAgent(hash, false)
	assert.Equal(t, "b", agentTestnet.Source)
	assert.Contains(t, agentTestnet.ConnectionID, "0x")
}

func TestSigningFlow(t *testing.T) {
	// Generate a test private key
	privateKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	
	// Test signing a USD transfer action with all required fields
	action := map[string]interface{}{
		"destination": "0x1234567890123456789012345678901234567890",
		"amount":      "1000000",
		"time":        uint64(utils.GetTimestampMs()),
	}
	
	signature, err := utils.SignUSDTransferAction(privateKey, action, false)
	require.NoError(t, err)
	assert.NotNil(t, signature)
	assert.NotEmpty(t, signature.R)
	assert.NotEmpty(t, signature.S)
	assert.True(t, signature.V >= 27)
}

func TestActionHash(t *testing.T) {
	action := map[string]interface{}{
		"type":   "order",
		"orders": []interface{}{},
	}
	
	hash, err := utils.ActionHash(action, nil, 12345, nil)
	require.NoError(t, err)
	assert.Len(t, hash, 32) // Keccak256 produces 32-byte hash
}

func TestGetTimestampMs(t *testing.T) {
	timestamp := utils.GetTimestampMs()
	assert.Greater(t, timestamp, int64(0))
	assert.Greater(t, timestamp, int64(1600000000000)) // Should be after 2020
}

func TestOrderWiresToOrderAction(t *testing.T) {
	orderWires := []utils.OrderWire{
		{
			A: 0,
			B: true,
			P: "50000",
			S: "1.5",
			R: false,
			T: utils.OrderTypeWire{
				Limit: &utils.LimitOrderType{TIF: utils.TIFGtc},
			},
		},
	}
	
	builder := "test-builder"
	action := utils.OrderWiresToOrderAction(orderWires, &builder)
	
	assert.Equal(t, "order", action["type"])
	assert.Equal(t, "na", action["grouping"])
	assert.Equal(t, "test-builder", action["builder"])
	assert.Len(t, action["orders"], 1)
}

// Benchmark tests
func BenchmarkFloatToWire(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = utils.FloatToWire(123.456789)
	}
}

func BenchmarkActionHash(b *testing.B) {
	action := map[string]interface{}{
		"type":   "order",
		"orders": []interface{}{},
	}
	
	for i := 0; i < b.N; i++ {
		_, _ = utils.ActionHash(action, nil, 12345, nil)
	}
}

func BenchmarkSignUSDTransferAction(b *testing.B) {
	privateKey, _ := crypto.GenerateKey()
	action := map[string]interface{}{
		"destination": "0x1234567890123456789012345678901234567890",
		"amount":      "1000000",
		"time":        utils.GetTimestampMs(),
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = utils.SignUSDTransferAction(privateKey, action, false)
	}
}
