package telegram

// Config — настройки Telegram-бота. Префикс: ROADMAP_TELEGRAM_*.
type Config struct {
	// BotToken — токен бота от BotFather.
	BotToken string `envconfig:"BOT_TOKEN"`
	// BotUsername — публичный @username бота без @ (для подсказок в UI).
	BotUsername string `envconfig:"BOT_USERNAME"`
	// WebhookSecret — секрет webhook (X-Telegram-Bot-Api-Secret-Token); нужен в prod.
	WebhookSecret string `envconfig:"WEBHOOK_SECRET"`
	// WebhookURL — полный HTTPS URL webhook, например https://api.example.com/api/v1/telegram/webhook.
	WebhookURL string `envconfig:"WEBHOOK_URL"`
	// PollingTimeout — long-poll timeout в секундах (только dev).
	PollingTimeout int `envconfig:"POLLING_TIMEOUT" default:"30"`
	// FeedbackTZ — IANA TZ для слота авто-ОС (четверг 12:00). Default Europe/Moscow.
	FeedbackTZ string `envconfig:"FEEDBACK_TZ" default:"Europe/Moscow"`
}

// Enabled — бот сконфигурирован (достаточно токена).
func (c Config) Enabled() bool {
	return c.BotToken != ""
}

// WebhookReady — можно ставить webhook (prod).
func (c Config) WebhookReady() bool {
	return c.WebhookURL != "" && c.WebhookSecret != ""
}
