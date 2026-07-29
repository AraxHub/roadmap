package admin

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"
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
	Upsert(ctx context.Context, submoduleID string, blocks []domain.ContentBlock) (*domain.SubmoduleContent, error)
}

type ImageStore interface {
	Create(ctx context.Context, submoduleID, mimeType string, data []byte) (*domain.ContentImage, error)
	GetByID(ctx context.Context, id string) (*domain.ContentImage, error)
	ExistsForSubmodule(ctx context.Context, imageID, submoduleID string) (bool, error)
	Delete(ctx context.Context, id string) error
}

// UseCase — админ-конфиг роадмапа.
type UseCase struct {
	sprints  SprintStore
	modules  ModuleStore
	subs     SubmoduleStore
	contents ContentStore
	images   ImageStore
	log      *slog.Logger
}

func New(
	sprints SprintStore,
	modules ModuleStore,
	subs SubmoduleStore,
	contents ContentStore,
	images ImageStore,
	log *slog.Logger,
) *UseCase {
	return &UseCase{
		sprints: sprints, modules: modules, subs: subs,
		contents: contents, images: images, log: log,
	}
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
	ModuleID    string                `json:"module_id"`
	Slug        string                `json:"slug"`
	Title       string                `json:"title"`
	Position    int                   `json:"position"`
	IsPublished bool                  `json:"is_published"`
	Blocks      []domain.ContentBlock `json:"blocks"`
	BodyMD      string                `json:"body_md"` // legacy: один markdown-блок при создании
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
	blocks := in.Blocks
	if len(blocks) == 0 && strings.TrimSpace(in.BodyMD) != "" {
		blocks = []domain.ContentBlock{{
			ID:   newBlockID(),
			Type: domain.BlockTypeMarkdown,
			MD:   in.BodyMD,
		}}
	}
	normalized, err := u.normalizeBlocks(ctx, sm.ID, blocks)
	if err != nil {
		return nil, err
	}
	if _, err := u.contents.Upsert(ctx, sm.ID, normalized); err != nil {
		return nil, err
	}
	return sm, nil
}

func (u *UseCase) UpdateSubmodule(ctx context.Context, id string, in SubmoduleInput) (*domain.Submodule, error) {
	slug := normalizeSlug(in.Slug, in.Title)
	return u.subs.Update(ctx, domain.Submodule{
		ID: id, ModuleID: in.ModuleID, Slug: slug, Title: in.Title,
		Position: in.Position, IsPublished: in.IsPublished,
	})
}

func (u *UseCase) DeleteSubmodule(ctx context.Context, id string) error {
	return u.subs.Delete(ctx, id)
}

func (u *UseCase) GetContent(ctx context.Context, submoduleID string) (*domain.SubmoduleContent, error) {
	c, err := u.contents.GetBySubmoduleID(ctx, submoduleID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return &domain.SubmoduleContent{SubmoduleID: submoduleID, Blocks: []domain.ContentBlock{}}, nil
		}
		return nil, err
	}
	return c, nil
}

func (u *UseCase) PutContent(ctx context.Context, submoduleID string, blocks []domain.ContentBlock) (*domain.SubmoduleContent, error) {
	normalized, err := u.normalizeBlocks(ctx, submoduleID, blocks)
	if err != nil {
		return nil, err
	}
	return u.contents.Upsert(ctx, submoduleID, normalized)
}

// PutMarkdownContent — загрузка .md файла как одного markdown-блока.
func (u *UseCase) PutMarkdownContent(ctx context.Context, submoduleID, bodyMD string) (*domain.SubmoduleContent, error) {
	blocks := []domain.ContentBlock{}
	if strings.TrimSpace(bodyMD) != "" {
		blocks = []domain.ContentBlock{{
			ID:   newBlockID(),
			Type: domain.BlockTypeMarkdown,
			MD:   bodyMD,
		}}
	}
	return u.PutContent(ctx, submoduleID, blocks)
}

type UploadedImage struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

func (u *UseCase) UploadImage(ctx context.Context, submoduleID, mimeType string, data []byte) (*UploadedImage, error) {
	mimeType = strings.TrimSpace(strings.ToLower(mimeType))
	if mimeType == "image/jpg" {
		mimeType = "image/jpeg"
	}
	if !domain.AllowedContentImageMIME[mimeType] {
		return nil, fmt.Errorf("%w: unsupported image type", domain.ErrInvalidInput)
	}
	if len(data) == 0 || len(data) > domain.MaxContentImageBytes {
		return nil, fmt.Errorf("%w: image size must be 1..%d bytes", domain.ErrInvalidInput, domain.MaxContentImageBytes)
	}
	img, err := u.images.Create(ctx, submoduleID, mimeType, data)
	if err != nil {
		return nil, err
	}
	return &UploadedImage{
		ID:  img.ID,
		URL: ContentImageURL(img.ID),
	}, nil
}

func (u *UseCase) DeleteImage(ctx context.Context, imageID string) error {
	return u.images.Delete(ctx, imageID)
}

func (u *UseCase) GetImage(ctx context.Context, imageID string) (*domain.ContentImage, error) {
	return u.images.GetByID(ctx, imageID)
}

// ContentImageURL — путь API для отдачи картинки.
func ContentImageURL(imageID string) string {
	return "/api/v1/content-images/" + imageID
}

func (u *UseCase) normalizeBlocks(ctx context.Context, submoduleID string, blocks []domain.ContentBlock) ([]domain.ContentBlock, error) {
	if blocks == nil {
		blocks = []domain.ContentBlock{}
	}
	out := make([]domain.ContentBlock, 0, len(blocks))
	for _, b := range blocks {
		id := strings.TrimSpace(b.ID)
		if id == "" {
			id = newBlockID()
		}
		switch b.Type {
		case domain.BlockTypeMarkdown:
			out = append(out, domain.ContentBlock{
				ID:   id,
				Type: b.Type,
				MD:   b.MD,
			})
		case domain.BlockTypeAnswer:
			title := strings.TrimSpace(b.Title)
			out = append(out, domain.ContentBlock{
				ID:    id,
				Type:  b.Type,
				MD:    b.MD,
				Title: title,
			})
		case domain.BlockTypeImage:
			imageID := strings.TrimSpace(b.ImageID)
			if imageID == "" {
				return nil, fmt.Errorf("%w: image block requires image_id", domain.ErrInvalidInput)
			}
			ok, err := u.images.ExistsForSubmodule(ctx, imageID, submoduleID)
			if err != nil {
				return nil, err
			}
			if !ok {
				return nil, fmt.Errorf("%w: image not found for submodule", domain.ErrInvalidInput)
			}
			out = append(out, domain.ContentBlock{
				ID:      id,
				Type:    domain.BlockTypeImage,
				ImageID: imageID,
				Alt:     b.Alt,
			})
		default:
			return nil, fmt.Errorf("%w: unknown block type %q", domain.ErrInvalidInput, b.Type)
		}
	}
	return out, nil
}

func newBlockID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("blk-%d", time.Now().UnixNano())
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	hexed := hex.EncodeToString(b[:])
	return hexed[0:8] + "-" + hexed[8:12] + "-" + hexed[12:16] + "-" + hexed[16:20] + "-" + hexed[20:32]
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
