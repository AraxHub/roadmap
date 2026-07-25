package system

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthStore — минимальный контракт для readiness.
type HealthStore interface {
	Ping(ctx context.Context) error
}

// Controller — системные маршруты.
type Controller struct {
	store HealthStore
	log   *slog.Logger
}

// New создаёт системный контроллер.
func New(store HealthStore, log *slog.Logger) *Controller {
	return &Controller{store: store, log: log}
}

// RegisterRoutes регистрирует маршруты.
func (c *Controller) RegisterRoutes(r *gin.Engine) {
	r.GET("/liveness", c.live)
	r.GET("/readyness", c.ready)
}

func (c *Controller) live(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"status": "alive"})
}

func (c *Controller) ready(ctx *gin.Context) {
	if err := c.store.Ping(ctx.Request.Context()); err != nil {
		c.log.Warn("ready check failed", "error", err)
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready", "error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "ready"})
}
