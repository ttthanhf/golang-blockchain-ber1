package p2p

import (
	"bytes"
	"context"
	"fmt"
	"go-blockchain-ber1/pkg/blockchain"
	"go-blockchain-ber1/pkg/consensus"
	"go-blockchain-ber1/pkg/p2p/pb"
	"go-blockchain-ber1/pkg/storage"
	"go-blockchain-ber1/pkg/util"
	"go-blockchain-ber1/pkg/wallet"
	"log"
	"log/slog"
	"net"
	"time"

	"google.golang.org/grpc"
)

type NodeStatus string

var (
	IDLE                          NodeStatus = "IDLE"
	SYNCING                       NodeStatus = "SYNCING"
	VALIDATING_BLOCK              NodeStatus = "VALIDATING_BLOCK"
	VERIFYING_TRANSACTION         NodeStatus = "VERIFYING_TRANSACTION"
	FORWARD_TRANSACTION_TO_LEADER NodeStatus = "FORWARD_TRANSACTION_TO_LEADER"
	WAITING_NEXT_BLOCK            NodeStatus = "WAITING_NEXT_BLOCK"
	SENT_VOTE_TO_LEADER           NodeStatus = "SENT_VOTE_TO_LEADER"
	PROCESSING_VOTE               NodeStatus = "PROCESSING_VOTE"
	COMMIT_BLOCK                  NodeStatus = "COMMIT_BLOCK"
)

type grpcServer struct {
	pb.UnimplementedBlockchainServer
	blockDB     *storage.BlockDB
	memPool     *blockchain.MemPool
	consensus   *consensus.Consensus
	peerManager *PeerManager

	isLeader   bool
	nodeId     string
	nodeStatus NodeStatus
}

func (s *grpcServer) SendTransaction(ctx context.Context, tx *pb.Transaction) (*pb.Empty, error) {
	bcTx := util.ConvertToBlockchainTransaction(tx)

	if !s.isLeader {
		s.nodeStatus = FORWARD_TRANSACTION_TO_LEADER

		slog.Warn("This node is node leader for handle send transaction service ! Forward to leader")
		if err := s.peerManager.SendTransactionToLeader(ctx, tx); err != nil {
			return nil, err
		}

		return nil, nil
	}
	slog.Info("Trigger Sent Transaction")

	// Verify Transaction
	s.nodeStatus = VERIFYING_TRANSACTION

	publicKey, err := util.DecodePublicKey(string(tx.PublicKey))
	if err != nil {
		return nil, err
	}

	if !wallet.VerifyTransaction(bcTx, publicKey) {
		return nil, fmt.Errorf("transaction cant verify")
	}

	// store transaction in pending
	s.memPool.AddPendingTransaction(tx)

	s.nodeStatus = WAITING_NEXT_BLOCK

	slog.Info("Transaction added in mempool")

	return nil, nil
}

func (s *grpcServer) GetBlock(ctx context.Context, blockHeight *pb.BlockHeight) (*pb.Block, error) {
	block, err := s.blockDB.GetBlock(blockHeight.Height)
	if err != nil {
		return nil, err
	}

	pbBlock := util.ConvertToPbBlock(block)

	return pbBlock, nil
}

func (s *grpcServer) GetLatestBlock(ctx context.Context, _ *pb.Empty) (*pb.Block, error) {
	block, err := s.blockDB.GetLatestBlock()
	if err != nil {
		return nil, err
	}

	pbBlock := util.ConvertToPbBlock(block)

	return pbBlock, nil
}

func (s *grpcServer) ProposeBlock(ctx context.Context, block *pb.Block) (*pb.Empty, error) {
	slog.Info("Trigger Propose Block")

	latestBlock, err := s.blockDB.GetLatestBlock()
	if err != nil {
		return nil, err
	}

	//
	s.nodeStatus = VALIDATING_BLOCK

	isAprrove, err := s.consensus.HandleProposeBlock(block, latestBlock)
	if err != nil {
		return nil, err
	}

	s.nodeStatus = SENT_VOTE_TO_LEADER

	s.peerManager.SendVoteToLeader(ctx, &pb.AVote{
		Approve:     isAprrove,
		NodeId:      s.nodeId,
		BlockHeight: block.Height,
	})

	return nil, nil
}

