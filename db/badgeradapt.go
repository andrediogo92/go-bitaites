package db

import (
	"log"
	"math/rand"

	"../timeline"
	"github.com/dgraph-io/badger"
)

type DBWrapper struct {
	db *badger.DB
	seq *badger.Sequence
}

type readClosure func(txn *badger.Txn, companion interface{}) error
type writeClosure func(txn *badger.Txn, companion interface{}) error
type iteratorClosure func(it *badger.Iterator, companion interface{}) error

func InitDB() (db DBWrapper, err error) {
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
	return
}

func (d DBWrapper) readDB(cl readClosure, prm interface{}) (val interface{}, err error) {
	err = d.db.View(func(txn *badger.Txn) error {
		defer txn.Discard()
		return cl(txn, prm)
	})
	val = prm
	return
}

func (d DBWrapper) writeDB(cl writeClosure, prm interface{}) (err error) {
	err = d.db.Update(func(txn *badger.Txn) error {
		defer txn.Discard()
		return cl(txn, prm)
	})
	return
}

func (d DBWrapper) iteratorDB(cl iteratorClosure, prm interface{}) (val interface{}, err error) {
	err = d.db.View(func(txn *badger.Txn) error {
		defer txn.Discard()
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()
		return cl(it, prm)
	})
	val = prm
	return
}

func (d DBWrapper) keyIteratorDB(cl iteratorClosure, prm interface{}) (val interface{}, err error) {
	err = d.db.View(func(txn *badger.Txn) error {
		defer txn.Discard()
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		cl(it, prm)
		return cl(it, prm)
	})
	val = prm
	return
}

/**
Should be a merger function.
 */
func (d DBWrapper) WriteToTimeline(new *timeline.Post) error {
	wr := func (b *badger.Txn, new interface{}) (err error) {
		post := new.(*timeline.Post)
		bytes, err := post.AsBinary()
		if err == nil {
			b.Set(post.Key(), bytes)
			b.Commit(nil)
		}
		return
	}
	return d.writeDB(wr, new)
}

func (d DBWrapper) WriteTimeline(new *timeline.Timeline) error {
	wr := func (b *badger.Txn, new interface{}) (err error) {
		timeline := new.(*timeline.Timeline)
		bytes, err := timeline.AsBinary()
		if err == nil {
			b.Set(timeline.Key(), bytes)
			b.Commit(nil)
		}
		return
	}
	return d.writeDB(wr, new)
}

func (d DBWrapper) GetNewId() (id uint64) {
	if d.seq != nil {
		var err error
		id, err = d.seq.Next()
		if err != nil {
			id = rand.Uint64()
		}
	} else {
		id = rand.Uint64()
	}
	return
}

