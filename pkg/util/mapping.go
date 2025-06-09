package util

import (
	"go-blockchain-ber1/pkg/blockchain"
	"go-blockchain-ber1/pkg/p2p/pb"
)

func ConvertToPbTransaction(tx *blockchain.Transaction) *pb.Transaction {
	return &pb.Transaction{
		Sender:    tx.Sender,
		Receiver:  tx.Receiver,
		Amount:    tx.Amount,
		Timestamp: tx.Timestamp,
		Signature: tx.Signature,
	}
}

func ConvertToBlockchainTransaction(tx *pb.Transaction) *blockchain.Transaction {
	return &blockchain.Transaction{
		Sender:    tx.Sender,
		Receiver:  tx.Receiver,
		Amount:    tx.Amount,
		Timestamp: tx.Timestamp,
		Signature: tx.Signature,
	}
}

func ConvertToBlockchainBlock(block *pb.Block) *blockchain.Block {
	var bcTransactions []*blockchain.Transaction
	for _, tx := range block.Transactions {
		bcTransactions = append(bcTransactions, ConvertToBlockchainTransaction(tx))
	}

	return &blockchain.Block{
		Transactions:      bcTransactions,
		MerkleRootHash:    block.MerkleRootHash,
		PreviousBlockHash: block.PreviousBlockHash,
		CurrentBlockHash:  block.CurrentBlockHash,
		Height:            block.Height,
	}
}

func ConvertToPbBlock(block *blockchain.Block) *pb.Block {
	var pbTransactions []*pb.Transaction
	for _, tx := range block.Transactions {
		pbTransactions = append(pbTransactions, ConvertToPbTransaction(tx))
	}

	return &pb.Block{
		Transactions:      pbTransactions,
		MerkleRootHash:    block.MerkleRootHash,
		PreviousBlockHash: block.PreviousBlockHash,
		CurrentBlockHash:  block.CurrentBlockHash,
		Height:            block.Height,
	}
}
