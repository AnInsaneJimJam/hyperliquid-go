package main

import (
	"context"
	"fmt"
	"log"

	"github.com/hyperliquid-go/hyperliquid-go/hyperliquid"
	"github.com/hyperliquid-go/hyperliquid-go/hyperliquid/utils"
)

func RunBasicOrderModify() {
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

	// Modify the order by oid
	if orderResult.Status == "ok" && len(orderResult.Response.Data.Statuses) > 0 {
		status := orderResult.Response.Data.Statuses[0]
		if status.Resting != nil {
			oid := status.Resting.Oid
			
			// Query order status first
			orderStatus, err := info.QueryOrderByOid(ctx, address, oid)
			if err != nil {
				log.Printf("Failed to query order by oid: %v", err)
			} else {
				fmt.Printf("Order status by oid: %+v\n", orderStatus)
			}

			// Modify the order - change size to 0.1 and price to 1105
			modifiedOrder := utils.Order{
				Asset:      1, // ETH
				IsBuy:      true,
				LimitPx:    1105.0,
				Sz:         0.1,
				ReduceOnly: false,
				Cloid:      &cloid,
			}

			modifyResult, err := exchange.ModifyOrder(ctx, []utils.ModifyRequest{
				{
					Oid:   oid,
					Order: modifiedOrder,
				},
			})
			if err != nil {
				log.Printf("Failed to modify order by oid: %v", err)
			} else {
				fmt.Printf("Modify result with oid: %+v\n", modifyResult)
			}

			// Modify the order again using cloid (if supported)
			// Note: Some exchanges support modifying by client order ID
			modifiedOrder2 := utils.Order{
				Asset:      1, // ETH
				IsBuy:      true,
				LimitPx:    1110.0,
				Sz:         0.05,
				ReduceOnly: false,
				Cloid:      &cloid,
			}

			modifyResult2, err := exchange.ModifyOrder(ctx, []utils.ModifyRequest{
				{
					Oid:   oid, // Still need oid for modification
					Order: modifiedOrder2,
				},
			})
			if err != nil {
				log.Printf("Failed to modify order by cloid: %v", err)
			} else {
				fmt.Printf("Modify result with cloid: %+v\n", modifyResult2)
			}

			// Cancel the order after modifications
			cancelResult, err := exchange.Cancel(ctx, []utils.CancelRequest{
				{
					Asset: 1, // ETH
					Oid:   oid,
				},
			})
			if err != nil {
				log.Printf("Failed to cancel order: %v", err)
			} else {
				fmt.Printf("Cancel result: %+v\n", cancelResult)
			}
		}
	}
}
