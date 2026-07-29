package roadmap

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"roadmap/internal/domain"
)

// SprintStore — доступ к спринтам.
type SprintStore interface {
	ListPublished(ctx context.Context) ([]domain.Sprint, error)
}

// ModuleStore — доступ к модулям.
type ModuleStore interface {
	ListPublishedBySprint(ctx context.Context, sprintID string) ([]domain.Module, error)
	GetPublishedBySlug(ctx context.Context, slug string) (*domain.Module, error)
}

// SubmoduleStore — доступ к подмодулям.
type SubmoduleStore interface {
	ListPublishedByModule(ctx context.Context, moduleID string) ([]domain.Submodule, error)
	GetPublishedBySlugs(ctx context.Context, moduleSlug, submoduleSlug string) (*domain.Submodule, *domain.Module, error)
	ListPublishedChain(ctx context.Context) ([]domain.ChainItem, error)
}

// SubmoduleContentStore — доступ к markdown-контенту.
type SubmoduleContentStore interface {
	GetBySubmoduleID(ctx context.Context, submoduleID string) (*domain.SubmoduleContent, error)
}

// UserStore — доступ к пользователям.
type UserStore interface {
	Exists(ctx context.Context, userID string) (bool, error)
	GetByID(ctx context.Context, userID string) (*domain.User, error)
}

// ProgressStore — доступ к прогрессу.
type ProgressStore interface {
	ListCompletedSubmoduleIDs(ctx context.Context, userID string) (map[string]bool, error)
	Complete(ctx context.Context, userID, submoduleID string) error
}

// TimerStore — персональные таймеры спринтов.
type TimerStore interface {
	ListByUser(ctx context.Context, userID string) ([]domain.UserSprintTimer, error)
	Start(ctx context.Context, userID, sprintID string, startedAt, deadlineAt time.Time) error
	Complete(ctx context.Context, userID, sprintID string, at time.Time) error
}

// UseCase — бизнес-логика роадмапа: страницы, unlock, complete, таймеры.
type UseCase struct {
	sprints   SprintStore
	modules   ModuleStore
	subs      SubmoduleStore
	contents  SubmoduleContentStore
	users     UserStore
	progress  ProgressStore
	timers    TimerStore
	duration  time.Duration
	tgBotName string
	log       *slog.Logger
}

// New создаёт use case.
func New(
	sprints SprintStore,
	modules ModuleStore,
	subs SubmoduleStore,
	contents SubmoduleContentStore,
	users UserStore,
	progress ProgressStore,
	timers TimerStore,
	sprintDuration time.Duration,
	telegramBotUsername string,
	log *slog.Logger,
) *UseCase {
	if sprintDuration <= 0 {
		sprintDuration = 7 * 24 * time.Hour
	}
	return &UseCase{
		sprints:   sprints,
		modules:   modules,
		subs:      subs,
		contents:  contents,
		users:     users,
		progress:  progress,
		timers:    timers,
		duration:  sprintDuration,
		tgBotName: strings.TrimPrefix(telegramBotUsername, "@"),
		log:       log,
	}
}

// SubmoduleAccess — флаги доступа к подмодулю.
type SubmoduleAccess struct {
	ID        string `json:"id"`
	Slug      string `json:"slug"`
	Title     string `json:"title"`
	Unlocked  bool   `json:"unlocked"`
	Completed bool   `json:"completed"`
}

// ModuleAccess — модуль с прогрессом.
type ModuleAccess struct {
	ID          string `json:"id"`
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Unlocked    bool   `json:"unlocked"`
	Completed   int    `json:"completed_count"`
	Total       int    `json:"total_count"`
}

// SprintAccess — спринт с модулями и таймером.
type SprintAccess struct {
	ID          string         `json:"id"`
	Slug        string         `json:"slug"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Unlocked    bool           `json:"unlocked"`
	Modules     []ModuleAccess `json:"modules"`
	StartedAt   *time.Time     `json:"started_at,omitempty"`
	DeadlineAt  *time.Time     `json:"deadline_at,omitempty"`
	CompletedAt *time.Time     `json:"completed_at,omitempty"`
	IsOverdue   bool           `json:"is_overdue"`
	ShowTimer   bool           `json:"show_timer"`
}

// HomeView — главная страница.
type HomeView struct {
	FeedbackRequired    bool           `json:"feedback_required"`
	TelegramBotUsername string         `json:"telegram_bot_username,omitempty"`
	Sprints             []SprintAccess `json:"sprints"`
}

// ModuleView — страница модуля с подмодулями.
type ModuleView struct {
	Module     ModuleAccess      `json:"module"`
	Submodules []SubmoduleAccess `json:"submodules"`
}

// MenuItem — элемент сайдбара.
type MenuItem struct {
	Type      string     `json:"type"` // sprint | module | submodule
	ID        string     `json:"id"`
	Slug      string     `json:"slug"`
	Title     string     `json:"title"`
	Unlocked  bool       `json:"unlocked"`
	Completed bool       `json:"completed,omitempty"`
	Children  []MenuItem `json:"children,omitempty"`
}

// SubmoduleView — страница контента подмодуля.
type SubmoduleView struct {
	Submodule SubmoduleAccess       `json:"submodule"`
	Module    ModuleAccess          `json:"module"`
	Blocks    []domain.ContentBlock `json:"blocks"`
	Menu      []MenuItem            `json:"menu"`
}
