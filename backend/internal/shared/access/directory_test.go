package access

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

type countingLoader struct {
	calls   int
	subject Subject
	err     error
	during  func()
}

func (l *countingLoader) load(context.Context, uuid.UUID) (Subject, error) {
	l.calls++
	if l.during != nil {
		l.during()
	}
	return l.subject, l.err
}

func newTestDirectory(loader *countingLoader, clock *time.Time) *Directory {
	directory := NewDirectory(loader.load, 30*time.Second)
	directory.now = func() time.Time { return *clock }
	return directory
}

func TestDirectoryLoadsOncePerUserWithinTTL(t *testing.T) {
	clock := time.Date(2026, 10, 3, 9, 0, 0, 0, time.UTC)
	loader := &countingLoader{subject: Subject{Role: RoleFinance, Active: true}}
	directory := newTestDirectory(loader, &clock)
	userID := uuid.New()

	for range 3 {
		subject, err := directory.Subject(context.Background(), userID)
		if err != nil || subject.Role != RoleFinance {
			t.Fatalf("got %+v, %v", subject, err)
		}
	}
	if loader.calls != 1 {
		t.Fatalf("got %d loads", loader.calls)
	}

	clock = clock.Add(31 * time.Second)
	if _, err := directory.Subject(context.Background(), userID); err != nil {
		t.Fatal(err)
	}
	if loader.calls != 2 {
		t.Fatalf("expired entry should reload, got %d loads", loader.calls)
	}
}

func TestDirectoryForgetReloadsTheNextRequest(t *testing.T) {
	clock := time.Now()
	loader := &countingLoader{subject: Subject{Role: RoleObserver, Active: true}}
	directory := newTestDirectory(loader, &clock)
	userID := uuid.New()

	_, _ = directory.Subject(context.Background(), userID)
	loader.subject = Subject{Role: RoleAdmin, Active: true}
	directory.Forget(userID)

	subject, _ := directory.Subject(context.Background(), userID)
	if subject.Role != RoleAdmin || loader.calls != 2 {
		t.Fatalf("got %+v after %d loads", subject, loader.calls)
	}
}

func TestDirectoryDropsALoadThatRacedWithForget(t *testing.T) {
	clock := time.Now()
	userID := uuid.New()
	loader := &countingLoader{subject: Subject{Role: RoleAdmin, Active: true}}
	directory := newTestDirectory(loader, &clock)
	loader.during = func() { directory.Forget(userID) }

	_, _ = directory.Subject(context.Background(), userID)
	loader.during = nil
	_, _ = directory.Subject(context.Background(), userID)

	if loader.calls != 2 {
		t.Fatalf("a stale load must not be cached, got %d loads", loader.calls)
	}
}

func TestDirectoryDoesNotCacheFailures(t *testing.T) {
	clock := time.Now()
	loader := &countingLoader{err: ErrUnknownSubject}
	directory := newTestDirectory(loader, &clock)
	userID := uuid.New()

	for range 2 {
		if _, err := directory.Subject(context.Background(), userID); !errors.Is(err, ErrUnknownSubject) {
			t.Fatalf("got %v", err)
		}
	}
	if loader.calls != 2 {
		t.Fatalf("got %d loads", loader.calls)
	}
}
