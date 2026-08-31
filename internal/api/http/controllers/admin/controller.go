package admin

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"roadmap/internal/domain"
	adminuc "roadmap/internal/usecase/admin"
	authuc "roadmap/internal/usecase/auth"
	feedbackuc "roadmap/internal/usecase/feedback"

	"github.com/gin-gonic/gin"
)

// UsersUC — создание ЛК.
type UsersUC interface {
	CreateUser(ctx context.Context, login, role string) (*authuc.CreatedUser, error)
}

// CatalogUC — конфиг роадмапа.
type CatalogUC interface {
	Tree(ctx context.Context) (*adminuc.Tree, error)
	CreateSprint(ctx context.Context, in adminuc.SprintInput) (*domain.Sprint, error)
	UpdateSprint(ctx context.Context, id string, in adminuc.SprintInput) (*domain.Sprint, error)
	DeleteSprint(ctx context.Context, id string) error
	CreateModule(ctx context.Context, in adminuc.ModuleInput) (*domain.Module, error)
	UpdateModule(ctx context.Context, id string, in adminuc.ModuleInput) (*domain.Module, error)
	DeleteModule(ctx context.Context, id string) error
	CreateSubmodule(ctx context.Context, in adminuc.SubmoduleInput) (*domain.Submodule, error)
	UpdateSubmodule(ctx context.Context, id string, in adminuc.SubmoduleInput) (*domain.Submodule, error)
	DeleteSubmodule(ctx context.Context, id string) error
	GetContent(ctx context.Context, submoduleID string) (*domain.SubmoduleContent, error)
	PutContent(ctx context.Context, submoduleID string, blocks []domain.ContentBlock) (*domain.SubmoduleContent, error)
	PutMarkdownContent(ctx context.Context, submoduleID, bodyMD string) (*domain.SubmoduleContent, error)
	UploadImage(ctx context.Context, submoduleID, mimeType string, data []byte) (*adminuc.UploadedImage, error)
	DeleteImage(ctx context.Context, imageID string) error
	GetImage(ctx context.Context, imageID string) (*domain.ContentImage, error)
}

// FeedbackUC — список юзеров и история ОС.
type FeedbackUC interface {
	ListUsers(ctx context.Context) ([]feedbackuc.UserListItem, error)
	ListFeedback(ctx context.Context, userID string) ([]feedbackuc.FeedbackItem, error)
	RequestRound(ctx context.Context) error
	RequestForUser(ctx context.Context, userID string) error
	Schedule(ctx context.Context) (*feedbackuc.ScheduleInfo, error)
	UpdateMessage(ctx context.Context, text string) error
}

// Controller — admin API.
type Controller struct {
	users    UsersUC
	catalog  CatalogUC
	feedback FeedbackUC
	authMW   gin.HandlerFunc
	adminMW  gin.HandlerFunc
	log      *slog.Logger
}

// New создаёт admin-контроллер.
func New(users UsersUC, catalog CatalogUC, feedback FeedbackUC, authMW, adminMW gin.HandlerFunc, log *slog.Logger) *Controller {
	return &Controller{users: users, catalog: catalog, feedback: feedback, authMW: authMW, adminMW: adminMW, log: log}
}

