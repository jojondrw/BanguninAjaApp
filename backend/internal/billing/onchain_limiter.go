package billing

import (
	"net/http"
	"sync"
	"time"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/apperror"
)

const maxTrackedLimiterKeys = 10_000

var errTooManyAttempts = apperror.New(http.StatusTooManyRequests, "too_many_requests", "Terlalu banyak percobaan. Tunggu sebentar lalu coba lagi")

type windowCount struct {
	start time.Time
	hits  int
}

type windowLimiter struct {
	mutex  sync.Mutex
	limit  int
	window time.Duration
	now    func() time.Time
	counts map[string]windowCount
}

func newWindowLimiter(limit int, window time.Duration) *windowLimiter {
	return &windowLimiter{limit: limit, window: window, now: time.Now, counts: map[string]windowCount{}}
}

func (l *windowLimiter) Allow(key string) bool {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	now := l.now()
	if len(l.counts) >= maxTrackedLimiterKeys {
		l.forgetExpired(now)
	}
	count := l.counts[key]
	if now.Sub(count.start) >= l.window {
		count = windowCount{start: now}
	}
	if count.hits >= l.limit {
		return false
	}
	count.hits++
	l.counts[key] = count
	return true
}

func (l *windowLimiter) forgetExpired(now time.Time) {
	for key, count := range l.counts {
		if now.Sub(count.start) >= l.window {
			delete(l.counts, key)
		}
	}
}
