package telegram

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

// UpdateHandler — обработчик входящего update (polling).
type UpdateHandler func(ctx context.Context, update *Update) error

// Poller — long polling getUpdates (для local/dev).
type Poller struct {
	client       *Client
	timeoutSec   int
	handler      UpdateHandler
	lastUpdateID int64
	log          *slog.Logger
	httpClient   *http.Client
}

// NewPoller создаёт poller.
func NewPoller(client *Client, timeoutSec int, handler UpdateHandler, log *slog.Logger) *Poller {
	if timeoutSec <= 0 {
		timeoutSec = 30
	}
	return &Poller{
		client:     client,
		timeoutSec: timeoutSec,
		handler:    handler,
		log:        log,
		httpClient: &http.Client{
			Timeout: time.Duration(timeoutSec+10) * time.Second,
		},
	}
}

// Start крутит long polling до отмены ctx.
func (p *Poller) Start(ctx context.Context) {
	p.log.Info("telegram polling started", "timeout_sec", p.timeoutSec)
	for {
		select {
		case <-ctx.Done():
			p.log.Info("telegram polling stopped")
			return
		default:
			updates, err := p.client.GetUpdates(ctx, p.lastUpdateID, p.timeoutSec, p.httpClient)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				p.log.Error("telegram getUpdates failed", "error", err)
				time.Sleep(5 * time.Second)
				continue
			}
			for i := range updates {
				u := updates[i]
				if u.UpdateID >= p.lastUpdateID {
					p.lastUpdateID = u.UpdateID + 1
				}
				if err := p.handler(ctx, &u); err != nil {
					p.log.Error("telegram update handle failed", "error", err, "update_id", u.UpdateID)
				}
			}
		}
	}
}
