package main

import (
	"context"
	"fmt"
	"log"

	"github.com/hyperliquid-go/hyperliquid-go/hyperliquid"
	"github.com/hyperliquid-go/hyperliquid-go/hyperliquid/utils"
)

func RunBasicVault() {
	// Setup clients
	_, _, exchange, err := Setup(utils.TestnetAPIURL, true)
	if err != nil {
		log.Fatal("Setup failed:", err)
	}

	ctx := context.Background()

	// Change this address to a vault that you lead or a subaccount that you own
	vault := "0x1719884eb866cb12b2287399b15f7db5e7d775ea"

	// Create exchange client for vault trading
	vaultExchange := hyperliquid.NewExchange(exchange.GetPrivateKey(), true) // testnet
	vaultExchange.SetVaultAddress(vault)

	// Place an order that should rest by setting the price very low
	order := utils.Order{
		Asset:      1, // ETH
		IsBuy:      true,
		LimitPx:    1100.0,
		Sz:         0.2,
		ReduceOnly: false,
		Cloid:      nil,
	}

	orderType := utils.OrderType{
		Limit: &utils.LimitOrderType{
			Tif: utils.TifGtc,
		},
	}

	orderResult, err := vaultExchange.Order(ctx, []utils.Order{order}, utils.OrderGroupingNa, &orderType)
	if err != nil {
		log.Fatal("Failed to place vault order:", err)
	}

	fmt.Printf("Vault order result: %+v\n", orderResult)

	// Cancel the order
	if orderResult.Status == "ok" && len(orderResult.Response.Data.Statuses) > 0 {
		status := orderResult.Response.Data.Statuses[0]
		if status.Resting != nil {
			cancelResult, err := vaultExchange.Cancel(ctx, []utils.CancelRequest{
				{
					Asset: 1, // ETH
					Oid:   status.Resting.Oid,
				},
			})
			if err != nil {
				log.Printf("Failed to cancel vault order: %v", err)
			} else {
				fmt.Printf("Cancel vault order result: %+v\n", cancelResult)
			}
		}
	}
}
