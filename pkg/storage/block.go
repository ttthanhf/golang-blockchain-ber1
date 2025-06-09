package storage

import (
	"encoding/json"
	"go-blockchain-ber1/pkg/blockchain"
	"go-blockchain-ber1/pkg/util"
	"log/slog"
	"strconv"

	"github.com/syndtr/goleveldb/leveldb"
)

type BlockDB struct {
	DB *leveldb.DB
}

func NewBlockDB(db *leveldb.DB) *BlockDB {
	return &BlockDB{
		DB: db,
	}
}

func (b *BlockDB) Init() {
	slog.Info("Init BlockDB success")

	b.CreateGenesisBlock()
}

func (b *BlockDB) CreateGenesisBlock() error {
	iter := b.DB.NewIterator(nil, nil)
	defer iter.Release()

	if err := iter.Error(); err != nil {
		return err
	}

	// If database is empty then create genesis block
	if !iter.Next() {
		block := &blockchain.Block{
			Transactions:      nil,
			MerkleRootHash:    nil,
			PreviousBlockHash: []byte("tran-tan-thanh"),
			Height:            1,
		}
		block.CurrentBlockHash = block.Hash()

		b.SaveBlock(block)

		slog.Info("Created genesis block success")
	}

	return nil
}

func (b *BlockDB) SaveBlock(block *blockchain.Block) error {
	slog.Debug("Save block", "block", *block)
	// Write latest block height
	blockHeight := strconv.Itoa(int(block.Height))
	err := b.DB.Put([]byte("latest_block_height"), []byte(blockHeight), nil)
	if err != nil {
		return err
	}

	data, _ := json.Marshal(block)
	return b.DB.Put([]byte(blockHeight), data, nil)
}

func (b *BlockDB) GetBlock(blockHeight uint64) (*blockchain.Block, error) {
	heightStr := strconv.Itoa(int(blockHeight))
	data, err := b.DB.Get([]byte(heightStr), nil)
	if err != nil {
		return nil, err
	}

	var block blockchain.Block
	if err := json.Unmarshal(data, &block); err != nil {
		return nil, err
	}

	return &block, nil
}

func (b *BlockDB) GetLatestBlock() (*blockchain.Block, error) {
	heighBytes, err := b.DB.Get([]byte("latest_block_height"), nil)
	if err != nil {
		slog.Error("GetLastestBlock Faild", "err", err)
		return nil, err
	}

	heightInt := util.BytesToInt(heighBytes)

	block, err := b.GetBlock(uint64(heightInt))
	if err != nil {
		slog.Error("GetLastestBlock Faild - GetBlock Faild", "err", err)
		return nil, err
	}

	return block, nil
}

func (b *BlockDB) GetlatestHeight() (int, error) {
	block, err := b.GetLatestBlock()
	if err != nil {
		return 0, err
	}

	return int(block.Height), nil
}
