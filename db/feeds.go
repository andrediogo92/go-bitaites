package db

import (
	"bytes"
	"encoding/gob"
	"log"

	"github.com/Seriyin/go-bitaites/peer"
	"github.com/Seriyin/go-bitaites/timeline"
	"github.com/dgraph-io/badger"
)

/**
	Joins a subscribed list and a map of ids to posts being constructed.
 */
type readStruct struct {
	fmap map[peer.ID]*timeline.Posts
	subs *peer.Subscribed
}

/**
	Read a bunch of keys from subscribed list.
 */
func (wrapper DBWrapper) ReadAsFeedMap(subList *peer.Subscribed) (map[peer.ID]*timeline.Posts, error) {
	if wrapper.db != nil {
		feedMap := make(map[peer.ID]*timeline.Posts)
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
