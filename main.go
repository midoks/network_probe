package main

import (
	"fmt"
	"os"

	"network-probe/internal/cli"
)

func main() {
	// fmt.Println("Network Probe starting...")
	if err := cli.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
