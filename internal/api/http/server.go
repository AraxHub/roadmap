package http

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"roadmap/internal/api/http/middlewares"
)

// ServerConfig — настройки HTTP-сервера. Переменные: ROADMAP_SERVER_HOST, ROADMAP_SERVER_PORT.
type ServerConfig struct {
	Host string `envconfig:"HOST" default:"0.0.0.0"`
	Port string `envconfig:"PORT" default:"8080"`
}

// Controller — контракт: контроллер регистрирует свои маршруты на роутере.
type Controller interface {
	RegisterRoutes(r *gin.Engine)
}

// Server — API-сервер: конфиг и список контроллеров.
type Server struct {
	cfg         ServerConfig
	controllers []Controller
	srv         *http.Server
}

// NewServer создаёт сервер с конфигом.
func NewServer(cfg ServerConfig) *Server {
	return &Server{cfg: cfg, controllers: nil}
}

// AddController добавляет один или несколько контроллеров.
func (s *Server) AddController(c ...Controller) {
	s.controllers = append(s.controllers, c...)
}

// Start поднимает роутер, запускает сервер и блокируется до отмены ctx (SIGINT/SIGTERM), затем делает graceful shutdown.
func (s *Server) Start(ctx context.Context) error {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middlewares.RequestLogger)
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://127.0.0.1:3000", "http://localhost:5173", "http://127.0.0.1:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	for _, c := range s.controllers {
		c.RegisterRoutes(r)
	}

	s.srv = &http.Server{
		Addr:         s.cfg.Host + ":" + s.cfg.Port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		_ = s.srv.ListenAndServe()
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := s.srv.Shutdown(shutdownCtx); err != nil {
		return err
	}
	return nil
}
