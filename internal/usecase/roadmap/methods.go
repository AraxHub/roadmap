package roadmap

import (
	"context"
	"errors"
	"time"

	"roadmap/internal/domain"
)

// unlockedMap строит множество разблокированных подмодулей по цепочке и прогрессу.
// Первый всегда открыт; N открыт, если N-1 завершён.
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

func (u *UseCase) loadAccess(ctx context.Context, userID string) ([]domain.ChainItem, map[string]bool, map[string]bool, error) {
	ok, err := u.users.Exists(ctx, userID)
	if err != nil {
		return nil, nil, nil, err
	}
	if !ok {
		return nil, nil, nil, domain.ErrInvalidUser
	}

	chain, err := u.subs.ListPublishedChain(ctx)
	if err != nil {
		return nil, nil, nil, err
	}
	completed, err := u.progress.ListCompletedSubmoduleIDs(ctx, userID)
	if err != nil {
		return nil, nil, nil, err
	}
	unlocked := unlockedMap(chain, completed)
	return chain, completed, unlocked, nil
}

// Home — главная: спринты и модули с lock/% прогресса и таймерами.
func (u *UseCase) Home(ctx context.Context, userID string) (*HomeView, error) {
	chain, completed, unlocked, err := u.loadAccess(ctx, userID)
	if err != nil {
		return nil, err
	}

	user, err := u.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	sprints, err := u.sprints.ListPublished(ctx)
	if err != nil {
		return nil, err
	}

	if err := u.ensureActiveSprintTimer(ctx, userID, chain, unlocked, completed, sprints); err != nil {
		return nil, err
	}

	timers, err := u.timers.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	timerBySprint := make(map[string]domain.UserSprintTimer, len(timers))
	for _, t := range timers {
		timerBySprint[t.SprintID] = t
	}

	type counters struct {
		total, done int
		anyUnlock   bool
	}
	byModule := make(map[string]*counters)
	for _, item := range chain {
		c := byModule[item.ModuleID]
		if c == nil {
			c = &counters{}
			byModule[item.ModuleID] = c
		}
		c.total++
		if completed[item.SubmoduleID] {
			c.done++
		}
		if unlocked[item.SubmoduleID] {
			c.anyUnlock = true
		}
	}

	now := time.Now().UTC()
	view := &HomeView{
		FeedbackRequired:    user.IsBlocked && !user.IsAdmin(),
		TelegramBotUsername: u.tgBotName,
		Sprints:             make([]SprintAccess, 0, len(sprints)),
	}
	for _, s := range sprints {
		mods, err := u.modules.ListPublishedBySprint(ctx, s.ID)
		if err != nil {
			return nil, err
		}
		sa := SprintAccess{
			ID:          s.ID,
			Slug:        s.Slug,
			Title:       s.Title,
			Description: s.Description,
			Modules:     make([]ModuleAccess, 0, len(mods)),
		}
		for _, m := range mods {
			c := byModule[m.ID]
			ma := ModuleAccess{
				ID:          m.ID,
				Slug:        m.Slug,
				Title:       m.Title,
				Description: m.Description,
			}
			if c != nil {
				ma.Total = c.total
				ma.Completed = c.done
				ma.Unlocked = c.anyUnlock
			}
			if ma.Unlocked {
				sa.Unlocked = true
			}
			sa.Modules = append(sa.Modules, ma)
		}
		if t, ok := timerBySprint[s.ID]; ok {
			started := t.StartedAt
			deadline := t.DeadlineAt
			sa.StartedAt = &started
			sa.DeadlineAt = &deadline
			sa.CompletedAt = t.CompletedAt
			sa.IsOverdue = t.IsOverdue(now)
			sa.ShowTimer = !t.IsCompleted()
		}
		view.Sprints = append(view.Sprints, sa)
	}
	return view, nil
}

