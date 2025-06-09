package cli

import (
	"go-blockchain-ber1/pkg/p2p/pb"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const leaderAddress = "localhost:50051"

var nodes = []string{leaderAddress, "localhost:50052", "localhost:50053"}
var nodeOrder = []string{"node1", "node2", "node3"}
var addressNodeMap map[string]string = make(map[string]string)

var infoNodes map[string]string = make(map[string]string)
var mu sync.Mutex

func GetClient(targetNode string) (pb.BlockchainClient, error) {
	conn, err := grpc.NewClient(targetNode, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	client := pb.NewBlockchainClient(conn)

	return client, nil
}
