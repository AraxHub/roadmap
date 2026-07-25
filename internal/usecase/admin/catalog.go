package admin

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"unicode"

	"roadmap/internal/domain"
)

type SprintStore interface {
	ListAll(ctx context.Context) ([]domain.Sprint, error)
	NextPosition(ctx context.Context) (int, error)
	Create(ctx context.Context, s domain.Sprint) (*domain.Sprint, error)
	Update(ctx context.Context, s domain.Sprint) (*domain.Sprint, error)
	Delete(ctx context.Context, id string) error
}

type ModuleStore interface {
	ListAllBySprint(ctx context.Context, sprintID string) ([]domain.Module, error)
	NextPosition(ctx context.Context, sprintID string) (int, error)
	Create(ctx context.Context, m domain.Module) (*domain.Module, error)
	Update(ctx context.Context, m domain.Module) (*domain.Module, error)
	Delete(ctx context.Context, id string) error
}

type SubmoduleStore interface {
	ListAllByModule(ctx context.Context, moduleID string) ([]domain.Submodule, error)
	NextPosition(ctx context.Context, moduleID string) (int, error)
	Create(ctx context.Context, sm domain.Submodule) (*domain.Submodule, error)
	Update(ctx context.Context, sm domain.Submodule) (*domain.Submodule, error)
	Delete(ctx context.Context, id string) error
}

type ContentStore interface {
	GetBySubmoduleID(ctx context.Context, submoduleID string) (*domain.SubmoduleContent, error)
	Upsert(ctx context.Context, submoduleID, bodyMD string) (*domain.SubmoduleContent, error)
}

// UseCase — админ-конфиг роадмапа.
type UseCase struct {
	sprints  SprintStore
	modules  ModuleStore
	subs     SubmoduleStore
	contents ContentStore
	log      *slog.Logger
}

func New(sprints SprintStore, modules ModuleStore, subs SubmoduleStore, contents ContentStore, log *slog.Logger) *UseCase {
	return &UseCase{sprints: sprints, modules: modules, subs: subs, contents: contents, log: log}
}

type TreeSubmodule struct {
	domain.Submodule
}

type TreeModule struct {
	domain.Module
	Submodules []TreeSubmodule `json:"submodules"`
}

type TreeSprint struct {
	domain.Sprint
	Modules []TreeModule `json:"modules"`
}

type Tree struct {
	Sprints []TreeSprint `json:"sprints"`
}

func (u *UseCase) Tree(ctx context.Context) (*Tree, error) {
	sprints, err := u.sprints.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	out := &Tree{Sprints: make([]TreeSprint, 0, len(sprints))}
	for _, s := range sprints {
		mods, err := u.modules.ListAllBySprint(ctx, s.ID)
		if err != nil {
			return nil, err
		}
		ts := TreeSprint{Sprint: s, Modules: make([]TreeModule, 0, len(mods))}
		for _, m := range mods {
			subs, err := u.subs.ListAllByModule(ctx, m.ID)
			if err != nil {
				return nil, err
			}
			tm := TreeModule{Module: m, Submodules: make([]TreeSubmodule, 0, len(subs))}
			for _, sm := range subs {
				tm.Submodules = append(tm.Submodules, TreeSubmodule{Submodule: sm})
			}
			ts.Modules = append(ts.Modules, tm)
		}
		out.Sprints = append(out.Sprints, ts)
	}
	return out, nil
}

type SprintInput struct {
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Position    int    `json:"position"`
	IsPublished bool   `json:"is_published"`
}

func (u *UseCase) CreateSprint(ctx context.Context, in SprintInput) (*domain.Sprint, error) {
	slug := normalizeSlug(in.Slug, in.Title)
	pos, err := u.sprints.NextPosition(ctx)
	if err != nil {
		return nil, err
	}
	return u.sprints.Create(ctx, domain.Sprint{
		Slug: slug, Title: in.Title, Description: in.Description,
		Position: pos, IsPublished: in.IsPublished,
	})
}

