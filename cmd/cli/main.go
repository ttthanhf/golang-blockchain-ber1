package main

import (
	"fmt"
	"go-blockchain-ber1/cmd/cli/cli"
	"os"
)

func main() {
	switch os.Args[1] {
	case "create-user":
		cli.CreateUserCLI()
	case "send-transaction":
		cli.SendTransactionCLI()
	case "get-block":
		cli.GetBlockCLI()
	case "get-current-block-height":
		cli.GetCurrentBlockHeightCLI()
	case "monitor-node":
		cli.MonitorNodesCLI()
	default:
		fmt.Println("Unknown CLI")
	}
}
