package db

import (
	"bytes"

	"github.com/Seriyin/go-bitaites/timeline"
	"github.com/dgraph-io/badger"
)

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