package main

import (
	"context"
	"fmt"
	"log"

	"github.com/hyperliquid-go/hyperliquid-go/hyperliquid"
	"github.com/hyperliquid-go/hyperliquid-go/hyperliquid/utils"
)

func RunBasicTransfer() {
	// Setup clients
	_, _, exchange, err := Setup(utils.TestnetAPIURL, true)
	if err != nil {
		log.Fatal("Setup failed:", err)
	}

	ctx := context.Background()

	// Check if this is an agent wallet
	if exchange.GetAccountAddress() != exchange.GetWalletAddress() {
		log.Fatal("Agents do not have permission to perform internal transfers")
	}

	// Transfer 1 USD to the zero address for demonstration purposes
	transferResult, err := exchange.USDTransfer(ctx, 1.0, "0x0000000000000000000000000000000000000000")
	if err != nil {
		log.Fatal("Failed to transfer USD:", err)
	}

	fmt.Printf("Transfer result: %+v\n", transferResult)
}
