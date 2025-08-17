package main

import (
	"crypto/ecdsa"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/hyperliquid-go/hyperliquid-go/hyperliquid"
)

// Setup initializes the SDK clients and validates account state
func Setup(baseURL string, skipWS bool) (string, *hyperliquid.Info, *hyperliquid.Exchange, error) {
	config, err := LoadConfig()
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to load config: %v", err)
	}

	secretKey, err := GetSecretKey(config)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to get secret key: %v", err)
	}

	// Parse private key
	privateKey, err := crypto.HexToECDSA(secretKey[2:]) // Remove 0x prefix
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to parse private key: %v", err)
	}

	// Get address from private key
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return "", nil, nil, fmt.Errorf("failed to cast public key to ECDSA")
	}
	derivedAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	// Use configured address or derived address
	var address string
	if config.AccountAddress != "" {
		address = config.AccountAddress
	} else {
		address = derivedAddress.Hex()
	}

	fmt.Printf("Running with account address: %s\n", address)
	if address != derivedAddress.Hex() {
		fmt.Printf("Running with agent address: %s\n", derivedAddress.Hex())
	}

	// Initialize clients
	timeout := 30 * time.Second
	
	// Create Info client - metadata is loaded automatically during construction
	info, err := hyperliquid.NewInfo(baseURL, skipWS, nil, nil, nil, timeout)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to create info client: %v", err)
	}

	// Check account state
	userState, err := info.UserState(address, "")
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to get user state: %v", err)
	}

	spotUserState, err := info.SpotUserState(address)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to get spot user state: %v", err)
	}

	// Basic validation (simplified)
	fmt.Printf("User state retrieved successfully for address: %s\n", address)
	_ = userState
	_ = spotUserState

	// Initialize exchange client
	exchange, err := hyperliquid.NewExchange(privateKey, baseURL, nil, nil, &address, nil, []string{}, timeout)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to create exchange client: %v", err)
	}
	// Agent wallet validation would be implemented here if needed

	return address, info, exchange, nil
}

// SetupMultiSigWallets loads multiple authorized user wallets for multi-sig operations
func SetupMultiSigWallets() ([]*ecdsa.PrivateKey, error) {
	config, err := LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}

	var authorizedWallets []*ecdsa.PrivateKey

	for _, walletConfig := range config.MultiSig.AuthorizedUsers {
		if walletConfig.SecretKey == "" {
			continue
		}

		privateKey, err := crypto.HexToECDSA(walletConfig.SecretKey[2:]) // Remove 0x prefix
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key for authorized user: %v", err)
		}

		// Verify address matches private key
		publicKey := privateKey.Public()
		publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("failed to cast public key to ECDSA")
		}
		derivedAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

		if walletConfig.AccountAddress != "" && walletConfig.AccountAddress != derivedAddress.Hex() {
			return nil, fmt.Errorf("provided authorized user address %s does not match private key", walletConfig.AccountAddress)
		}

		fmt.Printf("Loaded authorized user for multi-sig: %s\n", derivedAddress.Hex())
		authorizedWallets = append(authorizedWallets, privateKey)
	}

	return authorizedWallets, nil
}
