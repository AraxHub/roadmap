package feedback

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"time"

	"roadmap/internal/domain"
)

// UserStore — пользователи для ОС / TG.
type UserStore interface {
	GetByID(ctx context.Context, userID string) (*domain.User, error)
	GetByLogin(ctx context.Context, login string) (*domain.User, error)
	GetByTelegramChatID(ctx context.Context, chatID int64) (*domain.User, error)
	List(ctx context.Context) ([]domain.User, error)
	ListLearnersWithTelegram(ctx context.Context) ([]domain.User, error)
	SetTelegramChatID(ctx context.Context, userID string, chatID int64) error
	SetBlocked(ctx context.Context, userID string, blocked bool) error
}

// ChainStore — опубликованная цепочка обучения.
type ChainStore interface {
	ListPublishedChain(ctx context.Context) ([]domain.ChainItem, error)
}

// ProgressStore — прогресс учеников.
type ProgressStore interface {
	ListAllCompletedByUser(ctx context.Context) (map[string]map[string]bool, error)
}

// RequestStore — просьбы об ОС.
type RequestStore interface {
	HasPending(ctx context.Context, userID string) (bool, error)
	Create(ctx context.Context, userID string, tgMessageID *int64) (*domain.FeedbackRequest, error)
	AnswerPending(ctx context.Context, userID, text string) (*domain.FeedbackRequest, error)
	ListByUser(ctx context.Context, userID string) ([]domain.FeedbackRequest, error)
}

// Messenger — отправка в Telegram.
type Messenger interface {
	SendMessage(ctx context.Context, chatID int64, text string) (int64, error)
}

// UseCase — еженедельная ОС, webhook, админ-история.
type UseCase struct {
	users    UserStore
	requests RequestStore
	schedule ScheduleStore
	chain    ChainStore
	progress ProgressStore
	tg       Messenger
	loc      *time.Location
	log      *slog.Logger

	// awaitingLogin — chat_id, которые нажали /start и ждут логин следующим сообщением.
	awaitingLogin sync.Map // int64 → struct{}
}

// New создаёт use case. tg может быть nil, если бот выключен.
func New(
	users UserStore,
	requests RequestStore,
	schedule ScheduleStore,
	chain ChainStore,
	progress ProgressStore,
	tg Messenger,
	loc *time.Location,
	log *slog.Logger,
) *UseCase {
	if loc == nil {
		loc = time.FixedZone("MSK", 3*60*60)
	}
	return &UseCase{
		users: users, requests: requests, schedule: schedule,
		chain: chain, progress: progress, tg: tg, loc: loc, log: log,
	}
}

// ScheduleInfo — состояние расписания авто-ОС для админки.
type ScheduleInfo struct {
	NextRoundAt   time.Time  `json:"next_round_at"`
	LastRoundAt   *time.Time `json:"last_round_at,omitempty"`
	SecondsToNext int64      `json:"seconds_to_next"`
	MessageText   string     `json:"message_text"`
}

// Schedule — когда следующий авто-раунд ОС (четверг 12:00 в TZ конфига) + текст просьбы.
func (u *UseCase) Schedule(ctx context.Context) (*ScheduleInfo, error) {
	now := time.Now()
	next := NextThursdayNoon(now, u.loc)
	info := &ScheduleInfo{
		NextRoundAt:   next,
		SecondsToNext: int64(time.Until(next).Seconds()),
	}
	if u.schedule != nil {
		last, err := u.schedule.GetLastRoundAt(ctx)
		if err != nil {
			return nil, err
		}
		info.LastRoundAt = last
		text, err := u.schedule.GetMessageText(ctx)
		if err != nil {
			return nil, err
		}
		info.MessageText = text
	}
	return info, nil
}

// UpdateMessage — сохранить текст просьбы об ОС (без редеплоя).
func (u *UseCase) UpdateMessage(ctx context.Context, text string) error {
	text = strings.TrimSpace(text)
	if text == "" {
		return domain.ErrInvalidUser
	}
	if u.schedule == nil {
		return domain.ErrForbidden
	}
	return u.schedule.SetMessageText(ctx, text)
}

func (u *UseCase) messageText(ctx context.Context) (string, error) {
	if u.schedule == nil {
		return "Привет! Оставь, пожалуйста, обратную связь по обучению — ответь на это сообщение.", nil
	}
	return u.schedule.GetMessageText(ctx)
}

// UserStage — текущий этап обучения (спринт / модуль / подмодуль).
type UserStage struct {
	SprintTitle    string `json:"sprint_title"`
	ModuleTitle    string `json:"module_title"`
	SubmoduleTitle string `json:"submodule_title"`
}

