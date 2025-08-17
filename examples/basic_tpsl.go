package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/hyperliquid-go/hyperliquid-go/hyperliquid"
	"github.com/hyperliquid-go/hyperliquid-go/hyperliquid/utils"
)

func RunBasicTPSL() {
	// Parse command line arguments
	isBuy := flag.Bool("is_buy", false, "Whether to place a buy order")
	flag.Parse()

	// Setup clients
	_, _, exchange, err := Setup(utils.TestnetAPIURL, true)
	if err != nil {
		log.Fatal("Setup failed:", err)
	}

	ctx := context.Background()

	// Place an order that should execute by setting the price very aggressively
	var price float64
	if *isBuy {
		price = 2500
	} else {
		price = 1500
	}

	order := utils.Order{
		Asset:      1, // ETH
		IsBuy:      *isBuy,
		LimitPx:    price,
		Sz:         0.02,
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
		log.Fatal("Failed to place order:", err)
	}

	fmt.Printf("Order result: %+v\n", orderResult)

	// Place a stop loss order
	var stopTriggerPx, stopPrice float64
	if *isBuy {
		stopTriggerPx = 1600
		stopPrice = 1500
	} else {
		stopTriggerPx = 2400
		stopPrice = 2500
	}

	stopOrder := utils.Order{
		Asset:      1, // ETH
		IsBuy:      !*isBuy, // Opposite side to close position
		LimitPx:    stopPrice,
		Sz:         0.02,
		ReduceOnly: true,
		Cloid:      nil,
	}

	stopOrderType := utils.OrderType{
		Trigger: &utils.TriggerOrderType{
			TriggerPx: stopTriggerPx,
			IsMarket:  true,
			TPSL:      utils.TPSLSl, // Stop Loss
		},
	}

	stopResult, err := exchange.Order(ctx, []utils.Order{stopOrder}, utils.OrderGroupingNa, &stopOrderType)
	if err != nil {
		log.Printf("Failed to place stop order: %v", err)
	} else {
		fmt.Printf("Stop order result: %+v\n", stopResult)

		// Cancel the stop order
		if stopResult.Status == "ok" && len(stopResult.Response.Data.Statuses) > 0 {
			status := stopResult.Response.Data.Statuses[0]
			if status.Resting != nil {
				cancelResult, err := exchange.Cancel(ctx, []utils.CancelRequest{
					{
						Asset: 1, // ETH
						Oid:   status.Resting.Oid,
					},
				})
				if err != nil {
					log.Printf("Failed to cancel stop order: %v", err)
				} else {
					fmt.Printf("Cancel stop order result: %+v\n", cancelResult)
				}
			}
		}
	}

	// Place a take profit order
	var tpTriggerPx, tpPrice float64
	if *isBuy {
		tpTriggerPx = 1600
		tpPrice = 2500
	} else {
		tpTriggerPx = 2400
		tpPrice = 1500
	}

	tpOrder := utils.Order{
		Asset:      1, // ETH
		IsBuy:      !*isBuy, // Opposite side to close position
		LimitPx:    tpPrice,
		Sz:         0.02,
		ReduceOnly: true,
		Cloid:      nil,
	}

	tpOrderType := utils.OrderType{
		Trigger: &utils.TriggerOrderType{
			TriggerPx: tpTriggerPx,
			IsMarket:  true,
			TPSL:      utils.TPSLTp, // Take Profit
		},
	}

	tpResult, err := exchange.Order(ctx, []utils.Order{tpOrder}, utils.OrderGroupingNa, &tpOrderType)
	if err != nil {
		log.Printf("Failed to place take profit order: %v", err)
	} else {
		fmt.Printf("Take profit order result: %+v\n", tpResult)

		// Cancel the take profit order
		if tpResult.Status == "ok" && len(tpResult.Response.Data.Statuses) > 0 {
			status := tpResult.Response.Data.Statuses[0]
			if status.Resting != nil {
				cancelResult, err := exchange.Cancel(ctx, []utils.CancelRequest{
					{
						Asset: 1, // ETH
						Oid:   status.Resting.Oid,
					},
				})
				if err != nil {
					log.Printf("Failed to cancel take profit order: %v", err)
				} else {
					fmt.Printf("Cancel take profit order result: %+v\n", cancelResult)
				}
			}
		}
	}
}