// RegisterRoutes регистрирует /api/v1/admin/* под auth + admin.
func (c *Controller) RegisterRoutes(r *gin.Engine) {
	admin := r.Group("/api/v1/admin")
	admin.Use(c.authMW, c.adminMW)

	admin.POST("/users", c.createUser)
	admin.GET("/users", c.listUsers)
	admin.GET("/users/:id/feedback", c.listFeedback)
	admin.POST("/users/:id/feedback/request", c.requestForUser)
	admin.POST("/feedback/request-round", c.requestRound)
	admin.GET("/feedback/schedule", c.feedbackSchedule)
	admin.PUT("/feedback/message", c.updateFeedbackMessage)

	admin.GET("/tree", c.tree)

	admin.POST("/sprints", c.createSprint)
	admin.PUT("/sprints/:id", c.updateSprint)
	admin.DELETE("/sprints/:id", c.deleteSprint)

	admin.POST("/modules", c.createModule)
	admin.PUT("/modules/:id", c.updateModule)
	admin.DELETE("/modules/:id", c.deleteModule)

	admin.POST("/submodules", c.createSubmodule)
	admin.PUT("/submodules/:id", c.updateSubmodule)
	admin.DELETE("/submodules/:id", c.deleteSubmodule)
	admin.GET("/submodules/:id/content", c.getContent)
	admin.PUT("/submodules/:id/content", c.putContent)
	admin.POST("/submodules/:id/content/upload", c.uploadContent)
	admin.POST("/submodules/:id/images", c.uploadImage)
	admin.DELETE("/content-images/:id", c.deleteImage)
}

type createUserRequest struct {
	Login string `json:"login" binding:"required"`
	Role  string `json:"role"`
}

func (c *Controller) createUser(ctx *gin.Context) {
	var req createUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	created, err := c.users.CreateUser(ctx.Request.Context(), req.Login, req.Role)
	if err != nil {
		c.writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, created)
}

func (c *Controller) listUsers(ctx *gin.Context) {
	users, err := c.feedback.ListUsers(ctx.Request.Context())
	if err != nil {
		c.writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"users": users})
}

func (c *Controller) listFeedback(ctx *gin.Context) {
	items, err := c.feedback.ListFeedback(ctx.Request.Context(), ctx.Param("id"))
	if err != nil {
		c.writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"feedback": items})
}

