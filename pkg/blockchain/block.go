package blockchain

import (
	"crypto/sha256"
	"encoding/json"
)

type Block struct {
	Transactions      []*Transaction
	MerkleRootHash    []byte
	PreviousBlockHash []byte
	CurrentBlockHash  []byte
	Height            uint64
}

func NewBlock(transactions []*Transaction, latestBlock *Block) *Block {
	var txHashes [][]byte
	for _, tx := range transactions {
		txHashes = append(txHashes, tx.Hash())
	}
	merkleHash := BuildMerkleRoot(txHashes)

	block := &Block{
		Transactions:      transactions,
		MerkleRootHash:    merkleHash,
		PreviousBlockHash: latestBlock.CurrentBlockHash,
		Height:            latestBlock.Height + 1,
	}
	block.CurrentBlockHash = block.Hash()

	return block
}

func (b *Block) Hash() []byte {
	blockCopy := *b
	blockCopy.CurrentBlockHash = nil
	data, _ := json.Marshal(blockCopy)
	hash := sha256.Sum256(data)
	return hash[:]
}
