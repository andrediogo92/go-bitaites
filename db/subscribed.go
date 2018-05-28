package db

import (
	"bytes"

	"github.com/Seriyin/go-bitaites/peer"
	"github.com/dgraph-io/badger"
)

func (wrapper *DBWrapper) GetSubList() (sublist *peer.Subscribed, err error) {
	if wrapper == nil {
		b := []byte("sublist")
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
