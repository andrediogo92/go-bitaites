package db

import (
"bytes"
"encoding/gob"
"log"
"math/rand"

"github.com/dgraph-io/badger"
"github.com/pkg/errors"

)


type DBWrapper struct {
	db *badger.DB
	seq *badger.Sequence
}

type readClosure func(txn *badger.Txn, key []byte) error
type writeClosure func(txn *badger.Txn) error
type iteratorClosure func(it *badger.Iterator) error

var (
	db                 DBWrapper
	ErrUninitializedDB = errors.New("Uninitialized DB")
)


func init() {
	// Open the Badger database located in the ./feeds directory.
	// It will be created if it doesn't exist.
	opts := badger.DefaultOptions
	opts.Dir = "./feeds"
	opts.ValueDir = "./feeds"
	bdb, err := badger.Open(opts)
	if err != nil {
		log.Fatal(err)
	} else {
		seq, err2 := bdb.GetSequence([]byte(opts.Dir), 1000)
		if err == nil {
			db = DBWrapper{bdb, seq}
		} else {
			log.Fatal(err2)
			db = DBWrapper{ bdb, nil}
		}
	}
}


func decodeValue(bs *bytes.Buffer, bdg *badger.Txn, key []byte) (decoder *gob.Decoder, err error) {
	it, err := bdg.Get(key)
	decoder = gob.NewDecoder(bs)
	if err == nil {
		bts, err := it.ValueCopy(nil)
		if err == nil {
			bs.Write(bts)
		}
	}
	return
}

func (wrapper DBWrapper) readDB(cl readClosure, key []byte) error {
	return wrapper.db.View(func(txn *badger.Txn) error {
		defer txn.Discard()
		return cl(txn, key)
	})
}

func (wrapper DBWrapper) writeDB(cl writeClosure) error {
	return wrapper.db.Update(func(txn *badger.Txn) error {
		defer txn.Discard()
		return cl(txn)
	})
}

func (wrapper DBWrapper) iteratorDB(cl iteratorClosure) error {
	return wrapper.db.View(func(txn *badger.Txn) error {
		defer txn.Discard()
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 100
		it := txn.NewIterator(opts)
		defer it.Close()
		return cl(it)
	})
}

func (wrapper DBWrapper) keyIteratorDB(cl iteratorClosure) error {
	return wrapper.db.View(func(txn *badger.Txn) error {
		defer txn.Discard()
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		return cl(it)
	})
}


func (wrapper DBWrapper) GetNewId() (id uint64) {
	if wrapper.seq != nil {
		var err error
		id, err = wrapper.seq.Next()
		if err != nil {
			id = rand.Uint64()
		}
	} else {
		id = rand.Uint64()
	}
	return
}





func GetDBWrapper() (*DBWrapper) {
	return &db
}

