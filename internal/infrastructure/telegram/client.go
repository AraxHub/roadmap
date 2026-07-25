package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

const (
	apiBaseURL = "https://api.telegram.org/bot"
	apiTimeout = 30 * time.Second
)

// Client — клиент Telegram Bot API.
type Client struct {
	httpClient *http.Client
	baseURL    string
	log        *slog.Logger
}

// NewClient создаёт клиент.
func NewClient(token string, log *slog.Logger) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: apiTimeout},
		baseURL:    apiBaseURL + token,
		log:        log,
	}
}

type sendMessageRequest struct {
	ChatID    int64  `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode,omitempty"`
}

type apiResponse struct {
	OK          bool   `json:"ok"`
	ErrorCode   int    `json:"error_code"`
	Description string `json:"description"`
}

type sendMessageResponse struct {
	apiResponse
	Result struct {
		MessageID int64 `json:"message_id"`
	} `json:"result"`
}

type setWebhookRequest struct {
	URL         string `json:"url"`
	SecretToken string `json:"secret_token,omitempty"`
}

type getUpdatesResponse struct {
	apiResponse
	Result []Update `json:"result"`
}

// SendMessage отправляет текст и возвращает message_id.
func (c *Client) SendMessage(ctx context.Context, chatID int64, text string) (int64, error) {
	body, err := json.Marshal(sendMessageRequest{ChatID: chatID, Text: text})
	if err != nil {
		return 0, fmt.Errorf("telegram marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/sendMessage", bytes.NewReader(body))
	if err != nil {
		return 0, fmt.Errorf("telegram request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("telegram do: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("telegram read: %w", err)
	}

	var out sendMessageResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return 0, fmt.Errorf("telegram unmarshal: %w", err)
	}
	if !out.OK {
		c.log.Debug("telegram API error", "code", out.ErrorCode, "desc", out.Description, "chat_id", chatID)
		return 0, fmt.Errorf("telegram API error [%d]: %s", out.ErrorCode, out.Description)
	}
	return out.Result.MessageID, nil
}

// SetWebhook регистрирует webhook у Telegram.
func (c *Client) SetWebhook(ctx context.Context, url, secretToken string) error {
	body, err := json.Marshal(setWebhookRequest{URL: url, SecretToken: secretToken})
	if err != nil {
		return fmt.Errorf("telegram marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/setWebhook", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("telegram request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("telegram do: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("telegram read: %w", err)
	}

	var out apiResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return fmt.Errorf("telegram unmarshal: %w", err)
	}
	if !out.OK {
		return fmt.Errorf("telegram setWebhook error [%d]: %s", out.ErrorCode, out.Description)
	}
	c.log.Info("telegram webhook set", "url", url)
	return nil
}

// DeleteWebhook снимает webhook (нужно перед polling).
func (c *Client) DeleteWebhook(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/deleteWebhook?drop_pending_updates=true", nil)
	if err != nil {
		return fmt.Errorf("telegram request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("telegram do: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("telegram read: %w", err)
	}

	var out apiResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return fmt.Errorf("telegram unmarshal: %w", err)
	}
	if !out.OK {
		return fmt.Errorf("telegram deleteWebhook error [%d]: %s", out.ErrorCode, out.Description)
	}
	c.log.Info("telegram webhook deleted")
	return nil
}

// GetUpdates — long polling getUpdates.
func (c *Client) GetUpdates(ctx context.Context, offset int64, timeoutSec int, httpClient *http.Client) ([]Update, error) {
	if timeoutSec <= 0 {
		timeoutSec = 30
	}
	url := fmt.Sprintf("%s/getUpdates?offset=%d&timeout=%d", c.baseURL, offset, timeoutSec)

	client := httpClient
	if client == nil {
		client = c.httpClient
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("telegram request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("telegram do: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("telegram read: %w", err)
	}

	var out getUpdatesResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("telegram unmarshal: %w", err)
	}
	if !out.OK {
		if out.ErrorCode == 409 {
			c.log.Warn("telegram getUpdates conflict (webhook or another poller active)",
				"code", out.ErrorCode, "desc", out.Description)
			return nil, nil
		}
		return nil, fmt.Errorf("telegram getUpdates error [%d]: %s", out.ErrorCode, out.Description)
	}
	return out.Result, nil
}
