package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/hyperliquid-go/hyperliquid-go/hyperliquid/utils"
)

func RunBasicOrder() {
	// Setup clients
	address, info, exchange, err := Setup(utils.TestnetAPIURL, true)
	if err != nil {
		log.Fatal("Setup failed:", err)
	}

	// Get the user state and print out position information
	userState, err := info.UserState(address, "")
	if err != nil {
		log.Fatal("Failed to get user state:", err)
	}

	// Parse user state response
	if stateMap, ok := userState.(map[string]interface{}); ok {
		if assetPositions, ok := stateMap["assetPositions"].([]interface{}); ok {
			hasPositions := false
			for _, pos := range assetPositions {
				if posMap, ok := pos.(map[string]interface{}); ok {
					if position, ok := posMap["position"].(map[string]interface{}); ok {
						if szi, ok := position["szi"].(string); ok && szi != "0" {
							if !hasPositions {
								fmt.Println("positions:")
								hasPositions = true
							}
							positionJSON, _ := json.MarshalIndent(position, "", "  ")
							fmt.Println(string(positionJSON))
						}
					}
				}
			}
			if !hasPositions {
				fmt.Println("no open positions")
			}
		} else {
			fmt.Println("no open positions")
		}
	} else {
		fmt.Println("no open positions")
	}

	// Place an order that should rest by setting the price very low
	orderType := utils.OrderType{
		Limit: &utils.LimitOrderType{},
	}

	orderResult, err := exchange.Order("ETH", true, 0.2, 1100.0, orderType, false, nil, nil)
	if err != nil {
		log.Fatal("Failed to place order:", err)
	}

	fmt.Printf("Order result: %+v\n", orderResult)

	// Parse order result and query/cancel if successful
	if result, ok := orderResult.(map[string]interface{}); ok {
		if status, ok := result["status"].(string); ok && status == "ok" {
			if response, ok := result["response"].(map[string]interface{}); ok {
				if data, ok := response["data"].(map[string]interface{}); ok {
					if statuses, ok := data["statuses"].([]interface{}); ok && len(statuses) > 0 {
						if statusMap, ok := statuses[0].(map[string]interface{}); ok {
							if resting, ok := statusMap["resting"].(map[string]interface{}); ok {
								if oidFloat, ok := resting["oid"].(float64); ok {
									oid := int(oidFloat)
									fmt.Printf("Order placed successfully with OID: %d\n", oid)
									
									// Query the order status by oid
									orderStatus, err := info.QueryOrderByOID(address, oid)
									if err != nil {
										log.Printf("Failed to query order by oid: %v", err)
									} else {
										fmt.Printf("Order status by oid: %+v\n", orderStatus)
									}

									// Cancel the order
									cancelResult, err := exchange.Cancel("ETH", oid)
									if err != nil {
										log.Printf("Failed to cancel order: %v", err)
									} else {
										fmt.Printf("Cancel result: %+v\n", cancelResult)
									}
								}
							}
						}
					}
				}
			}
		} else {
			fmt.Printf("Order failed: %+v\n", result)
		}
	} else {
		fmt.Printf("Unexpected order result format: %+v\n", orderResult)
	}
}
