package peer

import (
	"net"

	"github.com/Seriyin/go-bitaites/db"
	"github.com/dgraph-io/badger"
	"github.com/go-mangos/mangos"
)

type Subscribed struct {
	list map[string]net.Addr
}

type Subbing struct {
	Subscribed
	sockets []mangos.Socket
}

var sublist *Subscribed

func init()  {
	dwrapper := db.GetDBWrapper()
	var err error
	sublist, err = dwrapper.GetSubList([]byte("sublist"))
	switch err {
	case nil:
	case badger.ErrKeyNotFound:
		sublist = &Subscribed{
			make(map[string]net.Addr, 50),
		}
	default:
		panic(err)
	}
}

func (s *Subscribed) SubList() (map[string]net.Addr) {
	return s.list
}