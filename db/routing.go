package db

import (
	"bytes"
	"encoding/gob"

	"github.com/Seriyin/go-bitaites/pex"
	"github.com/dgraph-io/badger"
)

func (wrapper *DBWrapper) GetAddrManager() (manager *pex.AddrManager, err error) {
	if wrapper != nil {
		b := []byte("addr-manager")
		db.readDB(func (bdg *badger.Txn, key []byte) (err error)  {
			bs := &bytes.Buffer{}
			decoder, err := decodeValue(bs, bdg, key)
			if err == nil {
				err = decoder.Decode(manager)
			}
			return
		}, b)
		return
	} else {
		return nil, ErrUninitializedDB
	}
}

func (wrapper *DBWrapper) GetRoutingTable() (rout *pex.RoutingTable, err error) {
	if wrapper != nil {
		b := []byte("routing-manager")
		db.readDB(func (bdg *badger.Txn, key []byte) (err error)  {
			bs := &bytes.Buffer{}
			decoder, err := decodeValue(bs, bdg, key)
			if err == nil {
				err = decoder.Decode(rout)
			}
			return
		}, b)
		return
	} else {
		return nil, ErrUninitializedDB
	}
}


func (wrapper *DBWrapper) SetRoutingTable(table *pex.RoutingTable) error {
	if wrapper != nil {
		b := []byte("routing-manager")
		return db.writeDB(func(txn *badger.Txn) error {
			bs := &bytes.Buffer{}
			enc := gob.NewEncoder(bs)
			enc.Encode(table)
			return txn.Set(b, bs.Bytes())
		})
	} else {
		return ErrUninitializedDB
	}
}


func (wrapper *DBWrapper) SetAddrManager(addr *pex.AddrManager) error {
	if wrapper != nil {
		b := []byte("addr-manager")
		return db.writeDB(func(txn *badger.Txn) error {
			bs := &bytes.Buffer{}
			enc := gob.NewEncoder(bs)
			enc.Encode(addr)
			return txn.Set(b, bs.Bytes())
		})
	} else {
		return ErrUninitializedDB
	}
}