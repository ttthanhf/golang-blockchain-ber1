package consensus

import (
	"bytes"
	"fmt"
	"go-blockchain-ber1/pkg/blockchain"
	"go-blockchain-ber1/pkg/p2p/pb"
	"go-blockchain-ber1/pkg/storage"
	"go-blockchain-ber1/pkg/util"
	"go-blockchain-ber1/pkg/wallet"
	"log/slog"
	"sync"
)

type Consensus struct {
	mu sync.Mutex

	voters              map[string]bool
	totalNodeValidators int

	proposalBlock *pb.Block

	blockDB *storage.BlockDB
}

func NewConsensus(blockDB *storage.BlockDB) *Consensus {
	slog.Info("Init Consensus success")
	return &Consensus{
		totalNodeValidators: 3,
		voters:              make(map[string]bool),
		blockDB:             blockDB,
	}
}

func (c *Consensus) HandleVote(vote *pb.AVote) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	//Threshold
	threshold := (2 * (c.totalNodeValidators / 3))

	slog.Debug("Trigger Consensus Handle Vote")
	slog.Debug("Info Vote : ", "voters", c.voters, "votes", vote, "threshold", threshold)

	// Add vote
	voteUnique := fmt.Sprintf("%s|%d", vote.NodeId, vote.BlockHeight)
	c.voters[voteUnique] = vote.Approve

	// Leader alway have approve Vote so that mean init count is 1
	approveVoteCount := 1
	for _, vote := range c.voters {
		if vote {
			approveVoteCount++
		}
	}

	if approveVoteCount >= threshold {
		c.voters = make(map[string]bool)
		return true
	}

	return false
}

func (c *Consensus) HandleProposeBlock(block *pb.Block, latestBlock *blockchain.Block) (bool, error) {
	// Check Previous Block Hash
	if !bytes.Equal(latestBlock.CurrentBlockHash, block.PreviousBlockHash) {
		slog.Info("Check Fail In: Check Previous Block Hash")
		slog.Debug("Debug Check Prevous Block Hash : ", "block previous hash", string(block.PreviousBlockHash), "latest block hash", string(latestBlock.CurrentBlockHash))

		return false, nil
	}

	// Check Merkle Root
	bcBlock := util.ConvertToBlockchainBlock(block)

	var txHashes [][]byte
	for _, tx := range bcBlock.Transactions {
		txHashes = append(txHashes, tx.Hash())
	}
	merkleRootHash := blockchain.BuildMerkleRoot(txHashes)

	if !bytes.Equal(merkleRootHash, block.MerkleRootHash) {
		slog.Info("Check Fail In: Check Merkle Root")
		return false, nil
	}

	// Check current block hash
	if !bytes.Equal(bcBlock.Hash(), block.CurrentBlockHash) {
		slog.Info("Check Fail In: Check current block hash")
		return false, nil
	}

	// Check block height
	if block.Height != latestBlock.Height+1 {
		slog.Info("Check Fail In: Check block height")
		return false, nil
	}

	//Check Transactions
	for _, tx := range block.Transactions {
		bcTx := util.ConvertToBlockchainTransaction(tx)
		publicKey, _ := util.DecodePublicKey(string(tx.PublicKey))

		if !wallet.VerifyTransaction(bcTx, publicKey) {
			slog.Info("Check Fail In: Check Transactions")
			return false, nil
		}
	}

	c.SetProposalBlock(block)

	return true, nil
}

func (c *Consensus) HandleCommitBlock() error {
	if c.proposalBlock == nil {
		return nil
	}

	bcBlock := util.ConvertToBlockchainBlock(c.proposalBlock)
	if err := c.blockDB.SaveBlock(bcBlock); err != nil {
		return err
	}

	c.RemoveProposalBlock()

	return nil
}

// Proposal block
func (c *Consensus) GetProposalBlock() *pb.Block {
	return c.proposalBlock
}

func (c *Consensus) SetProposalBlock(block *pb.Block) {
	slog.Info("Store Proposal Block")
	slog.Debug("Set proposal block", "block", block)

	c.proposalBlock = block
}

func (c *Consensus) RemoveProposalBlock() {
	slog.Info("Remove proposal block")

	c.proposalBlock = nil
}
