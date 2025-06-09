package storage

import (
	"log"
	"log/slog"

	"github.com/syndtr/goleveldb/leveldb"
)

func NewLevelDB(path string) *leveldb.DB {
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		log.Fatalf("LevelDB failed: %v", err)
	}

	slog.Info("Init LevelDB success")

	return db
}
