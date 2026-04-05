package main

import (
	"fmt"
	"os"

	"network-probe/internal/cli"
)

func main() {
	if err := cli.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
