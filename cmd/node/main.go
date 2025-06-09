package main

import (
	"go-blockchain-ber1/pkg/blockchain"
	"go-blockchain-ber1/pkg/config"
	"go-blockchain-ber1/pkg/consensus"
	"go-blockchain-ber1/pkg/node"
	"go-blockchain-ber1/pkg/p2p"
	"go-blockchain-ber1/pkg/storage"
	"log/slog"
	"os"
	"strings"
)

func main() {
	const addressPort = ":50051"

	// Get Eviroment
	nodeId := os.Getenv("NODE_ID")
	leaderAddress := os.Getenv("LEADER") + addressPort
	isLeader := strings.Split(leaderAddress, ":")[0] == nodeId
	peers := strings.Split(os.Getenv("PEERS"), ",")
	isLevelDebug := os.Getenv("LEVEL_DEBUG") == "true"

	// Config
	config.Logger(isLevelDebug)

	// START
	slog.Info("===============START=================")

	// Init Database
	db := storage.NewLevelDB("data")
	defer db.Close()

	// Init Block Database
	blockDB := storage.NewBlockDB(db)
	blockDB.Init()

	// Init Peer Manager
	peerManager := p2p.NewPeerManager(leaderAddress)
	peerManager.AddPeers(peers)

	//
	memPool := blockchain.NewMemPool()
	consensus := consensus.NewConsensus(blockDB)

	// Init Node
	node := node.NewNode(peerManager, blockDB, memPool, consensus, isLeader, nodeId)
	node.Init()

	// Init grpc server
	server := p2p.NewGRPCServer(blockDB, peerManager, memPool, consensus, isLeader, nodeId)
	server.Init(addressPort)

	// === END === //
}
