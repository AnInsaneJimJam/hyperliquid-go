package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hyperliquid-go/hyperliquid-go/hyperliquid"
	"github.com/hyperliquid-go/hyperliquid-go/hyperliquid/utils"
)

func RunBasicMarketOrder() {
	// Setup clients
	_, _, exchange, err := Setup(utils.TestnetAPIURL, true)
	if err != nil {
		log.Fatal("Setup failed:", err)
	}

	ctx := context.Background()

	coin := "ETH"
	isBuy := false
	sz := 0.05

	fmt.Printf("We try to Market %s %.3f %s.\n", map[bool]string{true: "Buy", false: "Sell"}[isBuy], sz, coin)

	// Place market order to open position
	orderResult, err := exchange.MarketOpen(ctx, coin, isBuy, sz, 0.01)
	if err != nil {
		log.Fatal("Failed to place market order:", err)
	}

	if orderResult.Status == "ok" {
		for _, status := range orderResult.Response.Data.Statuses {
			if status.Filled != nil {
				filled := status.Filled
				fmt.Printf("Order #%d filled %s @%s\n", filled.Oid, filled.TotalSz, filled.AvgPx)
			} else if status.Error != "" {
				fmt.Printf("Error: %s\n", status.Error)
			}
		}

		fmt.Println("We wait for 2s before closing")
		time.Sleep(2 * time.Second)

		fmt.Printf("We try to Market Close all %s.\n", coin)
		closeResult, err := exchange.MarketClose(ctx, coin)
		if err != nil {
			log.Printf("Failed to close position: %v", err)
		} else if closeResult.Status == "ok" {
			for _, status := range closeResult.Response.Data.Statuses {
				if status.Filled != nil {
					filled := status.Filled
					fmt.Printf("Order #%d filled %s @%s\n", filled.Oid, filled.TotalSz, filled.AvgPx)
				} else if status.Error != "" {
					fmt.Printf("Error: %s\n", status.Error)
				}
			}
		}
	}
}
