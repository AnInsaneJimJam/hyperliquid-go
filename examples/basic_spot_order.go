package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hyperliquid-go/hyperliquid-go/hyperliquid"
	"github.com/hyperliquid-go/hyperliquid-go/hyperliquid/utils"
)

const (
	PURR           = "PURR/USDC"
	OTHER_COIN     = "@8"
	OTHER_COIN_NAME = "KORILA/USDC"
)

func RunBasicSpotOrder() {
	// Setup clients
	address, info, exchange, err := Setup(utils.TestnetAPIURL, true)
	if err != nil {
		log.Fatal("Setup failed:", err)
	}

	ctx := context.Background()

	// Get the user state and print out spot balance information
	spotUserState, err := info.SpotClearinghouseState(ctx, address)
	if err != nil {
		log.Fatal("Failed to get spot user state:", err)
	}

	if len(spotUserState.Balances) > 0 {
		fmt.Println("spot balances:")
		for _, balance := range spotUserState.Balances {
			balanceJSON, _ := json.MarshalIndent(balance, "", "  ")
			fmt.Println(string(balanceJSON))
		}
	} else {
		fmt.Println("no available token balances")
	}

	// Get PURR asset ID from metadata
	meta, err := info.Meta(ctx)
	if err != nil {
		log.Fatal("Failed to get metadata:", err)
	}

	var purrAssetID int
	for i, asset := range meta.SpotUniverse {
		if asset.Name == PURR {
			purrAssetID = i
			break
		}
	}

	// Place an order that should rest by setting the price very low
	order := utils.Order{
		Asset:      purrAssetID, // PURR/USDC
		IsBuy:      true,
		LimitPx:    0.5,
		Sz:         24.0,
		ReduceOnly: false,
		Cloid:      nil,
	}

	orderType := utils.OrderType{
		Limit: &utils.LimitOrderType{
			Tif: utils.TifGtc,
		},
	}

	orderResult, err := exchange.Order(ctx, []utils.Order{order}, utils.OrderGroupingNa, &orderType)
	if err != nil {
		log.Fatal("Failed to place PURR order:", err)
	}

	fmt.Printf("PURR order result: %+v\n", orderResult)

	// Query the order status by oid
	if orderResult.Status == "ok" && len(orderResult.Response.Data.Statuses) > 0 {
		status := orderResult.Response.Data.Statuses[0]
		if status.Resting != nil {
			oid := status.Resting.Oid
			orderStatus, err := info.QueryOrderByOid(ctx, address, oid)
			if err != nil {
				log.Printf("Failed to query order by oid: %v", err)
			} else {
				fmt.Printf("Order status by oid: %+v\n", orderStatus)
			}

			// Cancel the order
			cancelResult, err := exchange.Cancel(ctx, []utils.CancelRequest{
				{
					Asset: purrAssetID,
					Oid:   oid,
				},
			})
			if err != nil {
				log.Printf("Failed to cancel PURR order: %v", err)
			} else {
				fmt.Printf("Cancel PURR order result: %+v\n", cancelResult)
			}
		}
	}

	// For other spot assets other than PURR/USDC use @{index} e.g. on testnet @8 is KORILA/USDC
	var otherAssetID int
	for i, asset := range meta.SpotUniverse {
		if asset.Name == OTHER_COIN_NAME {
			otherAssetID = i
			break
		}
	}

	otherOrder := utils.Order{
		Asset:      otherAssetID, // KORILA/USDC
		IsBuy:      true,
		LimitPx:    12.0,
		Sz:         1.0,
		ReduceOnly: false,
		Cloid:      nil,
	}

	otherOrderResult, err := exchange.Order(ctx, []utils.Order{otherOrder}, utils.OrderGroupingNa, &orderType)
	if err != nil {
		log.Printf("Failed to place other coin order: %v", err)
	} else {
		fmt.Printf("Other coin order result: %+v\n", otherOrderResult)

		if otherOrderResult.Status == "ok" && len(otherOrderResult.Response.Data.Statuses) > 0 {
			status := otherOrderResult.Response.Data.Statuses[0]
			if status.Resting != nil {
				// The SDK now also supports using spot names, although be careful as they might not always be unique
				cancelResult, err := exchange.Cancel(ctx, []utils.CancelRequest{
					{
						Asset: otherAssetID,
						Oid:   status.Resting.Oid,
					},
				})
				if err != nil {
					log.Printf("Failed to cancel other coin order: %v", err)
				} else {
					fmt.Printf("Cancel other coin order result: %+v\n", cancelResult)
				}
			}
		}
	}
}
