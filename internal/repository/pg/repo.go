package pg

import (
	"log/slog"
	"strings"

	infrapg "roadmap/internal/infrastructure/pg"
)

// Repo — набор репозиториев по таблицам.
type Repo struct {
	Sprints               *SprintRepo
	Modules               *ModuleRepo
	Submodules            *SubmoduleRepo
	SubmoduleContents     *SubmoduleContentRepo
	Users                 *UserRepo
	RefreshTokens         *RefreshTokenRepo
	UserSubmoduleProgress *UserSubmoduleProgressRepo
	UserSprintTimers      *UserSprintTimerRepo
	FeedbackRequests      *FeedbackRequestRepo
	FeedbackSchedule      *FeedbackScheduleRepo
}

// NewRepo создаёт репозитории поверх одной обёртки БД.
func NewRepo(db *infrapg.DB, log *slog.Logger) *Repo {
	return &Repo{
		Sprints:               NewSprintRepo(db, log),
		Modules:               NewModuleRepo(db, log),
		Submodules:            NewSubmoduleRepo(db, log),
		SubmoduleContents:     NewSubmoduleContentRepo(db, log),
		Users:                 NewUserRepo(db, log),
		RefreshTokens:         NewRefreshTokenRepo(db, log),
		UserSubmoduleProgress: NewUserSubmoduleProgressRepo(db, log),
		UserSprintTimers:      NewUserSprintTimerRepo(db, log),
		FeedbackRequests:      NewFeedbackRequestRepo(db, log),
		FeedbackSchedule:      NewFeedbackScheduleRepo(db, log),
	}
}

// selectList собирает список колонок для SELECT; при alias != "" префиксирует alias.
func selectList(alias string, cols []string) string {
	if alias == "" {
		return strings.Join(cols, ", ")
	}
	out := make([]string, len(cols))
	for i, c := range cols {
		out[i] = alias + "." + c
	}
	return strings.Join(out, ", ")
}
