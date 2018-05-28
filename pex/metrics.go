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
	"sync"
	"time"

	"github.com/Seriyin/go-bitaites/peer"
)

// LatencyEWMASmooting governs the decay of the EWMA (the speed
// at which it changes). This must be a normalized (0-1) value.
// 1 is 100% change, 0 is no change.
var LatencyEWMASmoothing = 0.1

// Metrics is just an object that tracks metrics
// across a set of peers.
type Metrics interface {

	// RecordLatency records a new latency measurement
	RecordLatency(peer.ID, time.Duration)

	// LatencyEWMA returns an exponentially-weighted moving avg.
	// of all measurements of a peer's latency.
	LatencyEWMA(peer.ID) time.Duration
}

type metrics struct {
	latmap map[peer.ID]time.Duration
	latmu  sync.RWMutex
}

func NewMetrics() *metrics {
	return &metrics{
		latmap: make(map[peer.ID]time.Duration),
	}
}

// RecordLatency records a new latency measurement
func (m *metrics) RecordLatency(p peer.ID, next time.Duration) {
	nextf := float64(next)
	s := LatencyEWMASmoothing
	if s > 1 || s < 0 {
		s = 0.1 // ignore the knob. it's broken. look, it jiggles.
	}

	m.latmu.Lock()
	ewma, found := m.latmap[p]
	ewmaf := float64(ewma)
	if !found {
		m.latmap[p] = next // when no data, just take it as the mean.
	} else {
		nextf = ((1.0 - s) * ewmaf) + (s * nextf)
		m.latmap[p] = time.Duration(nextf)
	}
	m.latmu.Unlock()
}

// LatencyEWMA returns an exponentially-weighted moving avg.
// of all measurements of a peer's latency.
func (m *metrics) LatencyEWMA(p peer.ID) time.Duration {
	m.latmu.RLock()
	lat := m.latmap[p]
	m.latmu.RUnlock()
	return time.Duration(lat)
}
