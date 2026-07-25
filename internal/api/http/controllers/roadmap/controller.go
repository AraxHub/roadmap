package roadmap

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"roadmap/internal/api/http/middlewares"
	"roadmap/internal/domain"
	roadmapuc "roadmap/internal/usecase/roadmap"
)

// UseCase — контракт для HTTP-контроллера.
type UseCase interface {
	Home(ctx context.Context, userID string) (*roadmapuc.HomeView, error)
	ModulePage(ctx context.Context, userID, moduleSlug string) (*roadmapuc.ModuleView, error)
	SubmodulePage(ctx context.Context, userID, moduleSlug, submoduleSlug string) (*roadmapuc.SubmoduleView, error)
	Complete(ctx context.Context, userID, submoduleID string) error
}

// Controller — HTTP API роадмапа.
type Controller struct {
	uc     UseCase
	authMW gin.HandlerFunc
	log    *slog.Logger
}

// New создаёт контроллер.
func New(uc UseCase, authMW gin.HandlerFunc, log *slog.Logger) *Controller {
	return &Controller{uc: uc, authMW: authMW, log: log}
}

// RegisterRoutes регистрирует защищённые маршруты.
func (c *Controller) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")
	api.Use(c.authMW)
	api.GET("/home", c.home)
	api.GET("/modules/:moduleSlug", c.modulePage)
	api.GET("/modules/:moduleSlug/submodules/:submoduleSlug", c.submodulePage)
	api.POST("/submodules/:submoduleID/complete", c.complete)
}

func (c *Controller) home(ctx *gin.Context) {
	userID, ok := middlewares.UserID(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": domain.ErrUnauthorized.Error()})
		return
	}
	view, err := c.uc.Home(ctx.Request.Context(), userID)
	if err != nil {
		c.writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, view)
}

func (c *Controller) modulePage(ctx *gin.Context) {
	userID, ok := middlewares.UserID(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": domain.ErrUnauthorized.Error()})
		return
	}
	view, err := c.uc.ModulePage(ctx.Request.Context(), userID, ctx.Param("moduleSlug"))
	if err != nil {
		c.writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, view)
}

func (c *Controller) submodulePage(ctx *gin.Context) {
	userID, ok := middlewares.UserID(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": domain.ErrUnauthorized.Error()})
		return
	}
	view, err := c.uc.SubmodulePage(
		ctx.Request.Context(),
		userID,
		ctx.Param("moduleSlug"),
		ctx.Param("submoduleSlug"),
	)
	if err != nil {
		c.writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, view)
}

func (c *Controller) complete(ctx *gin.Context) {
	userID, ok := middlewares.UserID(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": domain.ErrUnauthorized.Error()})
		return
	}
	if err := c.uc.Complete(ctx.Request.Context(), userID, ctx.Param("submoduleID")); err != nil {
		c.writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "completed"})
}

func (c *Controller) writeErr(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidUser), errors.Is(err, domain.ErrUnauthorized):
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrNotFound):
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrLocked):
		ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	default:
		c.log.Error("request failed", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	}
}
