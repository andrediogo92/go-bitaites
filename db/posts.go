package db

import (
	"bytes"
	"encoding/gob"
	"time"

	"github.com/Seriyin/go-bitaites/timeline"
	"github.com/dgraph-io/badger"
)

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