func (c *Controller) requestForUser(ctx *gin.Context) {
	if err := c.feedback.RequestForUser(ctx.Request.Context(), ctx.Param("id")); err != nil {
		c.writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (c *Controller) requestRound(ctx *gin.Context) {
	if err := c.feedback.RequestRound(ctx.Request.Context()); err != nil {
		c.writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (c *Controller) feedbackSchedule(ctx *gin.Context) {
	info, err := c.feedback.Schedule(ctx.Request.Context())
	if err != nil {
		c.writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, info)
}

func (c *Controller) updateFeedbackMessage(ctx *gin.Context) {
	var body struct {
		MessageText string `json:"message_text"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if err := c.feedback.UpdateMessage(ctx.Request.Context(), body.MessageText); err != nil {
		if errors.Is(err, domain.ErrInvalidUser) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "message_text required"})
			return
		}
		c.writeErr(ctx, err)
		return
	}
	info, err := c.feedback.Schedule(ctx.Request.Context())
	if err != nil {
		c.writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, info)
}

func (c *Controller) tree(ctx *gin.Context) {

	ctx.JSON(http.StatusInternalServerError, gin.H{"status": "поломка"})
	return

	tree, err := c.catalog.Tree(ctx.Request.Context())
	if err != nil {
		c.writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, tree)
}

func (c *Controller) createSprint(ctx *gin.Context) {
	var in adminuc.SprintInput
	if err := ctx.ShouldBindJSON(&in); err != nil || in.Title == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	s, err := c.catalog.CreateSprint(ctx.Request.Context(), in)
	if err != nil {
		c.writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, s)
}

func (c *Controller) updateSprint(ctx *gin.Context) {
	var in adminuc.SprintInput
	if err := ctx.ShouldBindJSON(&in); err != nil || in.Title == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	s, err := c.catalog.UpdateSprint(ctx.Request.Context(), ctx.Param("id"), in)
	if err != nil {
		c.writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, s)
}

func (c *Controller) deleteSprint(ctx *gin.Context) {
	if err := c.catalog.DeleteSprint(ctx.Request.Context(), ctx.Param("id")); err != nil {
		c.writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (c *Controller) createModule(ctx *gin.Context) {
	var in adminuc.ModuleInput
	if err := ctx.ShouldBindJSON(&in); err != nil || in.Title == "" || in.SprintID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	m, err := c.catalog.CreateModule(ctx.Request.Context(), in)
	if err != nil {
		c.writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, m)
}

func (c *Controller) updateModule(ctx *gin.Context) {
	var in adminuc.ModuleInput
	if err := ctx.ShouldBindJSON(&in); err != nil || in.Title == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	m, err := c.catalog.UpdateModule(ctx.Request.Context(), ctx.Param("id"), in)
	if err != nil {
		c.writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, m)
}

func (c *Controller) deleteModule(ctx *gin.Context) {
	if err := c.catalog.DeleteModule(ctx.Request.Context(), ctx.Param("id")); err != nil {
		c.writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (c *Controller) createSubmodule(ctx *gin.Context) {
	var in adminuc.SubmoduleInput
	if err := ctx.ShouldBindJSON(&in); err != nil || in.Title == "" || in.ModuleID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	sm, err := c.catalog.CreateSubmodule(ctx.Request.Context(), in)
	if err != nil {
		c.writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, sm)
}

func (c *Controller) updateSubmodule(ctx *gin.Context) {
	var in adminuc.SubmoduleInput
	if err := ctx.ShouldBindJSON(&in); err != nil || in.Title == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	sm, err := c.catalog.UpdateSubmodule(ctx.Request.Context(), ctx.Param("id"), in)
	if err != nil {
		c.writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, sm)
}

func (c *Controller) deleteSubmodule(ctx *gin.Context) {
	if err := c.catalog.DeleteSubmodule(ctx.Request.Context(), ctx.Param("id")); err != nil {
		c.writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (c *Controller) getContent(ctx *gin.Context) {
	content, err := c.catalog.GetContent(ctx.Request.Context(), ctx.Param("id"))
	if err != nil {
		c.writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, content)
}

func (c *Controller) putContent(ctx *gin.Context) {
	var body struct {
		Blocks []domain.ContentBlock `json:"blocks"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	content, err := c.catalog.PutContent(ctx.Request.Context(), ctx.Param("id"), body.Blocks)
	if err != nil {
		c.writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, content)
}

func (c *Controller) uploadContent(ctx *gin.Context) {
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "file required"})
		return
	}
	f, err := file.Open()
	if err != nil {
		c.writeErr(ctx, err)
		return
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		c.writeErr(ctx, err)
		return
	}
	content, err := c.catalog.PutMarkdownContent(ctx.Request.Context(), ctx.Param("id"), string(data))
	if err != nil {
		c.writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, content)
}

func (c *Controller) uploadImage(ctx *gin.Context) {
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "file required"})
		return
	}
	if file.Size > domain.MaxContentImageBytes {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "image too large (max 10MB)"})
		return
	}
	f, err := file.Open()
	if err != nil {
		c.writeErr(ctx, err)
		return
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, domain.MaxContentImageBytes+1))
	if err != nil {
		c.writeErr(ctx, err)
		return
	}
	if len(data) > domain.MaxContentImageBytes {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "image too large (max 10MB)"})
		return
	}
	mime := file.Header.Get("Content-Type")
	if mime == "" || mime == "application/octet-stream" {
		mime = http.DetectContentType(data)
	}
	uploaded, err := c.catalog.UploadImage(ctx.Request.Context(), ctx.Param("id"), mime, data)
	if err != nil {
		c.writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, uploaded)
}

func (c *Controller) deleteImage(ctx *gin.Context) {
	if err := c.catalog.DeleteImage(ctx.Request.Context(), ctx.Param("id")); err != nil {
		c.writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (c *Controller) writeErr(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrLoginTaken), errors.Is(err, domain.ErrConflict):
		ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrNotFound):
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrForbidden):
		ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrInvalidInput):
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.log.Error("admin failed", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	}
}
