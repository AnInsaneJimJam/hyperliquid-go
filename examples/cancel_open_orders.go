package main

import (
	"context"
	"fmt"
	"log"

	"github.com/hyperliquid-go/hyperliquid-go/hyperliquid/utils"
)

func RunCancelOpenOrders() {
	// Setup clients
	address, info, exchange, err := Setup(utils.TestnetAPIURL, true)
	if err != nil {
		log.Fatal("Setup failed:", err)
	}

	ctx := context.Background()

	// Get all open orders
	openOrders, err := info.OpenOrders(ctx, address)
	if err != nil {
		log.Fatal("Failed to get open orders:", err)
	}

	if len(openOrders) == 0 {
		fmt.Println("No open orders to cancel")
		return
	}

	// Cancel each open order
	var cancelRequests []utils.CancelRequest
	for _, openOrder := range openOrders {
		fmt.Printf("Cancelling order: %+v\n", openOrder)
		cancelRequests = append(cancelRequests, utils.CancelRequest{
			Coin: openOrder.Coin,
			OID:  openOrder.Oid,
		})
	}

	// Batch cancel all orders
	if len(cancelRequests) > 0 {
		cancelResult, err := exchange.BulkCancel(cancelRequests)
		if err != nil {
			log.Printf("Failed to cancel orders: %v", err)
		} else {
			fmt.Printf("Cancel result: %+v\n", cancelResult)
			
			// Print individual cancellation statuses
			if cancelResult.Status == "ok" {
				for i, status := range cancelResult.Response.Data.Statuses {
					fmt.Printf("Order %d cancellation status: %s\n", cancelRequests[i].OID, status)
				}
			}
		}
	}
}
