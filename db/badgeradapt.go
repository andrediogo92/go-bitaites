package db

import (
"bytes"
"encoding/gob"
"log"
"math/rand"
	"time"

	"../timeline"
"github.com/Seriyin/go-bitaites/peer"
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

func addPostToMerge(merge *badger.MergeOperator, post *timeline.Post) {
	bs := bytes.Buffer{}
	encoder := gob.NewEncoder(&bs)
	encoder.Encode(post)
	merge.Add(bs.Bytes())
}


func mergeNewPost(existing []byte, new []byte) []byte {
	var posts timeline.Posts
	bs := bytes.Buffer{}
	decoder := gob.NewDecoder(&bs)
	bs.Write(existing)
	decoder.Decode(posts)
	posts.Hashes() = append(posts.Hashes(), string(new))
	bs.Truncate(0)
	encoder := gob.NewEncoder(&bs)
	encoder.Encode(posts)
	return bs.Bytes()
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

/**
	Write a post to the DB and merge its hash
	into the user's own post feed.
 */
func (wrapper DBWrapper) WriteToTimeline(new *timeline.Post) error {
	if wrapper.db == nil {
		merge := wrapper.db.GetMergeOperator([]byte(
			timeline.GetPosts().Id()),
			mergeNewPost,
			200 * time.Millisecond)
		addPostToMerge(merge, new)
		merge.Stop()
		return wrapper.writeDB(func (b *badger.Txn) (err error) {
			bytes, err := new.AsBinary()
			if err == nil {
				b.Set(new.Key(), bytes)
			}
			return
		})
	} else {
		return ErrUninitializedDB
	}
}


func (wrapper DBWrapper) DumpToSubscribed(key []byte, new []*timeline.Post) error {
	if wrapper.db == nil {
		merge := wrapper.db.GetMergeOperator(key, mergeNewPost, 10 * time.Millisecond)
		for _, v := range new {
			addPostToMerge(merge, v)
		}
		merge.Stop()
		return wrapper.writeDB(func(txn *badger.Txn) error {
			var err error
			for _, v := range new {
				b, err := v.AsBinary()
				if err == nil {
					txn.Set(v.Key(), b)
				}
			}
			return err
		})
	} else {
		return ErrUninitializedDB
	}
}

/**

 */

type readStruct struct {
	fmap map[string]*timeline.Posts
	subs *peer.Subscribed
}

/**
	Read a bunch of keys from subscribed list.
 */
func (wrapper DBWrapper) ReadAsFeedMap(subList *peer.Subscribed) (map[string]*timeline.Posts, error) {
	if wrapper.db != nil {
		feedMap := make(map[string]*timeline.Posts)
		rs := &readStruct{
			feedMap,
			subList,
		}
		err := wrapper.keyIteratorDB(func (b *badger.Iterator) (err error) {
			var posts *timeline.Posts
			bs := bytes.Buffer{}
			decoder := gob.NewDecoder(&bs)
			var cp []byte
			for k := range rs.subs.SubList() {
				b.Seek([]byte(k))
				it := b.Item()
				if it == nil {
					log.Print(badger.ErrRetry)
				} else {
					cp = nil
					cp, err = it.ValueCopy(cp)
					if err == nil {
						log.Print(err)
					} else {
						bs.Write(cp)
						decoder.Decode(posts)
						rs.fmap[posts.Id()] = posts
						posts = nil
					}
				}
			}
			return
		})
		return feedMap, err
	} else {
		return nil, ErrUninitializedDB
	}
}


func (wrapper DBWrapper) PurgeFeed(feedKey string, posts *timeline.Posts) (err error) {
	if wrapper.db != nil {
		wr := func(b *badger.Txn) (err error) {
			err = b.Delete([]byte(feedKey))
			if err == nil {
				for _, v := range posts.Hashes() {
					//Can't guarantee keys are deleted properly
					err2 := b.Delete([]byte (v))
					if err2 != nil {
						log.Print(err2)
					}
				}
			}
			return
		}
		return wrapper.writeDB(wr)
	} else {
		return ErrUninitializedDB
	}
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

func (wrapper *DBWrapper) GetSubList(b []byte) (sublist *peer.Subscribed, err error) {
	if wrapper == nil {
		db.readDB(func(txn *badger.Txn, key []byte) error {
			bs := &bytes.Buffer{}
			decoder, err := decodeValue(bs, txn, key)
			if err == nil {
				err = decoder.Decode(sublist)
			}
			return err
		}, b)
		return
	} else {
		return nil, ErrUninitializedDB
	}
}

func (wrapper *DBWrapper) GetMyKeyPair(b []byte) (id *timeline.OwnId, err error) {
	if wrapper == nil {
		db.readDB(func (bdg *badger.Txn, key []byte) (err error)  {
			bs := &bytes.Buffer{}
			decoder, err := decodeValue(bs, bdg, key)
			if err == nil {
				err = decoder.Decode(id)
			}
			return
		}, b)
		return
	} else {
		return nil, ErrUninitializedDB
	}
}

func (wrapper *DBWrapper) GetOwnPosts(b []byte) (posts *timeline.Posts, err error) {
	if wrapper == nil {
		db.readDB(func (bdg *badger.Txn, key []byte) (err error)  {
			bs := &bytes.Buffer{}
			decoder, err := decodeValue(bs, bdg, key)
			if err == nil {
				err = decoder.Decode(posts)
			}
			return
		}, b)
		return
	} else {
		return nil, ErrUninitializedDB
	}
}

func (wrapper *DBWrapper) GetMyUser(b []byte) (user string, err error) {
	if wrapper == nil {
		var userPtr *[]byte
		db.readDB(func (bdg *badger.Txn, key []byte) (err error)  {
			it, err := bdg.Get(key)
			if err == nil {
				var bts []byte = nil
				bts, err = it.ValueCopy(bts)
				if err == nil {
					*userPtr = bts
				}
			}
			return
		}, b)
		user = string(*userPtr)
		return
	} else {
		return "", ErrUninitializedDB
	}

}


func GetDBWrapper() (*DBWrapper) {
	return &db
}

