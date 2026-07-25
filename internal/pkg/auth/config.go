package auth

import (
	"time"
)

// Config — настройки JWT и cookie. Переменные: ROADMAP_AUTH_*.
type Config struct {
	JWTSecret          string        `envconfig:"JWT_SECRET" required:"true"`
	AccessTTL          time.Duration `envconfig:"ACCESS_TTL" default:"15m"`
	RefreshTTL         time.Duration `envconfig:"REFRESH_TTL" default:"72h"`
	RefreshCookieName  string        `envconfig:"REFRESH_COOKIE_NAME" default:"refresh_token"`
	RefreshCookieSecure bool         `envconfig:"REFRESH_COOKIE_SECURE" default:"false"`
	RefreshCookiePath  string        `envconfig:"REFRESH_COOKIE_PATH" default:"/api/v1/auth"`
	BootstrapLogin     string        `envconfig:"BOOTSTRAP_LOGIN"`
	BootstrapPassword  string        `envconfig:"BOOTSTRAP_PASSWORD"`
}
