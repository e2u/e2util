package e2time

import (
	"testing"
	"time"
)

func TestMustParse(t *testing.T) {
	got := MustParse("2006-01-02", "2024-05-01")
	if got.Year() != 2024 || got.Month() != time.May || got.Day() != 1 {
		t.Errorf("MustParse = %v", got)
	}
	if !MustParse("2006-01-02", "not-a-date").IsZero() {
		t.Fatal("invalid parse should return zero time")
	}
}

func TestToDayAndAddDay(t *testing.T) {
	today := ToDay()
	now := time.Now()
	if today.Year() != now.Year() || today.Month() != now.Month() || today.Day() != now.Day() {
		t.Errorf("ToDay = %v, now = %v", today, now)
	}
	if today.Hour() != 0 || today.Minute() != 0 || today.Second() != 0 {
		t.Errorf("ToDay should truncate clock, got %v", today)
	}
	next := AddDay(today, 1)
	if next.Sub(today) != 24*time.Hour {
		t.Errorf("AddDay(1) delta = %v", next.Sub(today))
	}
}

func TestTimePointer(t *testing.T) {
	in := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	p := TimePointer(in)
	if p == nil || !p.Equal(in) {
		t.Errorf("TimePointer = %v, want %v", p, in)
	}
}

func TestSleepRandom(t *testing.T) {
	min := time.Millisecond
	max := 3 * time.Millisecond
	d := SleepRandom(min, max)
	if d < min || d > max {
		t.Errorf("SleepRandom duration %v not in [%v, %v]", d, min, max)
	}
}
