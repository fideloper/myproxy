package reverseproxy

import (
	"errors"
	"github.com/gorilla/mux"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

type Target struct {
	router       *mux.Router
	upstreams    []*Upstream
	lastUpstream int
	lock         sync.Mutex
}

type Upstream struct {
	url      *url.URL
	failures int32
}

func (u *Upstream) Eligible() bool {
	if u.failures >= 5 {
		return false
	}
	return true
}

func (u *Upstream) CountFailure() {
	// Add failure to upstream
	atomic.AddInt32(&u.failures, 1)

	// Subtract failure from upstream after
	// a set period of time
	go func(upstream *Upstream, d time.Duration) {
		timer := time.NewTimer(d)
		select {
		case <-timer.C:
		}
		atomic.AddInt32(&upstream.failures, -1)
	}(u, time.Second*5)
}

// SelectUpstream will load balance amongst available
// targets using a round-robin algorithm
func (t *Target) SelectUpstream() (*Upstream, error) {
	count := len(t.upstreams)
	if count == 1 {
		if !t.upstreams[0].Eligible() {
			return nil, errors.New("target has no eligible upstream")
		}
		return t.upstreams[0], nil
	}

	t.lock.Lock()
	defer t.lock.Unlock()

	next := t.lastUpstream + 1
	if next >= count {
		next = 0
	}

	attempts := 1
	for !t.upstreams[next].Eligible() {
		// We'll only attempt as many times as there are upstreams
		if attempts >= count {
			return nil, errors.New("target has no eligible upstream")
		}

		next := t.lastUpstream + 1
		if next >= count {
			next = 0
		}
		attempts++
	}

	t.lastUpstream = next

	return t.upstreams[next], nil
}
