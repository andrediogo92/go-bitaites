package pex

import (
	"time"

	"github.com/Seriyin/go-bitaites/db"
	"github.com/Seriyin/go-bitaites/peer"
	"github.com/Seriyin/go-bitaites/timeline"
	"github.com/dgraph-io/badger"
)

var dht *KadDHT

const (
	DefaultBucketSize = 8
	DefaultLatencyToleranceMillis = 1000
	NumBootstrapQueries = 5
	//Parallelism factor for lookups.
	Alpha = 3
)

type KadDHT struct {
	self peer.ID
	*RoutingTable
	*AddrManager
	birth time.Time

	quit chan struct{}
}

func init() {
	dbwrapper := db.GetDBWrapper()
	mgr, err := dbwrapper.GetAddrManager()
	switch err {
	case nil:
		rout, err := dbwrapper.GetRoutingTable()
		switch err {
		case nil:
			id := timeline.GetMyId().ID
			dht = &KadDHT{
				id,
				rout,
				mgr,
				time.Now(),
				make(chan struct{}, 1),
			}
			launchRePing(mgr, rout)
		case badger.ErrKeyNotFound:
			//Start from the beginning
			restartDHT(dbwrapper)
		default:
			panic(err)
		}
	case badger.ErrKeyNotFound:
		_, err := dbwrapper.GetRoutingTable()
		switch err {
		case nil, badger.ErrKeyNotFound:
			//Start from the beginning
			//Make new
			restartDHT(dbwrapper)
		default:
			panic(err)
		}
	default:
		panic(err)
	}
}


func restartDHT(dbwrapper *db.DBWrapper) {
	id := timeline.GetMyId().ID
	dht = &KadDHT{
		id,
		NewRoutingTable(DefaultBucketSize, id, DefaultLatencyToleranceMillis, NewMetrics()),
		NewAddrManager(),
		time.Now(),
		make(chan struct{}, 1),
	}
	dbwrapper.SetRoutingTable(dht.RoutingTable)
	dbwrapper.SetAddrManager(dht.AddrManager)
}


func launchRePing(manager *AddrManager, table *RoutingTable) {

}

func GetDHT() (*KadDHT) {
	return dht
}