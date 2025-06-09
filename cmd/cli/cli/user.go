package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"go-blockchain-ber1/pkg/types"
	"go-blockchain-ber1/pkg/util"
	"go-blockchain-ber1/pkg/wallet"
	"log"
	"os"
)

func createUser(name string) error {
	filePath := "wallet.json"
	var data []types.UserData
	if util.IsFileExist(filePath) {
		bytes, _ := os.ReadFile(filePath)
		json.Unmarshal(bytes, &data)
	}
	for _, wallet := range data {
		if wallet.Name == name {
			return fmt.Errorf("user name '%s' already exists", name)
		}
	}

	privKey, _ := wallet.GenerateKeyPair()
	address := wallet.PublicKeyToAddress(&privKey.PublicKey)

	privateKeyEncode := util.Base58CheckEncode(privKey.D.Bytes())
	publicKeyEncode := util.EncodePublicKey(privKey)

	jsonData := types.UserData{
		Name:       name,
		PublicKey:  publicKeyEncode,
		PrivateKey: privateKeyEncode,
		Address:    string(address),
	}
	data = append(data, jsonData)

	jsonBytes, _ := json.MarshalIndent(data, "", "  ")
	os.WriteFile(filePath, jsonBytes, 0644)

	return nil
}

func CreateUserCLI() {
	createUserCmd := flag.NewFlagSet("create-user", flag.ExitOnError)
	name := createUserCmd.String("name", "", "Input user name")
	createUserCmd.Parse(os.Args[2:])

	if *name == "" {
		log.Fatalf("Error: name is required")
	}

	if err := createUser(*name); err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Printf("Created new user with name: %s\n", *name)
}
