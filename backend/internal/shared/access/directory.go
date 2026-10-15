package access

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

const sweepThreshold = 1024

var ErrUnknownSubject = errors.New("subject not found")

type Subject struct {
	Role   string
	Active bool
}

type Loader func(ctx context.Context, userID uuid.UUID) (Subject, error)

type Directory struct {
	load       Loader
	ttl        time.Duration
	now        func() time.Time
	mutex      sync.Mutex
	entries    map[uuid.UUID]cachedSubject
	generation uint64
}

type cachedSubject struct {
	subject   Subject
	expiresAt time.Time
}

func NewDirectory(load Loader, ttl time.Duration) *Directory {
	return &Directory{load: load, ttl: ttl, now: time.Now, entries: map[uuid.UUID]cachedSubject{}}
}

func (d *Directory) Subject(ctx context.Context, userID uuid.UUID) (Subject, error) {
	subject, generation, ok := d.cached(userID)
	if ok {
		return subject, nil
	}

	subject, err := d.load(ctx, userID)
	if err != nil {
		return Subject{}, err
	}

	d.remember(userID, subject, generation)
	return subject, nil
}

func (d *Directory) Forget(userID uuid.UUID) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	delete(d.entries, userID)
	d.generation++
}

func (d *Directory) cached(userID uuid.UUID) (Subject, uint64, bool) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	entry, ok := d.entries[userID]
	if !ok || !d.now().Before(entry.expiresAt) {
		return Subject{}, d.generation, false
	}
	return entry.subject, d.generation, true
}

func (d *Directory) remember(userID uuid.UUID, subject Subject, loadedAt uint64) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if d.generation != loadedAt {
		return
	}
	now := d.now()
	if len(d.entries) >= sweepThreshold {
		d.sweep(now)
	}
	d.entries[userID] = cachedSubject{subject: subject, expiresAt: now.Add(d.ttl)}
}

func (d *Directory) sweep(now time.Time) {
	for userID, entry := range d.entries {
		if !now.Before(entry.expiresAt) {
			delete(d.entries, userID)
		}
	}
}
