package blockchain

import (
	"crypto/sha256"
)

func merkleHash(firstHash []byte, secondHash []byte) []byte {
	data := append(firstHash, secondHash...)
	hash := sha256.Sum256(data)
	return hash[:]
}

func BuildMerkleRoot(txHashes [][]byte) []byte {
	txHashesLen := len(txHashes)

	if txHashesLen == 0 {
		return nil
	}

	// Merkle Root
	if txHashesLen == 1 {
		return txHashes[0]
	}

	// If txHashes lenght is odd
	if txHashesLen%2 != 0 {
		// Copy the last hash
		lastTxHash := txHashes[txHashesLen-1]
		txHashes = append(txHashes, lastTxHash)
		txHashesLen = len(txHashes)
	}

	var newTxHashes [][]byte
	for i := 0; i < txHashesLen; i += 2 {
		hash := merkleHash(txHashes[i], txHashes[i+1])
		newTxHashes = append(newTxHashes, hash)
	}

	return BuildMerkleRoot(newTxHashes)
}
