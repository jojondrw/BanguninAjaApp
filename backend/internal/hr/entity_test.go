package hr

import (
	"sync"
	"testing"

	"gorm.io/gorm/schema"
)

func TestAttendanceClockColumnsStayTimeOfDay(t *testing.T) {
	parsed, err := schema.Parse(&Attendance{}, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		t.Fatalf("parse attendance schema: %v", err)
	}

	for _, name := range []string{"CheckInTime", "CheckOutTime"} {
		field := parsed.LookUpField(name)
		if field == nil {
			t.Fatalf("field %s not found", name)
		}
		if field.DataType == schema.Time {
			t.Fatalf("field %s maps to a timestamp column, want time of day", name)
		}
	}
}