// ModulePage — страница модуля с подмодулями и lock-флагами.
func (u *UseCase) ModulePage(ctx context.Context, userID, moduleSlug string) (*ModuleView, error) {
	_, completed, unlocked, err := u.loadAccess(ctx, userID)
	if err != nil {
		return nil, err
	}

	m, err := u.modules.GetPublishedBySlug(ctx, moduleSlug)
	if err != nil {
		return nil, err
	}

	subs, err := u.subs.ListPublishedByModule(ctx, m.ID)
	if err != nil {
		return nil, err
	}

	view := &ModuleView{
		Module: ModuleAccess{
			ID:          m.ID,
			Slug:        m.Slug,
			Title:       m.Title,
			Description: m.Description,
			Total:       len(subs),
		},
		Submodules: make([]SubmoduleAccess, 0, len(subs)),
	}
	for _, sm := range subs {
		sa := SubmoduleAccess{
			ID:        sm.ID,
			Slug:      sm.Slug,
			Title:     sm.Title,
			Unlocked:  unlocked[sm.ID],
			Completed: completed[sm.ID],
		}
		if sa.Unlocked {
			view.Module.Unlocked = true
		}
		if sa.Completed {
			view.Module.Completed++
		}
		view.Submodules = append(view.Submodules, sa)
	}
	return view, nil
}

// SubmodulePage — контент подмодуля + меню; доступ только если unlocked.
func (u *UseCase) SubmodulePage(ctx context.Context, userID, moduleSlug, submoduleSlug string) (*SubmoduleView, error) {
	chain, completed, unlocked, err := u.loadAccess(ctx, userID)
	if err != nil {
		return nil, err
	}

	sm, m, err := u.subs.GetPublishedBySlugs(ctx, moduleSlug, submoduleSlug)
	if err != nil {
		return nil, err
	}
	if !unlocked[sm.ID] {
		return nil, domain.ErrLocked
	}

	content, err := u.contents.GetBySubmoduleID(ctx, sm.ID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			content = &domain.SubmoduleContent{SubmoduleID: sm.ID, Blocks: []domain.ContentBlock{}}
		} else {
			return nil, err
		}
	}

	return &SubmoduleView{
		Submodule: SubmoduleAccess{
			ID:        sm.ID,
			Slug:      sm.Slug,
			Title:     sm.Title,
			Unlocked:  true,
			Completed: completed[sm.ID],
		},
		Module: ModuleAccess{
			ID:          m.ID,
			Slug:        m.Slug,
			Title:       m.Title,
			Description: m.Description,
			Unlocked:    true,
		},
		Blocks: content.Blocks,
		Menu:   buildMenu(chain, unlocked, completed),
	}, nil
}

// Complete — кнопка «Завершить»: только для разблокированного подмодуля.
func (u *UseCase) Complete(ctx context.Context, userID, submoduleID string) error {
	chain, completed, unlocked, err := u.loadAccess(ctx, userID)
	if err != nil {
		return err
	}
	if !unlocked[submoduleID] {
		return domain.ErrLocked
	}
	if completed[submoduleID] {
		return nil
	}
	if err := u.progress.Complete(ctx, userID, submoduleID); err != nil {
		return err
	}
	completed[submoduleID] = true
	return u.onSubmoduleCompleted(ctx, userID, chain, completed, submoduleID)
}

// ensureActiveSprintTimer стартует таймер текущего открытого незакрытого спринта.
func (u *UseCase) ensureActiveSprintTimer(
	ctx context.Context,
	userID string,
	chain []domain.ChainItem,
	unlocked, completed map[string]bool,
	sprints []domain.Sprint,
) error {
	currentID := activeSprintID(chain, unlocked, completed, sprints)
	if currentID == "" {
		return nil
	}
	now := time.Now().UTC()
	return u.timers.Start(ctx, userID, currentID, now, now.Add(u.duration))
}

// activeSprintID — первый опубликованный спринт, где есть unlocked-подмодуль и спринт ещё не весь пройден.
func activeSprintID(
	chain []domain.ChainItem,
	unlocked, completed map[string]bool,
	sprints []domain.Sprint,
) string {
	for _, s := range sprints {
		hasUnlock := false
		hasOpen := false
		hasAny := false
		for _, item := range chain {
			if item.SprintID != s.ID {
				continue
			}
			hasAny = true
			if unlocked[item.SubmoduleID] {
				hasUnlock = true
			}
			if !completed[item.SubmoduleID] {
				hasOpen = true
			}
		}
		if hasAny && hasUnlock && hasOpen {
			return s.ID
		}
	}
	return ""
}

