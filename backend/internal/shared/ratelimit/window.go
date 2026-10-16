package ratelimit

import (
	"sync"
	"time"
)

const maxTrackedKeys = 10_000

type count struct {
	start time.Time
	hits  int
}

type Window struct {
	mutex  sync.Mutex
	limit  int
	window time.Duration
	now    func() time.Time
	counts map[string]count
}

func NewWindow(limit int, window time.Duration) *Window {
	return &Window{limit: limit, window: window, now: time.Now, counts: map[string]count{}}
}

func (w *Window) SetClock(now func() time.Time) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	w.now = now
}

func (w *Window) Allow(key string) bool {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	now := w.now()
	if len(w.counts) >= maxTrackedKeys {
		w.forgetExpired(now)
	}
	current := w.counts[key]
	if now.Sub(current.start) >= w.window {
		current = count{start: now}
	}
	if current.hits >= w.limit {
		return false
	}
	current.hits++
	w.counts[key] = current
	return true
}

func (w *Window) forgetExpired(now time.Time) {
	for key, current := range w.counts {
		if now.Sub(current.start) >= w.window {
			delete(w.counts, key)
		}
	}
}
