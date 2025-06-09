package blockchain

import (
	"crypto/sha256"
	"encoding/json"
	"time"
)

type Transaction struct {
	Sender    []byte // Public Key or Address
	Receiver  []byte // Public Key or Address
	Amount    float64
	Timestamp int64
	Signature []byte // R and S concatenated
}

func NewTransaction(sender []byte, receiver []byte, amount float64) *Transaction {
	tx := &Transaction{
		Sender:    sender,
		Receiver:  receiver,
		Amount:    amount,
		Timestamp: time.Now().Unix(),
	}

	return tx
}

func (t *Transaction) Hash() []byte {
	// Create a hashable representation of the transaction
	txCopy := *t
	txCopy.Signature = nil // Exclude signature from hash
	data, _ := json.Marshal(txCopy)
	hash := sha256.Sum256(data)
	return hash[:]
}