func (u *UseCase) UpdateSprint(ctx context.Context, id string, in SprintInput) (*domain.Sprint, error) {
	slug := normalizeSlug(in.Slug, in.Title)
	return u.sprints.Update(ctx, domain.Sprint{
		ID: id, Slug: slug, Title: in.Title, Description: in.Description,
		Position: in.Position, IsPublished: in.IsPublished,
	})
}

func (u *UseCase) DeleteSprint(ctx context.Context, id string) error {
	return u.sprints.Delete(ctx, id)
}

type ModuleInput struct {
	SprintID    string `json:"sprint_id"`
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Position    int    `json:"position"`
	IsPublished bool   `json:"is_published"`
}

func (u *UseCase) CreateModule(ctx context.Context, in ModuleInput) (*domain.Module, error) {
	slug := normalizeSlug(in.Slug, in.Title)
	pos, err := u.modules.NextPosition(ctx, in.SprintID)
	if err != nil {
		return nil, err
	}
	return u.modules.Create(ctx, domain.Module{
		SprintID: in.SprintID, Slug: slug, Title: in.Title, Description: in.Description,
		Position: pos, IsPublished: in.IsPublished,
	})
}

func (u *UseCase) UpdateModule(ctx context.Context, id string, in ModuleInput) (*domain.Module, error) {
	slug := normalizeSlug(in.Slug, in.Title)
	return u.modules.Update(ctx, domain.Module{
		ID: id, SprintID: in.SprintID, Slug: slug, Title: in.Title, Description: in.Description,
		Position: in.Position, IsPublished: in.IsPublished,
	})
}

func (u *UseCase) DeleteModule(ctx context.Context, id string) error {
	return u.modules.Delete(ctx, id)
}

type SubmoduleInput struct {
	ModuleID    string `json:"module_id"`
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Position    int    `json:"position"`
	IsPublished bool   `json:"is_published"`
	BodyMD      string `json:"body_md"`
}

func (u *UseCase) CreateSubmodule(ctx context.Context, in SubmoduleInput) (*domain.Submodule, error) {
	slug := normalizeSlug(in.Slug, in.Title)
	pos, err := u.subs.NextPosition(ctx, in.ModuleID)
	if err != nil {
		return nil, err
	}
	sm, err := u.subs.Create(ctx, domain.Submodule{
		ModuleID: in.ModuleID, Slug: slug, Title: in.Title,
		Position: pos, IsPublished: in.IsPublished,
	})
	if err != nil {
		return nil, err
	}
	if _, err := u.contents.Upsert(ctx, sm.ID, in.BodyMD); err != nil {
		return nil, err
	}
	return sm, nil
}

func (u *UseCase) UpdateSubmodule(ctx context.Context, id string, in SubmoduleInput) (*domain.Submodule, error) {
	slug := normalizeSlug(in.Slug, in.Title)
	sm, err := u.subs.Update(ctx, domain.Submodule{
		ID: id, ModuleID: in.ModuleID, Slug: slug, Title: in.Title,
		Position: in.Position, IsPublished: in.IsPublished,
	})
	if err != nil {
		return nil, err
	}
	if _, err := u.contents.Upsert(ctx, sm.ID, in.BodyMD); err != nil {
		return nil, err
	}
	return sm, nil
}

func (u *UseCase) DeleteSubmodule(ctx context.Context, id string) error {
	return u.subs.Delete(ctx, id)
}

func (u *UseCase) GetContent(ctx context.Context, submoduleID string) (*domain.SubmoduleContent, error) {
	c, err := u.contents.GetBySubmoduleID(ctx, submoduleID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return &domain.SubmoduleContent{SubmoduleID: submoduleID, BodyMD: ""}, nil
		}
		return nil, err
	}
	return c, nil
}

func (u *UseCase) PutContent(ctx context.Context, submoduleID, bodyMD string) (*domain.SubmoduleContent, error) {
	return u.contents.Upsert(ctx, submoduleID, bodyMD)
}

func normalizeSlug(slug, title string) string {
	s := strings.TrimSpace(slug)
	if s == "" {
		s = title
	}
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	prevDash := false
	for _, r := range s {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
			prevDash = false
		case r == ' ' || r == '_' || r == '-':
			if !prevDash && b.Len() > 0 {
				b.WriteByte('-')
				prevDash = true
			}
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "item"
	}
	return out
}
