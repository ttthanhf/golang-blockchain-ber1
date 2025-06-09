package p2p

import (
	"context"
	"fmt"
	"go-blockchain-ber1/pkg/p2p/pb"
	"log/slog"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Peer struct {
	Address string

	client pb.BlockchainClient
	conn   *grpc.ClientConn
}

type PeerManager struct {
	peers         map[string]*Peer
	leaderAddress string
}

func NewPeerManager(leaderAddress string) *PeerManager {
	slog.Info("Init peer manager success")
	return &PeerManager{
		peers:         make(map[string]*Peer),
		leaderAddress: leaderAddress,
	}
}

func (pm *PeerManager) AddPeer(address string) error {
	conn, err := grpc.NewClient(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return err
	}

	client := pb.NewBlockchainClient(conn)

	pm.peers[address] = &Peer{
		Address: address,

		conn:   conn,
		client: client,
	}

	slog.Debug(fmt.Sprintf("Added peer: %v", pm.peers[address]))

	return nil
}

func (pm *PeerManager) AddPeers(addresses []string) {
	for _, address := range addresses {
		if err := pm.AddPeer(address); err != nil {
			slog.Error(fmt.Sprintf("cant add peer : %s - Error: %v", address, err))
		}
	}
}

// BroastCast
func (pm *PeerManager) BroastCastProposeBlock(block *pb.Block) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	slog.Info("Trigger Broast cast Propose Block")

	slog.Debug("Propose Block", "block", block, "peers", pm.peers)

	var wg sync.WaitGroup
	for _, peer := range pm.peers {
		wg.Add(1)
		go func(p *Peer) {
			defer wg.Done()
			if _, err := p.client.ProposeBlock(ctx, block); err != nil {
				slog.Error("cant propose block to peer", "err", err, "peer", peer, "client", peer.client)
			}
		}(peer)
	}

	wg.Wait()
}

func (pm *PeerManager) BroastCastCommitBlock() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	var wg sync.WaitGroup
	for _, peer := range pm.peers {
		wg.Add(1)
		go func(p *Peer) {
			defer wg.Done()
			if _, err := peer.client.CommitBlock(ctx, nil); err != nil {
				slog.Error("Commit block failed to peer", "err", err, "peer", peer, "client", peer.client)
			}
		}(peer)
	}
	wg.Wait()
}

// Leader
func (pm *PeerManager) GetLeader() *Peer {
	return pm.peers[pm.leaderAddress]
}

func (pm *PeerManager) GetBlockFromLeader(blockHeight uint64) (*pb.Block, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	block, err := pm.GetLeader().client.GetBlock(ctx, &pb.BlockHeight{Height: blockHeight})
	if err != nil {
		return nil, err
	}

	return block, nil
}

func (pm *PeerManager) GetLatestBlockFromLeader() (*pb.Block, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	block, err := pm.GetLeader().client.GetLatestBlock(ctx, nil)
	if err != nil {
		return nil, err
	}

	return block, nil
}

func (pm *PeerManager) SendTransactionToLeader(ctx context.Context, tx *pb.Transaction) error {
	if _, err := pm.GetLeader().client.SendTransaction(ctx, tx); err != nil {
		slog.Error("Cant not send transaction to leader", "err", err)
		return err
	}

	slog.Info("Sent transaction to leader")

	return nil
}

func (pm *PeerManager) SendVoteToLeader(ctx context.Context, vote *pb.AVote) error {
	if _, err := pm.GetLeader().client.Vote(ctx, vote); err != nil {
		slog.Error("Cant not send vote to leader", "err", err)
		return err
	}

	slog.Info("Sent vote to leader")

	return nil
}
