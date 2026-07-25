package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	apihttp "roadmap/internal/api/http"
	adminctrl "roadmap/internal/api/http/controllers/admin"
	authctrl "roadmap/internal/api/http/controllers/auth"
	roadmapctrl "roadmap/internal/api/http/controllers/roadmap"
	"roadmap/internal/api/http/controllers/system"
	tgctrl "roadmap/internal/api/http/controllers/telegram"
	"roadmap/internal/api/http/middlewares"
	infrapg "roadmap/internal/infrastructure/pg"
	infrag "roadmap/internal/infrastructure/telegram"
	pkgauth "roadmap/internal/pkg/auth"
	"roadmap/internal/pkg/logger"
	"roadmap/internal/pkg/migrate"
	repopg "roadmap/internal/repository/pg"
	adminuc "roadmap/internal/usecase/admin"
	authuc "roadmap/internal/usecase/auth"
	feedbackuc "roadmap/internal/usecase/feedback"
	roadmapuc "roadmap/internal/usecase/roadmap"
)

// App — приложение, хранит только конфиг.
type App struct {
	cfg Config
}

// New создаёт приложение с конфигом.
func New(cfg Config) *App {
	return &App{cfg: cfg}
}

// Run собирает зависимости и запускает HTTP-сервер.
func (a *App) Run() error {
	db, err := infrapg.New(&a.cfg.DB)
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}
	defer db.Close()

	if err := migrate.Up(a.cfg.DB.URL()); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	log := logger.New()
	slog.SetDefault(log)

	repo := repopg.NewRepo(db, log)
	tokens := pkgauth.NewTokenService(a.cfg.Auth.JWTSecret, a.cfg.Auth.AccessTTL, a.cfg.Auth.RefreshTTL)

	authUC := authuc.New(repo.Users, repo.RefreshTokens, tokens, log)
	if err := authUC.BootstrapAdmin(context.Background(), a.cfg.Auth.BootstrapLogin, a.cfg.Auth.BootstrapPassword); err != nil {
		return fmt.Errorf("bootstrap admin: %w", err)
	}

	roadmapUC := roadmapuc.New(
		repo.Sprints,
		repo.Modules,
		repo.Submodules,
		repo.SubmoduleContents,
		repo.Users,
		repo.UserSubmoduleProgress,
		repo.UserSprintTimers,
		a.cfg.Sprint.Duration,
		a.cfg.Telegram.BotUsername,
		log,
	)
	catalogUC := adminuc.New(repo.Sprints, repo.Modules, repo.Submodules, repo.SubmoduleContents, log)

	var tgClient *infrag.Client
	var messenger feedbackuc.Messenger
	if a.cfg.Telegram.Enabled() {
		tgClient = infrag.NewClient(a.cfg.Telegram.BotToken, log)
		messenger = tgClient
	}
	feedbackLoc := loadFeedbackLocation(a.cfg.Telegram.FeedbackTZ, log)
	feedbackUC := feedbackuc.New(repo.Users, repo.FeedbackRequests, repo.FeedbackSchedule, messenger, feedbackLoc, log)

	authMW := middlewares.Auth(tokens)
	adminMW := middlewares.RequireAdmin()

	srv := apihttp.NewServer(a.cfg.Server)
	controllers := []apihttp.Controller{
		system.New(db, log),
		authctrl.New(authUC, a.cfg.Auth, authMW, log),
		roadmapctrl.New(roadmapUC, authMW, log),
		adminctrl.New(authUC, catalogUC, feedbackUC, authMW, adminMW, log),
	}

	// Webhook HTTP endpoint нужен в prod; в dev тоже регистрируем — не мешает polling.
	if a.cfg.Telegram.Enabled() && a.cfg.Telegram.WebhookSecret != "" {
		controllers = append(controllers, tgctrl.New(feedbackUC, a.cfg.Telegram.WebhookSecret, log))
	}
	srv.AddController(controllers...)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if a.cfg.Telegram.Enabled() {
		go feedbackuc.NewWorker(feedbackUC, repo.FeedbackSchedule, feedbackLoc, log).Run(ctx)
		a.startTelegramIngress(ctx, tgClient, feedbackUC, log)
	}

	mode := "dev"
	if a.cfg.IsProd() {
		mode = "prod"
	}
	addr := a.cfg.Server.Host + ":" + a.cfg.Server.Port
	slog.Info("application started",
		"addr", addr,
		"env", mode,
		"sprint_duration", a.cfg.Sprint.Duration.String(),
		"telegram", a.cfg.Telegram.Enabled(),
		"telegram_bot", a.cfg.Telegram.BotUsername,
	)
	return srv.Start(ctx)
}

func (a *App) startTelegramIngress(
	ctx context.Context,
	client *infrag.Client,
	feedbackUC *feedbackuc.UseCase,
	log *slog.Logger,
) {
	if client == nil {
		return
	}

	if a.cfg.IsProd() {
		if !a.cfg.Telegram.WebhookReady() {
			log.Error("telegram prod: WEBHOOK_URL and WEBHOOK_SECRET required")
			return
		}
		go func() {
			if err := client.SetWebhook(ctx, a.cfg.Telegram.WebhookURL, a.cfg.Telegram.WebhookSecret); err != nil {
				log.Error("telegram setWebhook failed", "error", err)
				return
			}
		}()
		return
	}

	// dev: снимаем webhook и слушаем polling
	go func() {
		if err := client.DeleteWebhook(ctx); err != nil {
			log.Warn("telegram deleteWebhook failed", "error", err)
		}
		poller := infrag.NewPoller(
			client,
			a.cfg.Telegram.PollingTimeout,
			func(ctx context.Context, update *infrag.Update) error {
				return tgctrl.DispatchUpdate(ctx, feedbackUC, update)
			},
			log,
		)
		poller.Start(ctx)
	}()
}

// loadFeedbackLocation грузит TZ слота авто-ОС с fallback на Europe/Moscow → фиксированный MSK.
func loadFeedbackLocation(tz string, log *slog.Logger) *time.Location {
	if loc, err := time.LoadLocation(tz); err == nil {
		return loc
	} else {
		log.Warn("invalid FEEDBACK_TZ, fallback Europe/Moscow", "tz", tz, "error", err)
	}
	if loc, err := time.LoadLocation("Europe/Moscow"); err == nil {
		return loc
	}
	return time.FixedZone("MSK", 3*60*60)
}
