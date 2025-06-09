package cli

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"go-blockchain-ber1/pkg/blockchain"
	"go-blockchain-ber1/pkg/types"
	"go-blockchain-ber1/pkg/util"
	"go-blockchain-ber1/pkg/wallet"
	"log"
	"os"
	"slices"
)

func SendTransactionCLI() {
	sendTransactionCmd := flag.NewFlagSet("send-transaction", flag.ExitOnError)
	sender := sendTransactionCmd.String("sender", "", "Input sender address")
	receiver := sendTransactionCmd.String("receiver", "", "Input receiver address")
	amount := sendTransactionCmd.Float64("amount", 0, "Input amount")
	node := sendTransactionCmd.String("node", leaderAddress, "Input node target")

	sendTransactionCmd.Parse(os.Args[2:])

	if *sender == "" {
		log.Fatalf("Error: sender is required")
	}
	if *receiver == "" {
		log.Fatalf("Error: receiver is required")
	}
	if *amount <= 0 {
		log.Fatalf("Error: amount must be greater than 0")
	}

	isNodeExist := slices.Contains(nodes, *node)
	if !isNodeExist {
		log.Fatalf("invalid node '%s'; allowed nodes: %v", *node, nodes)
	}
	fmt.Printf("Connect node `%s`\n", *node)

	filePath := "wallet.json"
	var data []types.UserData
	if util.IsFileExist(filePath) {
		bytes, _ := os.ReadFile(filePath)
		json.Unmarshal(bytes, &data)
	} else {
		log.Fatalf("Error: Database not found")
	}

	senderData, err := util.FindUserByAddress(*sender)
	if err != nil {
		log.Fatalf("Error: Sender not found")
	}

	_, err = util.FindUserByAddress(*receiver)
	if err != nil {
		log.Fatalf("Error: Receiver not found")
	}

	// Create transaction
	privKey, _ := util.DecodePrivateKey(senderData.PrivateKey)
	tx := blockchain.NewTransaction([]byte(*sender), []byte(*receiver), *amount)
	wallet.SignTransaction(tx, privKey)

	publicKey := util.EncodePublicKey(privKey)

	client, err := GetClient(*node)
	if err != nil {
		log.Fatalf("Error: Cant connect node: %s", *node)
	}

	pbTx := util.ConvertToPbTransaction(tx)
	pbTx.PublicKey = []byte(publicKey)

	if _, err := client.SendTransaction(context.Background(), pbTx); err != nil {
		log.Fatalf("Error: Send Transaction Failed: %v", err)
	}
}
