package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"go-blockchain-ber1/pkg/util"
)

func GenerateKeyPair() (*ecdsa.PrivateKey, error) {
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key pair: %w", err)
	}
	return privKey, nil
}

func PublicKeyToAddress(pubKey *ecdsa.PublicKey) []byte {
	pubKeyBytes := append(pubKey.X.Bytes(), pubKey.Y.Bytes()...)
	hash := sha256.Sum256(pubKeyBytes)
	base58Check := util.Base58CheckEncode(hash[:])

	return []byte(base58Check)
}
