package app

import (
	"log"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"roadmap/internal/api/http"
	"roadmap/internal/infrastructure/pg"
	"roadmap/internal/infrastructure/telegram"
	pkgauth "roadmap/internal/pkg/auth"
)

const AppName = "ROADMAP"

// SprintConfig — длительность спринта. ROADMAP_SPRINT_DURATION.
type SprintConfig struct {
	Duration time.Duration `envconfig:"DURATION" default:"168h"`
}

// Config — конфиг приложения. Заполняется через envconfig с префиксом ROADMAP.
type Config struct {
	// Env — окружение: dev | prod. В dev — TG polling, в prod — setWebhook.
	Env      string             `envconfig:"ENV" default:"dev"`
	Server   http.ServerConfig  `envconfig:"SERVER"`
	DB       pg.Config          `envconfig:"DB"`
	Auth     pkgauth.Config     `envconfig:"AUTH"`
	Sprint   SprintConfig       `envconfig:"SPRINT"`
	Telegram telegram.Config    `envconfig:"TELEGRAM"`
}

// IsProd — продакшен-режим.
func (c Config) IsProd() bool {
	return strings.EqualFold(c.Env, "prod")
}

// LoadCfg загружает конфиг: подтягивает .env (godotenv), затем заполняет структуру из окружения (envconfig).
func LoadCfg() (Config, error) {
	if err := godotenv.Load("deployment/roadmap-local/.env"); err != nil {
		log.Printf("config: .env не найден, используем окружение: %v", err)
	}

	var cfg Config
	if err := envconfig.Process(AppName, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
