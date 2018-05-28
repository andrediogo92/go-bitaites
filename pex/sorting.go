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
	"container/list"
	"sort"

	"github.com/Seriyin/go-bitaites/peer"
	"github.com/Seriyin/go-bitaites/peer/keyspace"
)

// A helper struct to sort peers by their distance to the local node
type peerDistance struct {
	p        peer.ID
	distance peer.ID
}

// peerSorterArr implements sort.Interface to sort peers by xor distance
type peerSorterArr []*peerDistance

func (p peerSorterArr) Len() int      { return len(p) }
func (p peerSorterArr) Swap(a, b int) { p[a], p[b] = p[b], p[a] }
func (p peerSorterArr) Less(a, b int) bool {
	return p[a].distance.Less(p[b].distance)
}


func copyPeersFromList(target peer.ID, peerArr peerSorterArr, peerList *list.List) peerSorterArr {
	if cap(peerArr) < len(peerArr)+peerList.Len() {
		newArr := make(peerSorterArr, 0, len(peerArr)+peerList.Len())
		copy(newArr, peerArr)
		peerArr = newArr
	}
	for e := peerList.Front(); e != nil; e = e.Next() {
		pID := e.Value.(peer.ID)
		pd := peerDistance{
			p:        pID,
			distance: peer.ID(keyspace.XOR([]byte(target), []byte(pID))),
		}
		peerArr = append(peerArr, &pd)
	}
	return peerArr
}

func SortClosestPeers(peers []peer.ID, target peer.ID) []peer.ID {
	psarr := make(peerSorterArr, 0, len(peers))
	for _, p := range peers {
		pd := &peerDistance{
			p:        p,
			distance: peer.ID(keyspace.XOR([]byte(target), []byte(p))),
		}
		psarr = append(psarr, pd)
	}
	sort.Sort(psarr)
	out := make([]peer.ID, 0, len(psarr))
	for _, p := range psarr {
		out = append(out, p.p)
	}
	return out
}
