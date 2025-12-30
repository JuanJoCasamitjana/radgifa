package server

import (
	"fmt"
	"log"
	"os"
	"time"

	badger "github.com/dgraph-io/badger/v4"
)

const (
	defaultPath = "./tmp"
)

var (
	kvstoragePath = getKVStoragePath()
)

type kvmanager struct {
	db     *badger.DB
	ticker *time.Ticker
	done   chan struct{}
}

type KVManager interface {
	// Define methods for the KVManager interface here
	Close() error
	InsertWithTTL(key, value []byte, ttlSeconds int64) error
	InsertPersistent(key, value []byte) error
	Get(key []byte) ([]byte, error)
	Delete(key []byte) error
}

func NewKVManager() KVManager {
	tmpStat, err := os.Stat(kvstoragePath)
	if err != nil || (tmpStat != nil && !tmpStat.IsDir()) {
		if err := os.MkdirAll(kvstoragePath, 0755); err != nil {
			log.Fatalf("Failed to create KV storage directory at %s: %v", kvstoragePath, err)
		}
	}
	badgerkvstoragePath := fmt.Sprintf("%s/badger", kvstoragePath)

	db, err := badger.Open(badger.
		DefaultOptions(badgerkvstoragePath).
		WithLogger(nil).
		WithMemTableSize(64 << 20).
		WithNumMemtables(3).
		WithSyncWrites(false))
	if err != nil {
		log.Fatal(err)
	}

	kvm := &kvmanager{
		db:     db,
		ticker: time.NewTicker(5 * time.Minute),
		done:   make(chan struct{}),
	}

	// Start garbage collection routine
	go kvm.runGC()

	return kvm
}

func getKVStoragePath() string {
	path := os.Getenv("KV_STORAGE_PATH")
	if path == "" {
		return defaultPath
	}
	return path
}

func (kvm *kvmanager) Close() error {
	// Stop the GC ticker and goroutine
	kvm.ticker.Stop()
	close(kvm.done)

	return kvm.db.Close()
}

// runGC runs the BadgerDB garbage collector periodically
func (kvm *kvmanager) runGC() {
	for {
		select {
		case <-kvm.ticker.C:
			// Run GC with 0.5 discard ratio (collect if 50% or more can be reclaimed)
			err := kvm.db.RunValueLogGC(0.5)
			if err != nil && err != badger.ErrNoRewrite {
				log.Printf("BadgerDB GC error: %v", err)
			}
		case <-kvm.done:
			return
		}
	}
}

func (kvm *kvmanager) InsertWithTTL(key, value []byte, ttlSeconds int64) error {
	err := kvm.db.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry(key, value).WithTTL(time.Duration(ttlSeconds) * time.Second)
		return txn.SetEntry(e)
	})
	return err
}

func (kvm *kvmanager) InsertPersistent(key, value []byte) error {
	err := kvm.db.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry(key, value)
		return txn.SetEntry(e)
	})
	return err
}

func (kvm *kvmanager) Get(key []byte) ([]byte, error) {
	var valCopy []byte
	err := kvm.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		val, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		valCopy = val
		return nil
	})
	if err != nil {
		return nil, err
	}
	return valCopy, nil
}

func (kvm *kvmanager) Delete(key []byte) error {
	err := kvm.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
	return err
}
