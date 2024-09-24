package tracker

import (
	"sync"
	"time"
)

type RequestTrackerPair struct {
	StartTime time.Time // start
	EndTime   time.Time
}

type Tracker struct {
	reqTracker map[string]RequestTrackerPair
	lock       sync.RWMutex
}

func NewRequestTracker() *Tracker {
	return &Tracker{
		reqTracker: make(map[string]RequestTrackerPair),
	}
}

func (t *Tracker) GetAll() map[string]RequestTrackerPair {
	return t.reqTracker
}

func (t *Tracker) Get(key string) RequestTrackerPair {
	t.lock.RLock()
	defer t.lock.RUnlock()
	return t.reqTracker[key]
}
func (t *Tracker) Set(key string, value RequestTrackerPair) {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.reqTracker[key] = value
}
