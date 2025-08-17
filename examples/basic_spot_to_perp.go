package main

import (
	"context"
	"fmt"
	"log"

	"github.com/hyperliquid-go/hyperliquid-go/hyperliquid"
	"github.com/hyperliquid-go/hyperliquid-go/hyperliquid/utils"
)

func RunBasicSpotToPerp() {
	// Setup clients
	_, _, exchange, err := Setup(utils.TestnetAPIURL, true)
	if err != nil {
		log.Fatal("Setup failed:", err)
	}

	ctx := context.Background()

	// Transfer 1.23 USDC from perp wallet to spot wallet
	transferResult, err := exchange.USDClassTransfer(ctx, 1.23, false) // false = to spot
	if err != nil {
		log.Printf("Failed to transfer from perp to spot: %v", err)
	} else {
		fmt.Printf("Transfer from perp to spot result: %+v\n", transferResult)
	}

	// Transfer 1.23 USDC from spot wallet to perp wallet
	transferResult, err = exchange.USDClassTransfer(ctx, 1.23, true) // true = to perp
	if err != nil {
		log.Printf("Failed to transfer from spot to perp: %v", err)
	} else {
		fmt.Printf("Transfer from spot to perp result: %+v\n", transferResult)
	}
}
