package ratelimit

import (
	"strconv"
	"testing"
	"time"
)

func TestWindowAllowsUpToTheLimitAndThenBlocksUntilTheWindowEnds(t *testing.T) {
	current := time.Date(2026, 10, 17, 9, 0, 0, 0, time.UTC)
	window := NewWindow(3, 15*time.Minute)
	window.SetClock(func() time.Time { return current })

	for attempt := 1; attempt <= 3; attempt++ {
		if !window.Allow("demo@banguninaja.id") {
			t.Fatalf("attempt %d was blocked inside the limit", attempt)
		}
	}
	if window.Allow("demo@banguninaja.id") {
		t.Fatal("the fourth attempt inside the window was allowed")
	}

	current = current.Add(14*time.Minute + 59*time.Second)
	if window.Allow("demo@banguninaja.id") {
		t.Fatal("the key was released before the window ended")
	}

	current = current.Add(time.Second)
	if !window.Allow("demo@banguninaja.id") {
		t.Fatal("the key stayed blocked after the window ended")
	}
}

func TestWindowCountsEachKeySeparately(t *testing.T) {
	window := NewWindow(1, time.Minute)

	if !window.Allow("a@uji.test") || !window.Allow("b@uji.test") {
		t.Fatal("one key used up the limit of another")
	}
	if window.Allow("a@uji.test") {
		t.Fatal("the second attempt for the same key was allowed")
	}
}

func TestWindowForgetsExpiredKeysWhenFull(t *testing.T) {
	current := time.Date(2026, 10, 17, 9, 0, 0, 0, time.UTC)
	window := NewWindow(1, time.Minute)
	window.SetClock(func() time.Time { return current })

	for i := 0; i < maxTrackedKeys; i++ {
		window.Allow("kunci-" + strconv.Itoa(i))
	}
	current = current.Add(time.Minute)
	window.Allow("kunci-baru")

	if len(window.counts) != 1 {
		t.Fatalf("tracking %d keys after every old key expired, want 1", len(window.counts))
	}
}