func (s *grpcServer) Vote(ctx context.Context, vote *pb.AVote) (*pb.Empty, error) {
	if !s.isLeader {
		slog.Warn("This node is not leader for trigger grpc server vote service")
		return nil, nil
	}
	slog.Info("Trigger Leader Recevice Vote")

	s.nodeStatus = PROCESSING_VOTE

	if s.consensus.GetProposalBlock() == nil {
		s.nodeStatus = IDLE
		return nil, nil
	}

	if s.consensus.HandleVote(vote) {
		// Commit block
		s.CommitBlock(context.Background(), nil)

		s.peerManager.BroastCastCommitBlock()

		// Clear mem pool
		s.memPool.ClearAllPendingTransaction()
	}

	return nil, nil
}

func (s *grpcServer) CommitBlock(ctx context.Context, _ *pb.Empty) (*pb.Empty, error) {
	slog.Info("Trigger Commit Block")

	if !s.isLeader && s.consensus.GetProposalBlock() == nil {
		s.nodeStatus = SYNCING

		if err := s.syncWithLeader(); err != nil {
			return nil, err
		}
	} else {
		s.nodeStatus = COMMIT_BLOCK

		if err := s.consensus.HandleCommitBlock(); err != nil {
			return nil, err
		}
	}

	s.nodeStatus = IDLE

	slog.Info("Flow DONE")

	return nil, nil
}

func (s *grpcServer) StreamNodeInfo(_ *pb.Empty, stream pb.Blockchain_StreamNodeInfoServer) error {
	slog.Info("Trigger Steam Node Info: On")
	timer := time.NewTicker(1 * time.Nanosecond)
	defer timer.Stop()

	oldStatus := ""

	for {
		select {
		case <-stream.Context().Done():
			slog.Info("Trigger Steam Node Info: Off")
			return nil
		case <-timer.C:
			newStatus := string(s.nodeStatus)
			if oldStatus != newStatus {

				slog.Debug("Change : ", "oldStatus", oldStatus, "newStatus", newStatus)

				response := pb.SteamNodeInfoResponse{
					NodeId:     s.nodeId,
					NodeStatus: newStatus,
				}
				oldStatus = response.NodeStatus

				if err := stream.Send(&response); err != nil {
					return err
				}
			}
		}
	}
}

func NewGRPCServer(db *storage.BlockDB, pm *PeerManager, memPool *blockchain.MemPool, consensus *consensus.Consensus, isLeader bool, nodeId string) *grpcServer {
	return &grpcServer{
		blockDB:     db,
		peerManager: pm,
		memPool:     memPool,
		consensus:   consensus,
		isLeader:    isLeader,
		nodeId:      nodeId,
		nodeStatus:  IDLE,
	}
}

func (g *grpcServer) Init(addressPort string) {
	slog.Info("GRPC server Init")

	serviceRegister := grpc.NewServer()
	pb.RegisterBlockchainServer(serviceRegister, g)

	listener, err := net.Listen("tcp", addressPort)
	if err != nil {
		log.Fatalf("tcp listener failed: %v", err)
	}

	if err := serviceRegister.Serve(listener); err != nil {
		log.Fatalf("failed to serve gRPC: %v", err)
	}
}

func (s *grpcServer) syncWithLeader() error {
	slog.Info("Syncing block with leader")

	leaderLatestBlock, err := s.peerManager.GetLatestBlockFromLeader()
	if err != nil {
		slog.Error("Fail to get lastest block from leader", "err", err)
		return err
	}

	latestBlock, err := s.blockDB.GetLatestBlock()
	if err != nil {
		slog.Error("Fail to get lastest block", "err", err)
		return err
	}

	if leaderLatestBlock.Height == latestBlock.Height {
		return nil
	}

	for height := latestBlock.Height + 1; height <= leaderLatestBlock.Height; height++ {
		pbLeaderBlock, err := s.peerManager.GetBlockFromLeader(height)
		if err != nil {
			slog.Error("Failed to get block from leader", "height", height, "err", err)
			return err
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
			return err
		}

		// Check Previous Hash
		if !bytes.Equal(bcLeaderBlock.CurrentBlockHash, bcLeaderBlock.Hash()) {
			slog.Info("Block hash not match")
			return err
		}

		// Save Block
		if err := s.blockDB.SaveBlock(bcLeaderBlock); err != nil {
			slog.Error(fmt.Sprintf("Recovery faild - Cant not save block: %v; Error: %v", bcLeaderBlock, err))
			return err
		}
	}

	slog.Info("Sync successfully In Commit Block")

	return nil
}
