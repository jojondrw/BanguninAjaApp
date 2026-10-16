package clock

import (
	"testing"
	"time"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/duedate"
)

func keepLocal(t *testing.T) {
	t.Helper()
	original := time.Local
	t.Cleanup(func() { time.Local = original })
}

func TestUseAppZoneDefaultsToJakartaSoTodayFollowsIndonesianDates(t *testing.T) {
	keepLocal(t)
	t.Setenv(zoneEnv, "")

	name, err := UseAppZone()
	if err != nil {
		t.Fatalf("use zone: %v", err)
	}
	if name != "Asia/Jakarta" || time.Local.String() != "Asia/Jakarta" {
		t.Fatalf("zone = %s, local = %s", name, time.Local)
	}

	justAfterMidnightInJakarta := time.Date(2026, 10, 16, 17, 30, 0, 0, time.UTC).In(time.Local)
	if got := duedate.Today(justAfterMidnightInJakarta); !got.Equal(time.Date(2026, 10, 17, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("today at 00:30 WIB = %s, want 2026-10-17", got.Format(time.DateOnly))
	}
}

func TestUseAppZoneHonoursTheEnvironment(t *testing.T) {
	keepLocal(t)
	t.Setenv(zoneEnv, "Asia/Makassar")

	if name, err := UseAppZone(); err != nil || name != "Asia/Makassar" || time.Local.String() != "Asia/Makassar" {
		t.Fatalf("zone = %s, local = %s, err = %v", name, time.Local, err)
	}
}

func TestUseAppZoneRejectsUnknownZonesWithoutChangingLocal(t *testing.T) {
	keepLocal(t)
	before := time.Local
	t.Setenv(zoneEnv, "Mars/Olympus")

	if _, err := UseAppZone(); err == nil {
		t.Fatal("an unknown zone was accepted")
	}
	if time.Local != before {
		t.Fatal("time.Local changed after a failed zone lookup")
	}
}
