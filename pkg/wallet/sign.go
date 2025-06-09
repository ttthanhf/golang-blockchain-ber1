package wallet

import (
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"go-blockchain-ber1/pkg/blockchain"
	"math/big"
)

func VerifyTransaction(tx *blockchain.Transaction, pubKey *ecdsa.PublicKey) bool {
	txHash := tx.Hash()
	// Assume signature is r and s concatenated, parse them back to big.Int
	r := new(big.Int).SetBytes(tx.Signature[:len(tx.Signature)/2])
	s := new(big.Int).SetBytes(tx.Signature[len(tx.Signature)/2:])
	return ecdsa.Verify(pubKey, txHash, r, s)
}

func SignTransaction(tx *blockchain.Transaction, privKey *ecdsa.PrivateKey) error {
	txHash := tx.Hash()
	r, s, err := ecdsa.Sign(rand.Reader, privKey, txHash)
	if err != nil {
		return fmt.Errorf("failed to sign transaction: %w", err)
	}
	// Store R and S as a concatenated byte slice
	tx.Signature = append(r.Bytes(), s.Bytes()...)
	return nil
}
