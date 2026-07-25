package feedback

import (
	"time"
)

// LastThursdayNoon — последний четверг 12:00 в loc, который уже наступил (≤ now).
func LastThursdayNoon(now time.Time, loc *time.Location) time.Time {
	now = now.In(loc)
	daysSince := (int(now.Weekday()) - int(time.Thursday) + 7) % 7
	candidate := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, loc).AddDate(0, 0, -daysSince)
	if now.Before(candidate) {
		candidate = candidate.AddDate(0, 0, -7)
	}
	return candidate
}

// NextThursdayNoon — ближайший будущий четверг 12:00 строго после now.
func NextThursdayNoon(now time.Time, loc *time.Location) time.Time {
	return LastThursdayNoon(now, loc).AddDate(0, 0, 7)
}

// RoundDue — нужно ли слать авто-раунд: слот четверга 12:00 уже прошёл, а last ещё не помечен этим слотом.
// last == nil означает «расписание ещё не инициализировано» — не догоняем прошлый слот (см. Worker.tick).
func RoundDue(last *time.Time, now time.Time, loc *time.Location) (due bool, slot time.Time) {
	slot = LastThursdayNoon(now, loc)
	if last == nil {
		return false, slot
	}
	if last.UTC().Before(slot.UTC()) {
		return true, slot
	}
	return false, slot
}