// UserListItem — пользователь в админ-списке.
type UserListItem struct {
	ID             string     `json:"id"`
	Login          string     `json:"login"`
	Role           string     `json:"role"`
	IsBlocked      bool       `json:"is_blocked"`
	TelegramLinked bool       `json:"telegram_linked"`
	CreatedAt      time.Time  `json:"created_at"`
	StageStatus    string     `json:"stage_status"` // in_progress | completed | empty
	CurrentStage   *UserStage `json:"current_stage,omitempty"`
}

// FeedbackItem — запись истории ОС.
type FeedbackItem struct {
	ID          string     `json:"id"`
	Status      string     `json:"status"`
	RequestedAt time.Time  `json:"requested_at"`
	AnsweredAt  *time.Time `json:"answered_at,omitempty"`
	AnswerText  string     `json:"answer_text"`
}

// ListUsers — список всех ЛК для админа с этапом обучения.
func (u *UseCase) ListUsers(ctx context.Context) ([]UserListItem, error) {
	users, err := u.users.List(ctx)
	if err != nil {
		return nil, err
	}

	var chain []domain.ChainItem
	var completedByUser map[string]map[string]bool
	if u.chain != nil {
		chain, err = u.chain.ListPublishedChain(ctx)
		if err != nil {
			return nil, err
		}
	}
	if u.progress != nil {
		completedByUser, err = u.progress.ListAllCompletedByUser(ctx)
		if err != nil {
			return nil, err
		}
	}
	if completedByUser == nil {
		completedByUser = map[string]map[string]bool{}
	}

	out := make([]UserListItem, 0, len(users))
	for _, usr := range users {
		item := UserListItem{
			ID:             usr.ID,
			Login:          usr.Login,
			Role:           usr.Role,
			IsBlocked:      usr.IsBlocked,
			TelegramLinked: usr.HasTelegram(),
			CreatedAt:      usr.CreatedAt,
			StageStatus:    "empty",
		}
		if usr.Role == domain.RoleUser {
			status, stage := resolveStage(chain, completedByUser[usr.ID])
			item.StageStatus = status
			item.CurrentStage = stage
		}
		out = append(out, item)
	}
	return out, nil
}

func resolveStage(chain []domain.ChainItem, completed map[string]bool) (string, *UserStage) {
	if len(chain) == 0 {
		return "empty", nil
	}
	if completed == nil {
		completed = map[string]bool{}
	}
	unlocked := unlockedMap(chain, completed)
	allDone := true
	for _, item := range chain {
		if !completed[item.SubmoduleID] {
			allDone = false
			break
		}
	}
	if allDone {
		return "completed", nil
	}
	for _, item := range chain {
		if unlocked[item.SubmoduleID] && !completed[item.SubmoduleID] {
			return "in_progress", &UserStage{
				SprintTitle:    item.SprintTitle,
				ModuleTitle:    item.ModuleTitle,
				SubmoduleTitle: item.SubmoduleTitle,
			}
		}
	}
	return "empty", nil
}

func unlockedMap(chain []domain.ChainItem, completed map[string]bool) map[string]bool {
	out := make(map[string]bool, len(chain))
	for i, item := range chain {
		if i == 0 {
			out[item.SubmoduleID] = true
			continue
		}
		if completed[chain[i-1].SubmoduleID] {
			out[item.SubmoduleID] = true
		}
	}
	return out
}

// ListFeedback — история ОС по пользователю.
func (u *UseCase) ListFeedback(ctx context.Context, userID string) ([]FeedbackItem, error) {
	if _, err := u.users.GetByID(ctx, userID); err != nil {
		return nil, err
	}
	rows, err := u.requests.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]FeedbackItem, 0, len(rows))
	for _, fr := range rows {
		out = append(out, FeedbackItem{
			ID:          fr.ID,
			Status:      fr.Status,
			RequestedAt: fr.RequestedAt,
			AnsweredAt:  fr.AnsweredAt,
			AnswerText:  fr.AnswerText,
		})
	}
	return out, nil
}

// RequestRound — разослать просьбу об ОС всем привязанным learners без pending.
func (u *UseCase) RequestRound(ctx context.Context) error {
	if u.tg == nil {
		u.log.Info("feedback round skipped: telegram disabled")
		return nil
	}
	users, err := u.users.ListLearnersWithTelegram(ctx)
	if err != nil {
		return err
	}
	for _, usr := range users {
		if err := u.requestForUser(ctx, usr); err != nil {
			u.log.Error("feedback request failed", "user_id", usr.ID, "error", err)
		}
	}
	return nil
}

