package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hyperliquid-go/hyperliquid-go/hyperliquid"
	"github.com/hyperliquid-go/hyperliquid-go/hyperliquid/utils"
)

func RunBasicLeverage() {
	// Setup clients
	address, info, exchange, err := Setup(utils.TestnetAPIURL, true)
	if err != nil {
		log.Fatal("Setup failed:", err)
	}

	ctx := context.Background()

	// Get the user state and print out leverage information for ETH
	userState, err := info.ClearinghouseState(ctx, address)
	if err != nil {
		log.Fatal("Failed to get user state:", err)
	}

	for _, assetPosition := range userState.AssetPositions {
		if assetPosition.Position.Coin == 1 { // ETH asset ID
			leverageJSON, _ := json.MarshalIndent(assetPosition.Position.Leverage, "", "  ")
			fmt.Printf("Current leverage for ETH: %s\n", string(leverageJSON))
		}
	}

	// Set the ETH leverage to 21x (cross margin)
	leverageResult, err := exchange.UpdateLeverage(ctx, 1, true, 21) // ETH, cross margin, 21x
	if err != nil {
		log.Printf("Failed to update leverage (cross): %v", err)
	} else {
		fmt.Printf("Update leverage (cross) result: %+v\n", leverageResult)
	}

	// Set the ETH leverage to 21x (isolated margin)
	leverageResult, err = exchange.UpdateLeverage(ctx, 1, false, 21) // ETH, isolated margin, 21x
	if err != nil {
		log.Printf("Failed to update leverage (isolated): %v", err)
	} else {
		fmt.Printf("Update leverage (isolated) result: %+v\n", leverageResult)
	}

	// Add 1 dollar of extra margin to the ETH position
	marginResult, err := exchange.UpdateIsolatedMargin(ctx, 1, 1.0) // ETH, +$1
	if err != nil {
		log.Printf("Failed to update isolated margin: %v", err)
	} else {
		fmt.Printf("Update isolated margin result: %+v\n", marginResult)
	}

	// Get the user state and print out the final leverage information after our changes
	userState, err = info.ClearinghouseState(ctx, address)
	if err != nil {
		log.Fatal("Failed to get updated user state:", err)
	}

	for _, assetPosition := range userState.AssetPositions {
		if assetPosition.Position.Coin == 1 { // ETH asset ID
			leverageJSON, _ := json.MarshalIndent(assetPosition.Position.Leverage, "", "  ")
			fmt.Printf("Final leverage for ETH: %s\n", string(leverageJSON))
		}
	}
}
