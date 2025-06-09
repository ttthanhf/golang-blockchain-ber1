package node

import (
	"bytes"
	"fmt"
	"go-blockchain-ber1/pkg/blockchain"
	"go-blockchain-ber1/pkg/consensus"
	"go-blockchain-ber1/pkg/p2p"
	"go-blockchain-ber1/pkg/p2p/pb"
	"go-blockchain-ber1/pkg/storage"
	"go-blockchain-ber1/pkg/util"
	"log/slog"
	"time"
)

type Node struct {
	peerManager *p2p.PeerManager
	blockDB     *storage.BlockDB
	memPool     *blockchain.MemPool
	consensus   *consensus.Consensus

	IsLeader bool
	NodeId   string
}

func NewNode(peerManager *p2p.PeerManager, blockDB *storage.BlockDB, mempool *blockchain.MemPool, consensus *consensus.Consensus, isLeader bool, nodeId string) *Node {
	return &Node{
		peerManager: peerManager,
		blockDB:     blockDB,
		memPool:     mempool,
		consensus:   consensus,

		IsLeader: isLeader,
		NodeId:   nodeId,
	}
}

func (n *Node) Init() {
	slog.Info("Init Node success")

	n.recovery()
	go n.taskQueue()
}

func (n *Node) recovery() {
	if n.IsLeader {
		slog.Debug("This node is leader")
		return
	}
	slog.Info("Checking sync with leader node")

	leaderLatestBlock, err := n.peerManager.GetLatestBlockFromLeader()
	if err != nil {
		slog.Error("Fail to get lastest block from leader", "err", err)
		return
	}

	latestBlock, err := n.blockDB.GetLatestBlock()
	if err != nil {
		slog.Error("Fail to get lastest block", "err", err)
		return
	}

	if leaderLatestBlock.Height == latestBlock.Height {
		slog.Info("Latest Block with Leader")
		return
	}

	slog.Warn("Not Latest Block With Leader ! Syncing...")
	for height := latestBlock.Height + 1; height <= leaderLatestBlock.Height; height++ {
		pbLeaderBlock, err := n.peerManager.GetBlockFromLeader(height)
		if err != nil {
			slog.Error("Failed to get block from leader", "height", height, "err", err)
			return
		}
		bcLeaderBlock := util.ConvertToBlockchainBlock(pbLeaderBlock)

		// Validate block
		// Check merkle root
		var txHashes [][]byte
		for _, tx := range bcLeaderBlock.Transactions {
			txHashes = append(txHashes, tx.Hash())
		}
		merkleRoot := blockchain.BuildMerkleRoot(txHashes)

		if !bytes.Equal(merkleRoot, pbLeaderBlock.MerkleRootHash) {
			slog.Info("Merkle Root not match")
			return
		}

		// Check Previous Hash
		if !bytes.Equal(bcLeaderBlock.CurrentBlockHash, bcLeaderBlock.Hash()) {
			slog.Info("Block hash not match")
			return
		}

		// Save Block
		if err := n.blockDB.SaveBlock(bcLeaderBlock); err != nil {
			slog.Error(fmt.Sprintf("Recovery faild - Cant not save block: %v; Error: %v", bcLeaderBlock, err))
			return
		}
	}

	slog.Info("Sync successfully with leader node")
}

func (n *Node) createNewBlock() *pb.Block {
	var bcTransactions []*blockchain.Transaction
	pendingTransactions := n.memPool.GetAllPendingTransactions()
	for _, tx := range pendingTransactions {
		bcTx := util.ConvertToBlockchainTransaction(tx)
		bcTransactions = append(bcTransactions, bcTx)
	}

	latestBlock, err := n.blockDB.GetLatestBlock()
	if err != nil {
		slog.Error("Cant get latest block", "err", err)
		return nil
	}
	block := blockchain.NewBlock(bcTransactions, latestBlock)

	pbBlock := util.ConvertToPbBlock(block)
	pbBlock.Transactions = pendingTransactions

	return pbBlock

}

func (n *Node) taskQueue() {
	if !n.IsLeader {
		return
	}
	slog.Info("Running Task Queue")

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case t := <-ticker.C:
			pendingTransactions := n.memPool.GetAllPendingTransactions()
			if len(pendingTransactions) > 0 {
				slog.Info("Task Queue Create Block: Creating new block", "t", t)
				block := n.createNewBlock()

				n.consensus.SetProposalBlock(block)
				n.peerManager.BroastCastProposeBlock(block)
			} else {
				slog.Debug("Task Queue Create Block: No Transaction Found", "t", t)
			}
		}
	}
}
