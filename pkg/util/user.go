package util

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"go-blockchain-ber1/pkg/types"
	"os"
)

func FindUserByAddress(address string) (*types.UserData, error) {
	filePath := "wallet.json"
	var data []types.UserData
	if IsFileExist(filePath) {
		bytes, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(bytes, &data); err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("database not found")
	}

	for _, wallet := range data {
		if wallet.Address == address {
			return &wallet, nil
		}
	}

	return nil, fmt.Errorf("user not found")
}

func GetPublicKeyByAddress(address string) (*ecdsa.PublicKey, error) {
	sender, err := FindUserByAddress(address)
	if err != nil {
		return nil, err
	}

	return DecodePublicKey(sender.PublicKey)
}

func GetPrivatekeyByAddress(address string) (*ecdsa.PrivateKey, error) {
	sender, err := FindUserByAddress(address)
	if err != nil {
		return nil, err
	}

	privKey, err := DecodePrivateKey(sender.PrivateKey)
	if err != nil {
		return nil, err
	}
	return privKey, nil
}
