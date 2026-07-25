package auth

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"roadmap/internal/api/http/middlewares"
	"roadmap/internal/domain"
	pkgauth "roadmap/internal/pkg/auth"
	authuc "roadmap/internal/usecase/auth"
)

// UseCase — контракт auth-контроллера.
type UseCase interface {
	Login(ctx context.Context, login, password string) (*authuc.TokenPair, error)
	Refresh(ctx context.Context, rawRefresh string) (*authuc.TokenPair, error)
	Logout(ctx context.Context, rawRefresh string) error
	Me(ctx context.Context, userID string) (*domain.User, error)
}

// Controller — login/refresh/logout/me.
type Controller struct {
	uc      UseCase
	authCfg pkgauth.Config
	authMW  gin.HandlerFunc
	log     *slog.Logger
}

// New создаёт auth-контроллер.
func New(uc UseCase, authCfg pkgauth.Config, authMW gin.HandlerFunc, log *slog.Logger) *Controller {
	return &Controller{uc: uc, authCfg: authCfg, authMW: authMW, log: log}
}

// RegisterRoutes регистрирует публичные и защищённые auth-роуты.
func (c *Controller) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/v1/auth")
	api.POST("/login", c.login)
	api.POST("/refresh", c.refresh)

	protected := api.Group("")
	protected.Use(c.authMW)
	protected.POST("/logout", c.logout)
	protected.GET("/me", c.me)
}

type loginRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (c *Controller) login(ctx *gin.Context) {
	var req loginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	pair, err := c.uc.Login(ctx.Request.Context(), req.Login, req.Password)
	if err != nil {
		c.writeErr(ctx, err)
		return
	}
	c.setRefreshCookie(ctx, pair.RefreshToken, pair.RefreshExpiresAt)
	ctx.JSON(http.StatusOK, gin.H{
		"access_token": pair.AccessToken,
		"expires_at":   pair.AccessExpiresAt,
		"user": gin.H{
			"id":    pair.UserID,
			"login": pair.Login,
			"role":  pair.Role,
		},
	})
}

func (c *Controller) refresh(ctx *gin.Context) {
	raw, _ := ctx.Cookie(c.authCfg.RefreshCookieName)
	pair, err := c.uc.Refresh(ctx.Request.Context(), raw)
	if err != nil {
		c.clearRefreshCookie(ctx)
		c.writeErr(ctx, err)
		return
	}
	c.setRefreshCookie(ctx, pair.RefreshToken, pair.RefreshExpiresAt)
	ctx.JSON(http.StatusOK, gin.H{
		"access_token": pair.AccessToken,
		"expires_at":   pair.AccessExpiresAt,
		"user": gin.H{
			"id":    pair.UserID,
			"login": pair.Login,
			"role":  pair.Role,
		},
	})
}

func (c *Controller) logout(ctx *gin.Context) {
	raw, _ := ctx.Cookie(c.authCfg.RefreshCookieName)
	_ = c.uc.Logout(ctx.Request.Context(), raw)
	c.clearRefreshCookie(ctx)
	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (c *Controller) me(ctx *gin.Context) {
	userID, ok := middlewares.UserID(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": domain.ErrUnauthorized.Error()})
		return
	}
	user, err := c.uc.Me(ctx.Request.Context(), userID)
	if err != nil {
		c.writeErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"id":    user.ID,
		"login": user.Login,
		"role":  user.Role,
	})
}

func (c *Controller) setRefreshCookie(ctx *gin.Context, raw string, exp time.Time) {
	maxAge := int(time.Until(exp).Seconds())
	if maxAge < 0 {
		maxAge = 0
	}
	ctx.SetSameSite(http.SameSiteLaxMode)
	ctx.SetCookie(
		c.authCfg.RefreshCookieName,
		raw,
		maxAge,
		c.authCfg.RefreshCookiePath,
		"",
		c.authCfg.RefreshCookieSecure,
		true,
	)
}

func (c *Controller) clearRefreshCookie(ctx *gin.Context) {
	ctx.SetSameSite(http.SameSiteLaxMode)
	ctx.SetCookie(c.authCfg.RefreshCookieName, "", -1, c.authCfg.RefreshCookiePath, "", c.authCfg.RefreshCookieSecure, true)
}

func (c *Controller) writeErr(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidCredentials),
		errors.Is(err, domain.ErrInvalidToken),
		errors.Is(err, domain.ErrUnauthorized):
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrNotFound):
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	default:
		c.log.Error("auth failed", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	}
}
