package main

import (
	"testing"
	"time"
)

func TestTaskDateUsesDhakaDayAndBounds(t *testing.T) {
	now := time.Date(2026, 10, 21, 20, 0, 0, 0, time.UTC) // Oct 22 in Dhaka.
	for _, date := range []string{"2025-10-22", "2026-10-22", "2027-10-22"} {
		if !validTaskDate(date, now) {
			t.Errorf("rejected valid date %s", date)
		}
	}
	for _, date := range []string{"2025-10-21", "2027-10-23", "2026-02-30", "", "tomorrow"} {
		if validTaskDate(date, now) {
			t.Errorf("accepted invalid date %s", date)
		}
	}
}
