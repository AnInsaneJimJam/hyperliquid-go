package main

import (
	"context"
	"fmt"
	"log"

	"github.com/hyperliquid-go/hyperliquid-go/hyperliquid"
	"github.com/hyperliquid-go/hyperliquid-go/hyperliquid/utils"
)

func RunBasicOrderWithCloid() {
	// Setup clients
	address, info, exchange, err := Setup(utils.TestnetAPIURL, true)
	if err != nil {
		log.Fatal("Setup failed:", err)
	}

	ctx := context.Background()

	// Create client order ID
	cloid := "0x00000000000000000000000000000001"
	
	// Place an order that should rest by setting the price very low
	order := utils.Order{
		Asset:      1, // ETH
		IsBuy:      true,
		LimitPx:    1100.0,
		Sz:         0.2,
		ReduceOnly: false,
		Cloid:      &cloid,
	}

	orderType := utils.OrderType{
		Limit: &utils.LimitOrderType{
			Tif: utils.TifGtc,
		},
	}

	orderResult, err := exchange.Order(ctx, []utils.Order{order}, utils.OrderGroupingNa, &orderType)
	if err != nil {
		log.Fatal("Failed to place order:", err)
	}

	fmt.Printf("Order result: %+v\n", orderResult)

	// Query the order status by cloid
	orderStatus, err := info.QueryOrderByCloid(ctx, address, cloid)
	if err != nil {
		log.Printf("Failed to query order by cloid: %v", err)
	} else {
		fmt.Printf("Order status by cloid: %+v\n", orderStatus)
	}

	// Non-existent cloid example
	invalidCloid := "0x00000000000000000000000000000002"
	orderStatus, err = info.QueryOrderByCloid(ctx, address, invalidCloid)
	if err != nil {
		log.Printf("Failed to query order by invalid cloid: %v", err)
	} else {
		fmt.Printf("Order status by invalid cloid: %+v\n", orderStatus)
	}

	// Cancel the order by cloid
	if orderResult.Status == "ok" && len(orderResult.Response.Data.Statuses) > 0 {
		status := orderResult.Response.Data.Statuses[0]
		if status.Resting != nil {
			cancelResult, err := exchange.CancelByCloid(ctx, []utils.CancelByCloidRequest{
				{
					Asset: 1, // ETH
					Cloid: cloid,
				},
			})
			if err != nil {
				log.Printf("Failed to cancel order by cloid: %v", err)
			} else {
				fmt.Printf("Cancel result: %+v\n", cancelResult)
			}
		}
	}
}
