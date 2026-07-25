package feedback

import (
	"testing"
	"time"
)

func TestLastAndNextThursdayNoon(t *testing.T) {
	loc := time.FixedZone("MSK", 3*60*60)

	// Среда 15:00 → last = прошлый четверг 12:00, next = этот четверг 12:00
	wed := time.Date(2026, 7, 22, 15, 0, 0, 0, loc) // Wed
	last := LastThursdayNoon(wed, loc)
	wantLast := time.Date(2026, 7, 16, 12, 0, 0, 0, loc)
	if !last.Equal(wantLast) {
		t.Fatalf("wed last: got %v want %v", last, wantLast)
	}
	next := NextThursdayNoon(wed, loc)
	wantNext := time.Date(2026, 7, 23, 12, 0, 0, 0, loc)
	if !next.Equal(wantNext) {
		t.Fatalf("wed next: got %v want %v", next, wantNext)
	}

	// Четверг 11:00 → last = прошлый четверг, next = сегодня 12:00
	thuMorning := time.Date(2026, 7, 23, 11, 0, 0, 0, loc)
	last = LastThursdayNoon(thuMorning, loc)
	if !last.Equal(wantLast) {
		t.Fatalf("thu morning last: got %v want %v", last, wantLast)
	}
	next = NextThursdayNoon(thuMorning, loc)
	if !next.Equal(wantNext) {
		t.Fatalf("thu morning next: got %v want %v", next, wantNext)
	}

	// Четверг 12:00 ровно → last = сегодня
	thuNoon := time.Date(2026, 7, 23, 12, 0, 0, 0, loc)
	last = LastThursdayNoon(thuNoon, loc)
	if !last.Equal(wantNext) {
		t.Fatalf("thu noon last: got %v want %v", last, wantNext)
	}
	next = NextThursdayNoon(thuNoon, loc)
	wantNextWeek := time.Date(2026, 7, 30, 12, 0, 0, 0, loc)
	if !next.Equal(wantNextWeek) {
		t.Fatalf("thu noon next: got %v want %v", next, wantNextWeek)
	}

	// Пятница → last = вчерашний четверг
	fri := time.Date(2026, 7, 24, 10, 0, 0, 0, loc)
	last = LastThursdayNoon(fri, loc)
	if !last.Equal(wantNext) {
		t.Fatalf("fri last: got %v want %v", last, wantNext)
	}
}

func TestRoundDue(t *testing.T) {
	loc := time.FixedZone("MSK", 3*60*60)
	now := time.Date(2026, 7, 24, 10, 0, 0, 0, loc) // Fri after Thu noon
	slot := time.Date(2026, 7, 23, 12, 0, 0, 0, loc)

	due, gotSlot := RoundDue(nil, now, loc)
	if due {
		t.Fatalf("nil last must not catch up: due=%v slot=%v", due, gotSlot)
	}
	if !gotSlot.Equal(slot) {
		t.Fatalf("nil last slot: got %v want %v", gotSlot, slot)
	}

	due, _ = RoundDue(&slot, now, loc)
	if due {
		t.Fatal("same slot should not be due")
	}

	prev := slot.AddDate(0, 0, -7)
	due, gotSlot = RoundDue(&prev, now, loc)
	if !due || !gotSlot.Equal(slot) {
		t.Fatalf("prev slot: due=%v slot=%v", due, gotSlot)
	}

	// Четверг утром — ещё не due для этого четверга
	thuMorning := time.Date(2026, 7, 23, 11, 0, 0, 0, loc)
	prevSlot := time.Date(2026, 7, 16, 12, 0, 0, 0, loc)
	due, gotSlot = RoundDue(&prevSlot, thuMorning, loc)
	if due {
		t.Fatalf("thu morning should wait for noon, got due slot=%v", gotSlot)
	}
}
