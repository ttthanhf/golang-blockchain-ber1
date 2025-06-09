package cli

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"go-blockchain-ber1/pkg/p2p/pb"
	"go-blockchain-ber1/pkg/util"
	"log"
	"os"
	"slices"
	"strings"
)

type TransactionView struct {
	Sender    string  `json:"sender"`
	Receiver  string  `json:"receiver"`
	Amount    float64 `json:"amount"`
	Timestamp int64   `json:"timestamp"`
	Signature string  `json:"signature"`
}

type BlockView struct {
	MerkleRootHash    string            `json:"merkle_root_hash"`
	PreviousBlockHash string            `json:"previous_block_hash"`
	CurrentBlockHash  string            `json:"current_block_hash"`
	Height            uint64            `json:"height"`
	Transactions      []TransactionView `json:"transactions"`
}

func GetCurrentBlockHeightCLI() {
	getCurrentBlockHeightCmd := flag.NewFlagSet("get-current-block-height", flag.ExitOnError)
	node := getCurrentBlockHeightCmd.String("node", leaderAddress, "Input node target")

	getCurrentBlockHeightCmd.Parse(os.Args[2:])

	isNodeExist := slices.Contains(nodes, *node)
	if !isNodeExist {
		log.Fatalf("invalid node '%s'; allowed nodes: %v", *node, nodes)
	}
	fmt.Printf("Connect node `%s`\n", *node)

	client, err := GetClient(*node)
	if err != nil {
		log.Fatalf("Error: Cant connect node: %s", *node)
	}

	block, err := client.GetLatestBlock(context.Background(), nil)
	if err != nil {
		log.Fatalf("Error: Send Transaction Failed: %v", err)
	}

	fmt.Printf("Current Block Height: %d\n", block.Height)
}

func GetBlockCLI() {
	getBlockCmd := flag.NewFlagSet("get-block", flag.ExitOnError)
	blockHeight := getBlockCmd.Uint64("block-height", 0, "Input block height")
	node := getBlockCmd.String("node", leaderAddress, "Input node target")

	getBlockCmd.Parse(os.Args[2:])

	isNodeExist := slices.Contains(nodes, *node)
	if !isNodeExist {
		log.Fatalf("invalid node '%s'; allowed nodes: %v", *node, nodes)
	}
	fmt.Printf("Connect node `%s`\n", *node)

	if *blockHeight <= 0 {
		log.Fatalf("Error: blockHeight is required and larger than 0")
	}

	client, err := GetClient(*node)
	if err != nil {
		log.Fatalf("Error: Cant connect node: %s", *node)
	}

	block, err := client.GetBlock(context.Background(), &pb.BlockHeight{
		Height: *blockHeight,
	})
	if err != nil {
		if strings.Contains(err.Error(), "leveldb: not found") {
			fmt.Println("\033[1;31mBlock not found\033[0m")
			return
		}

		log.Fatalf("Error: Send Transaction Failed: %v", err)
	}

	transactionViews := []TransactionView{}
	for _, tx := range block.Transactions {
		txView := TransactionView{
			Sender:    string(tx.Sender),
			Receiver:  string(tx.Receiver),
			Amount:    tx.Amount,
			Timestamp: tx.Timestamp,
			Signature: util.Base58Encode(tx.Signature),
		}
		transactionViews = append(transactionViews, txView)
	}

	blockView := BlockView{
		MerkleRootHash:    util.Base58Encode(block.MerkleRootHash),
		PreviousBlockHash: util.Base58Encode(block.PreviousBlockHash),
		CurrentBlockHash:  util.Base58Encode(block.CurrentBlockHash),
		Height:            block.Height,
		Transactions:      transactionViews,
	}

	out, _ := json.MarshalIndent(blockView, "", "  ")
	fmt.Println(string(out))
}
