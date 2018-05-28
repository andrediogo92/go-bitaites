package peer

import (
	"net"

	"github.com/Seriyin/go-bitaites/db"
	"github.com/dgraph-io/badger"
	"github.com/go-mangos/mangos"
)

type Subscribed struct {
	list map[ID]net.Addr
}

type Subbing struct {
	Subscribed
	subs []mangos.Socket
	directs []mangos.Socket
}

var sublist *Subscribed

func init()  {
	dwrapper := db.GetDBWrapper()
	var err error
	sublist, err = dwrapper.GetSubList()
	switch err {
	case nil:
	case badger.ErrKeyNotFound:
		sublist = &Subscribed{
			make(map[ID]net.Addr, 50),
		}
	default:
		panic(err)
	}
}

func (s *Subscribed) SubList() (map[ID]net.Addr) {
	return s.list
}