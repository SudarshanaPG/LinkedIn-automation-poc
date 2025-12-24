package messaging

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"

	"linkedin-automation-poc/internal/browser"
	"linkedin-automation-poc/internal/config"
	"linkedin-automation-poc/internal/limits"
	"linkedin-automation-poc/internal/logger"
	"linkedin-automation-poc/internal/profile"
	"linkedin-automation-poc/internal/storage"
)

type Messenger struct {
	cfg   config.MessagingConfig
	store *storage.Storage
	log   *logger.Logger
}

func New(cfg config.MessagingConfig, store *storage.Storage, log *logger.Logger) *Messenger {
	return &Messenger{cfg: cfg, store: store, log: log}
}

func (m *Messenger) SyncAccepted(ctx context.Context, session *browser.Session) error {
	pending := m.store.PendingRequests()
	for _, record := range pending {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := session.NavigateWithRetry(ctx, record.ProfileURL); err != nil {
			m.log.Warn("navigate profile failed", map[string]any{"profile": record.ProfileURL, "error": err.Error()})
			continue
		}
		session.Stealth.Think(ctx)
		accepted, _ := hasButton(ctx, session, "Message")
		if accepted {
			m.store.MarkAccepted(record.ProfileURL)
		}
		session.Stealth.ActionPause(ctx)
	}
	return nil
}

func (m *Messenger) SendFollowUps(ctx context.Context, session *browser.Session, limiter *limits.Limiter) error {
	accepted := m.store.AcceptedConnections()
	sent := 0
	for _, record := range accepted {
		if err := ctx.Err(); err != nil {
			return err
		}
		if m.cfg.MaxPerRun > 0 && sent >= m.cfg.MaxPerRun {
			break
		}
		if m.store.HasMessaged(record.ProfileURL) {
			continue
		}
		if time.Since(record.AcceptedAt) < time.Duration(m.cfg.FollowUpDelayHr)*time.Hour {
			continue
		}
		if err := limiter.CheckMessageLimits(); err != nil {
			limiter.Note(err.Error())
			break
		}
		if err := m.sendMessage(ctx, session, record.ProfileURL); err != nil {
			m.log.Warn("message failed", map[string]any{"profile": record.ProfileURL, "error": err.Error()})
			continue
		}
		sent++
		session.Stealth.ActionPause(ctx)
	}
	return nil
}

func (m *Messenger) sendMessage(ctx context.Context, session *browser.Session, profileURL string) error {
	if err := session.NavigateWithRetry(ctx, profileURL); err != nil {
		return err
	}
	session.Stealth.Think(ctx)

	messageBtn, err := findButtonByText(ctx, session, "Message")
	if err != nil {
		return err
	}
	if err := session.ClickElementWithRetry(ctx, messageBtn, "message"); err != nil {
		return err
	}

	editor, err := waitForMessageEditor(ctx, session)
	if err != nil {
		return err
	}
	_ = session.ClickElementWithRetry(ctx, editor, "message_editor")

	persona, _ := profile.FromPage(session.Page)
	body := profile.ApplyTemplate(m.cfg.Template, persona)
	if body == "" {
		return fmt.Errorf("message template resolved to empty")
	}
	if err := session.Stealth.TypeHuman(ctx, session.Page, body); err != nil {
		return err
	}
	if err := session.Page.Keyboard.Type(input.Enter); err != nil {
		return err
	}

	m.store.AddMessage(profileURL, m.cfg.Template, body)
	return nil
}

func waitForMessageEditor(ctx context.Context, session *browser.Session) (*rod.Element, error) {
	editor, err := session.ElementWithRetry(ctx, "div.msg-form__contenteditable", 8*time.Second)
	if err == nil {
		return editor, nil
	}
	return session.ElementWithRetry(ctx, "div[role='textbox']", 8*time.Second)
}

func findButtonByText(ctx context.Context, session *browser.Session, label string) (*rod.Element, error) {
	buttons, err := session.ElementsWithRetry(ctx, "button", 8*time.Second)
	if err != nil {
		return nil, err
	}
	needle := strings.ToLower(label)
	for _, btn := range buttons {
		text, err := btn.Text()
		if err != nil {
			continue
		}
		if strings.Contains(strings.ToLower(text), needle) {
			return btn, nil
		}
	}
	return nil, fmt.Errorf("button not found: %s", label)
}

func hasButton(ctx context.Context, session *browser.Session, label string) (bool, error) {
	_, err := findButtonByText(ctx, session, label)
	if err != nil {
		return false, nil
	}
	return true, nil
}