func (u *UseCase) onSubmoduleCompleted(
	ctx context.Context,
	userID string,
	chain []domain.ChainItem,
	completed map[string]bool,
	submoduleID string,
) error {
	var sprintID string
	for _, item := range chain {
		if item.SubmoduleID == submoduleID {
			sprintID = item.SprintID
			break
		}
	}
	if sprintID == "" {
		return nil
	}

	now := time.Now().UTC()
	if !sprintFullyCompleted(chain, completed, sprintID) {
		return u.timers.Start(ctx, userID, sprintID, now, now.Add(u.duration))
	}

	if err := u.timers.Complete(ctx, userID, sprintID, now); err != nil {
		return err
	}

	nextID := nextSprintID(chain, sprintID)
	if nextID == "" {
		return nil
	}
	return u.timers.Start(ctx, userID, nextID, now, now.Add(u.duration))
}

func sprintFullyCompleted(chain []domain.ChainItem, completed map[string]bool, sprintID string) bool {
	found := false
	for _, item := range chain {
		if item.SprintID != sprintID {
			continue
		}
		found = true
		if !completed[item.SubmoduleID] {
			return false
		}
	}
	return found
}

func nextSprintID(chain []domain.ChainItem, currentSprintID string) string {
	order := make([]string, 0)
	seen := map[string]bool{}
	for _, item := range chain {
		if seen[item.SprintID] {
			continue
		}
		seen[item.SprintID] = true
		order = append(order, item.SprintID)
	}
	for i, id := range order {
		if id == currentSprintID && i+1 < len(order) {
			return order[i+1]
		}
	}
	return ""
}

func buildMenu(chain []domain.ChainItem, unlocked, completed map[string]bool) []MenuItem {
	sprintOrder := make([]string, 0)
	sprintMeta := map[string]domain.ChainItem{}
	moduleOrder := map[string][]string{}
	moduleMeta := map[string]domain.ChainItem{}
	subsByModule := map[string][]domain.ChainItem{}

	for _, item := range chain {
		if _, ok := sprintMeta[item.SprintID]; !ok {
			sprintMeta[item.SprintID] = item
			sprintOrder = append(sprintOrder, item.SprintID)
		}
		if _, ok := moduleMeta[item.ModuleID]; !ok {
			moduleMeta[item.ModuleID] = item
			moduleOrder[item.SprintID] = append(moduleOrder[item.SprintID], item.ModuleID)
		}
		subsByModule[item.ModuleID] = append(subsByModule[item.ModuleID], item)
	}

	menu := make([]MenuItem, 0, len(sprintOrder))
	for _, sprintID := range sprintOrder {
		s := sprintMeta[sprintID]
		sprintItem := MenuItem{
			Type:  "sprint",
			ID:    s.SprintID,
			Slug:  s.SprintSlug,
			Title: s.SprintTitle,
		}
		for _, moduleID := range moduleOrder[sprintID] {
			m := moduleMeta[moduleID]
			modItem := MenuItem{
				Type:  "module",
				ID:    m.ModuleID,
				Slug:  m.ModuleSlug,
				Title: m.ModuleTitle,
			}
			for _, sm := range subsByModule[moduleID] {
				subUnlocked := unlocked[sm.SubmoduleID]
				if subUnlocked {
					modItem.Unlocked = true
					sprintItem.Unlocked = true
				}
				modItem.Children = append(modItem.Children, MenuItem{
					Type:      "submodule",
					ID:        sm.SubmoduleID,
					Slug:      sm.SubmoduleSlug,
					Title:     sm.SubmoduleTitle,
					Unlocked:  subUnlocked,
					Completed: completed[sm.SubmoduleID],
				})
			}
			sprintItem.Children = append(sprintItem.Children, modItem)
		}
		menu = append(menu, sprintItem)
	}
	return menu
}
