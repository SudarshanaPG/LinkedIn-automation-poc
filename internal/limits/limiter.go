package limits

import (
	"fmt"
	"time"

	"linkedin-automation-poc/internal/config"
	"linkedin-automation-poc/internal/logger"
	"linkedin-automation-poc/internal/storage"
)

type Limiter struct {
	cfg   config.Config
	store *storage.Storage
	log   *logger.Logger
}

func New(cfg config.Config, store *storage.Storage, log *logger.Logger) *Limiter {
	return &Limiter{cfg: cfg, store: store, log: log}
}

func (l *Limiter) CheckConnectionLimits() error {
	now := time.Now()
	dayCount := l.store.CountRequestsSince(now.Add(-24 * time.Hour))
	hourCount := l.store.CountRequestsSince(now.Add(-1 * time.Hour))
	if dayCount >= l.cfg.Connect.DailyLimit {
		return fmt.Errorf("daily connection limit reached")
	}
	if l.cfg.Limits.ConnectionPerHour > 0 && hourCount >= l.cfg.Limits.ConnectionPerHour {
		return fmt.Errorf("hourly connection limit reached")
	}
	return nil
}

func (l *Limiter) CheckMessageLimits() error {
	now := time.Now()
	dayCount := l.store.CountMessagesSince(now.Add(-24 * time.Hour))
	hourCount := l.store.CountMessagesSince(now.Add(-1 * time.Hour))
	if dayCount >= l.cfg.Messaging.DailyLimit {
		return fmt.Errorf("daily message limit reached")
	}
	if l.cfg.Limits.MessagePerHour > 0 && hourCount >= l.cfg.Limits.MessagePerHour {
		return fmt.Errorf("hourly message limit reached")
	}
	return nil
}

func (l *Limiter) Note(reason string) {
	if l.log != nil {
		l.log.Warn("rate limiting triggered", map[string]any{"reason": reason})
	}
}
