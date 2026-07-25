package feedback

import (
	"context"
	"log/slog"
	"time"
)

// ScheduleStore — персист последнего авто-раунда ОС и текста просьбы.
type ScheduleStore interface {
	GetLastRoundAt(ctx context.Context) (*time.Time, error)
	MarkRoundAt(ctx context.Context, at time.Time) error
	GetMessageText(ctx context.Context) (string, error)
	SetMessageText(ctx context.Context, text string) error
}

// Worker — авто-рассылка ОС каждый четверг в 12:00 (TZ из конфига).
// Время последнего раунда хранится в БД — рестарт не сбрасывает расписание.
type Worker struct {
	uc       *UseCase
	schedule ScheduleStore
	loc      *time.Location
	log      *slog.Logger
}

// NewWorker создаёт воркер. loc — часовой пояс слота (по умолчанию Europe/Moscow).
func NewWorker(uc *UseCase, schedule ScheduleStore, loc *time.Location, log *slog.Logger) *Worker {
	if loc == nil {
		loc = time.FixedZone("MSK", 3*60*60)
	}
	return &Worker{uc: uc, schedule: schedule, loc: loc, log: log}
}

const pollInterval = time.Minute

// Run крутит цикл до отмены ctx. При пропуске четверга (даунтайм) догоняет раунд при старте.
func (w *Worker) Run(ctx context.Context) {
	w.log.Info("feedback worker started", "tz", w.loc.String(), "slot", "Thursday 12:00")
	for {
		if err := w.tick(ctx); err != nil {
			if ctx.Err() != nil {
				w.log.Info("feedback worker stopped")
				return
			}
			w.log.Error("feedback worker tick failed", "error", err)
		}

		now := time.Now()
		wait := time.Until(NextThursdayNoon(now, w.loc))
		if wait < pollInterval {
			wait = pollInterval
		}
		// Не спим дольше часа за раз — чтобы быстрее подхватить cancel и не промахнуться из‑за drift.
		if wait > time.Hour {
			wait = time.Hour
		}

		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			w.log.Info("feedback worker stopped")
			return
		case <-timer.C:
		}
	}
}

func (w *Worker) tick(ctx context.Context) error {
	last, err := w.schedule.GetLastRoundAt(ctx)
	if err != nil {
		return err
	}
	now := time.Now()
	slot := LastThursdayNoon(now, w.loc)

	// Первый запуск: фиксируем текущий/прошлый слот без рассылки — следующий опрос в ближайший четверг 12:00.
	if last == nil {
		if err := w.schedule.MarkRoundAt(ctx, slot); err != nil {
			return err
		}
		w.log.Info("feedback schedule initialized",
			"seeded_slot", slot,
			"next", NextThursdayNoon(now, w.loc),
		)
		return nil
	}

	due, slot := RoundDue(last, now, w.loc)
	if !due {
		w.log.Debug("feedback round not due",
			"last_round_at", last,
			"next", NextThursdayNoon(now, w.loc),
		)
		return nil
	}

	w.log.Info("feedback round starting", "slot", slot)
	if err := w.uc.RequestRound(ctx); err != nil {
		return err
	}
	if err := w.schedule.MarkRoundAt(ctx, slot); err != nil {
		return err
	}
	w.log.Info("feedback round done", "slot", slot, "next", NextThursdayNoon(now, w.loc))
	return nil
}