// RequestForUser — просьба об ОС одному пользователю (кнопка в админке).
func (u *UseCase) RequestForUser(ctx context.Context, userID string) error {
	if u.tg == nil {
		return domain.ErrForbidden
	}
	usr, err := u.users.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if usr.IsAdmin() {
		return domain.ErrForbidden
	}
	if !usr.HasTelegram() {
		return domain.ErrNotFound
	}
	return u.requestForUser(ctx, *usr)
}

func (u *UseCase) requestForUser(ctx context.Context, usr domain.User) error {
	pending, err := u.requests.HasPending(ctx, usr.ID)
	if err != nil {
		return err
	}
	if pending {
		return nil
	}
	if !usr.HasTelegram() {
		return nil
	}
	text, err := u.messageText(ctx)
	if err != nil {
		return err
	}
	msgID, err := u.tg.SendMessage(ctx, *usr.TelegramChatID, text)
	if err != nil {
		return err
	}
	id := msgID
	if _, err := u.requests.Create(ctx, usr.ID, &id); err != nil {
		return err
	}
	return u.users.SetBlocked(ctx, usr.ID, true)
}

// HandleUpdate — обработка update от Telegram (polling или webhook).
//
// Привязка ЛК:
//  1. /start → бот просит прислать логин
//  2. следующим сообщением — логин → привязка chat_id
//
// Уже привязанный chat: любой текст (кроме /start) = ответ на pending ОС.
func (u *UseCase) HandleUpdate(ctx context.Context, chatID int64, text string) error {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}

	if strings.HasPrefix(text, "/start") {
		return u.handleStart(ctx, chatID, text)
	}

	if _, waiting := u.awaitingLogin.Load(chatID); waiting {
		u.awaitingLogin.Delete(chatID)
		return u.linkTelegram(ctx, chatID, text)
	}

	usr, err := u.users.GetByTelegramChatID(ctx, chatID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			u.log.Debug("ignore telegram message from unknown chat", "chat_id", chatID)
			return nil
		}
		return err
	}

	if _, err := u.requests.AnswerPending(ctx, usr.ID, text); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			u.log.Debug("no pending feedback for user", "user_id", usr.ID)
			return nil
		}
		return err
	}
	return u.users.SetBlocked(ctx, usr.ID, false)
}

func (u *UseCase) handleStart(ctx context.Context, chatID int64, text string) error {
	// На всякий: /start login всё ещё работает одним сообщением.
	if login, ok := parseStartPayload(text); ok {
		u.awaitingLogin.Delete(chatID)
		return u.linkTelegram(ctx, chatID, login)
	}

	if usr, err := u.users.GetByTelegramChatID(ctx, chatID); err == nil && usr != nil {
		u.reply(ctx, chatID, "Ты уже привязан как «"+usr.Login+"». Если нужна обратная связь — просто ответь на сообщение бота.")
		return nil
	}

	u.awaitingLogin.Store(chatID, struct{}{})
	u.reply(ctx, chatID, "Привет! Пришли следующим сообщением свой логин от личного кабинета Roadmap.")
	return nil
}

func (u *UseCase) linkTelegram(ctx context.Context, chatID int64, login string) error {
	login = strings.TrimSpace(login)
	if login == "" || strings.HasPrefix(login, "/") {
		u.awaitingLogin.Store(chatID, struct{}{})
		u.reply(ctx, chatID, "Пришли логин обычным текстом, без команд.")
		return nil
	}

	usr, err := u.users.GetByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			u.awaitingLogin.Store(chatID, struct{}{})
			u.reply(ctx, chatID, "Логин «"+login+"» не найден. Проверь написание и пришли ещё раз.")
			return nil
		}
		return err
	}
	if usr.IsAdmin() {
		u.reply(ctx, chatID, "Админские аккаунты к боту не привязываются.")
		return nil
	}
	if err := u.users.SetTelegramChatID(ctx, usr.ID, chatID); err != nil {
		if errors.Is(err, domain.ErrConflict) {
			u.reply(ctx, chatID, "Этот Telegram уже привязан к другому кабинету.")
			return nil
		}
		return err
	}
	u.reply(ctx, chatID, "Готово: аккаунт «"+usr.Login+"» привязан. Жди еженедельный опрос по обратной связи.")
	return nil
}

func (u *UseCase) reply(ctx context.Context, chatID int64, text string) {
	if u.tg == nil {
		return
	}
	if _, err := u.tg.SendMessage(ctx, chatID, text); err != nil {
		u.log.Error("telegram reply failed", "chat_id", chatID, "error", err)
	}
}

func parseStartPayload(text string) (string, bool) {
	parts := strings.Fields(text)
	if len(parts) < 2 {
		return "", false
	}
	if !strings.HasPrefix(parts[0], "/start") {
		return "", false
	}
	login := strings.TrimSpace(parts[1])
	if login == "" {
		return "", false
	}
	return login, true
}
