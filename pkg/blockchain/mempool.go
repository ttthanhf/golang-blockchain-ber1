package blockchain

import (
	"go-blockchain-ber1/pkg/p2p/pb"
	"log/slog"
	"sync"
)

type MemPool struct {
	mu                  sync.Mutex
	pendingTransactions []*pb.Transaction
}

func NewMemPool() *MemPool {
	slog.Info("Init mem pool success")

	return &MemPool{
		pendingTransactions: make([]*pb.Transaction, 0),
	}
}

func (m *MemPool) AddPendingTransaction(tx *pb.Transaction) {
	m.mu.Lock()
	defer m.mu.Unlock()

	slog.Debug("Add transaction in mempool", "tx", tx)

	m.pendingTransactions = append(m.pendingTransactions, tx)
}

func (m *MemPool) GetAllPendingTransactions() []*pb.Transaction {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.pendingTransactions
}

func (m *MemPool) ClearAllPendingTransaction() {
	m.mu.Lock()
	defer m.mu.Unlock()

	slog.Info("Remove all pending transactions in mempool")

	m.pendingTransactions = []*pb.Transaction{}
}
