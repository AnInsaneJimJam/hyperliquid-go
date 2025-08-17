package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run . basic_order")
		os.Exit(1)
	}

	exampleName := os.Args[1]
	
	switch exampleName {
	case "basic_order":
		RunBasicOrder()
	default:
		fmt.Printf("Unknown example: %s\n", exampleName)
		os.Exit(1)
	}
}
