package telegram

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	infratg "roadmap/internal/infrastructure/telegram"
)

// FeedbackUC — обработка входящих сообщений бота.
type FeedbackUC interface {
	HandleUpdate(ctx context.Context, chatID int64, text string) error
}

// Controller — HTTP webhook Telegram (prod).
type Controller struct {
	uc            FeedbackUC
	webhookSecret string
	log           *slog.Logger
}

// New создаёт контроллер. secret — ROADMAP_TELEGRAM_WEBHOOK_SECRET.
func New(uc FeedbackUC, webhookSecret string, log *slog.Logger) *Controller {
	return &Controller{uc: uc, webhookSecret: webhookSecret, log: log}
}

// RegisterRoutes регистрирует POST /api/v1/telegram/webhook.
func (c *Controller) RegisterRoutes(r *gin.Engine) {
	r.POST("/api/v1/telegram/webhook", c.handleWebhook)
}

func (c *Controller) handleWebhook(ctx *gin.Context) {
	secret := ctx.GetHeader("X-Telegram-Bot-Api-Secret-Token")
	if secret == "" || secret != c.webhookSecret {
		c.log.Error("telegram webhook: bad secret token")
		ctx.JSON(http.StatusOK, gin.H{"ok": false, "error": "unauthorized"})
		return
	}

	var update infratg.Update
	if err := ctx.ShouldBindJSON(&update); err != nil {
		c.log.Error("telegram webhook: invalid body", "error", err)
		ctx.JSON(http.StatusOK, gin.H{"ok": false, "error": "invalid request"})
		return
	}

	if err := dispatchUpdate(ctx.Request.Context(), c.uc, &update); err != nil {
		c.log.Error("telegram webhook: handle failed", "error", err)
		ctx.JSON(http.StatusOK, gin.H{"ok": false, "error": "failed to process update"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}

// DispatchUpdate — общая обработка update для webhook и polling.
func DispatchUpdate(ctx context.Context, uc FeedbackUC, update *infratg.Update) error {
	return dispatchUpdate(ctx, uc, update)
}

func dispatchUpdate(ctx context.Context, uc FeedbackUC, update *infratg.Update) error {
	if update == nil || update.Message == nil || update.Message.Text == "" {
		return nil
	}
	return uc.HandleUpdate(ctx, update.Message.Chat.ID, update.Message.Text)
}
