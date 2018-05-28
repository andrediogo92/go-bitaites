/**
The MIT License (MIT)

Copyright (c) 2016 Protocol Labs, Inc.

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package pex

import (
	"log"
	"math"
	"net"
	"sync"
	"time"

	"github.com/Seriyin/go-bitaites/peer"
)


// Permanent TTLs (distinct so we can distinguish between them, constant as they
// are, in fact, permanent)
const (
	// TempAddrTTL is the ttl used for a short lived address
	TempAddrTTL = time.Second * 10

	// ProviderAddrTTL is the TTL of an address we've received from a provider.
	// This is also a temporary address, but lasts longer. After this expires,
	// the records we return will require an extra lookup.
	ProviderAddrTTL = time.Minute * 10

	// RecentlyConnectedAddrTTL is used when we recently connected to a peer.
	// It means that we are reasonably certain of the peer's address.
	RecentlyConnectedAddrTTL = time.Minute * 10

	// OwnObservedAddrTTL is used for our own external addresses observed by peers.
	OwnObservedAddrTTL = time.Minute * 10


	// PermanentAddrTTL is the ttl for a "permanent address" (e.g. bootstrap nodes).
	PermanentAddrTTL = math.MaxInt64 - iota

	// ConnectedAddrTTL is the ttl used for the addresses of a peer to whom
	// we're connected directly. This is basically permanent, as we will
	// clear them + re-add under a TempAddrTTL after disconnecting.
	ConnectedAddrTTL
)

type expiringAddr struct {
	Addr    net.Addr
	TTL     time.Duration
	Expires time.Time
}

func (e *expiringAddr) ExpiredBy(t time.Time) bool {
	return t.After(e.Expires)
}


// AddrManager manages addresses.
// The zero-value is ready to be used.
type AddrManager struct {
	sync.Mutex // guards addrs
	addrs  map[peer.ID]expiringAddr

}

// ensures the AddrManager is initialized.
// So we can use the zero value.
func NewAddrManager() (*AddrManager) {
	return &AddrManager{
		sync.Mutex{},
		make(map[peer.ID]expiringAddr),
	}
}

func (mgr *AddrManager) Peers() []peer.ID {
	mgr.Lock()
	defer mgr.Unlock()

	pids := make([]peer.ID, 0, len(mgr.addrs))
	for pid := range mgr.addrs {
		pids = append(pids, pid)
	}
	return pids
}

// AddAddr gives AddrManager an address to use, with a given ttl
// (time-to-live), after which the address is no longer valid.
// If the manager has a longer TTL, the operation is a no-op for that address
func (mgr *AddrManager) AddAddr(p peer.ID, addr net.Addr, ttl time.Duration) {

	// if ttl is zero, exit. nothing to do.
	if ttl <= 0 {
		return
	}

	if addr == nil {
		log.Printf("was passed nil Addr for %s\n", p)
	}

	mgr.Lock()
	defer mgr.Unlock()
	oldAddr, found := mgr.addrs[p]


	// only expand ttls
	exp := time.Now().Add(ttl)


	if !found || (addr.String() == oldAddr.Addr.String() &&
				  addr.Network() == oldAddr.Addr.Network() &&
		 		  exp.After(oldAddr.Expires)) {
		mgr.addrs[p] = expiringAddr{Addr: addr, Expires: exp, TTL: ttl}
	}
}


// SetAddr sets the ttl on address. This clears any TTL there previously.
// This is used when we receive the best estimate of the validity of an address.
func (mgr *AddrManager) SetAddr(p peer.ID, addr net.Addr, ttl time.Duration) {

	exp := time.Now().Add(ttl)

	if addr == nil {
		log.Printf("was passed nil Addr for %s\n", p)
		return
	}

	mgr.Lock()
	defer mgr.Unlock()

	_, found := mgr.addrs[p]

	if found {
		if ttl > 0 {
			mgr.addrs[p] = expiringAddr{Addr: addr, Expires: exp, TTL: ttl}
		} else {
			delete(mgr.addrs, p)
		}
	}

}

// UpdateAddr updates the addresses associated with the given peer that have
// the given oldTTL to have the given newTTL.
func (mgr *AddrManager) UpdateAddr(p peer.ID, oldTTL time.Duration, newTTL time.Duration) {
	exp := time.Now().Add(newTTL)

	mgr.Lock()
	defer mgr.Unlock()

	addr, found := mgr.addrs[p]
	if !found {
		return
	}

	aexp := &addr
	if oldTTL == aexp.TTL {
		aexp.TTL = newTTL
		aexp.Expires = exp
	}
}

// Addresses returns all known (and valid) addresses for a given peer ID.
func (mgr *AddrManager) Addr(p peer.ID) net.Addr {
	now := time.Now()

	mgr.Lock()
	defer mgr.Unlock()

	addr, found := mgr.addrs[p]
	if !found {
		return nil
	}

	if !addr.ExpiredBy(now) {
		delete(mgr.addrs, p)
		return nil
	}

	return addr.Addr
}

// ClearAddress removes previously stored address for a peer ID.
func (mgr *AddrManager) ClearAddr(p peer.ID) {
	mgr.Lock()
	defer mgr.Unlock()

	delete(mgr.addrs, p)
}


//Return all known peers and ids to ping.
//These are copies of the address Managers stored ones.
//This method is called on reinitialization of the client.
func (mgr *AddrManager) ForRePing() ([]peer.ID, []net.Addr) {
	mgr.Lock()
	defer mgr.Unlock()

	peers, addrs := make([]peer.ID, len(mgr.addrs)),
					make([]net.Addr, len(mgr.addrs))

	i:=0
	for k, v := range mgr.addrs {
		peers[i] = k
		addrs[i] = v.Addr
		i++
	}

	return peers, addrs
}