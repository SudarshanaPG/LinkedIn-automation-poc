package connect

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-rod/rod"

	"linkedin-automation-poc/internal/browser"
	"linkedin-automation-poc/internal/config"
	"linkedin-automation-poc/internal/limits"
	"linkedin-automation-poc/internal/logger"
	"linkedin-automation-poc/internal/profile"
	"linkedin-automation-poc/internal/storage"
)

const maxNoteChars = 300

type Connector struct {
	cfg   config.ConnectConfig
	log   *logger.Logger
	store *storage.Storage
}

func New(cfg config.ConnectConfig, store *storage.Storage, log *logger.Logger) *Connector {
	return &Connector{cfg: cfg, store: store, log: log}
}

func (c *Connector) SendRequests(ctx context.Context, session *browser.Session, limiter *limits.Limiter, leads []string) error {
	sent := 0
	for _, lead := range leads {
		if err := ctx.Err(); err != nil {
			return err
		}
		if c.cfg.SkipIfSent && c.store.HasSent(lead) {
			continue
		}
		if c.cfg.MaxPerRun > 0 && sent >= c.cfg.MaxPerRun {
			break
		}
		if err := limiter.CheckConnectionLimits(); err != nil {
			limiter.Note(err.Error())
			break
		}

		if err := c.SendRequest(ctx, session, lead); err != nil {
			c.log.Warn("connect failed", map[string]any{"profile": lead, "error": err.Error()})
			continue
		}
		sent++
		session.Stealth.ActionPause(ctx)
	}
	return nil
}

func (c *Connector) SendRequest(ctx context.Context, session *browser.Session, profileURL string) error {
	if err := session.NavigateWithRetry(ctx, profileURL); err != nil {
		return fmt.Errorf("navigate profile: %w", err)
	}
	session.Stealth.Think(ctx)

	if isConnected, _ := hasButton(ctx, session, "Message"); isConnected {
		return nil
	}
	if pending, _ := hasButton(ctx, session, "Pending"); pending {
		return nil
	}

	persona, _ := profile.FromPage(session.Page)
	note := profile.ApplyTemplate(c.cfg.NoteTemplate, persona)
	note = truncate(note, maxNoteChars)

	connectButton, err := findConnectButton(ctx, session)
	if err != nil {
		return err
	}
	if err := session.ClickElementWithRetry(ctx, connectButton, "connect"); err != nil {
		return err
	}

	session.Stealth.ActionPause(ctx)
	if note != "" {
		addNoteBtn, err := session.ElementWithRetry(ctx, "button[aria-label*='Add a note']", 5*time.Second)
		if err == nil && addNoteBtn != nil {
			_ = session.ClickElementWithRetry(ctx, addNoteBtn, "add_note")
			session.Stealth.ActionPause(ctx)
		}

		noteBox, err := session.ElementWithRetry(ctx, "textarea[name='message']", 5*time.Second)
		if err == nil && noteBox != nil {
			_ = session.ClickElementWithRetry(ctx, noteBox, "note_textarea")
			if err := session.Stealth.TypeHuman(ctx, session.Page, note); err != nil {
				return err
			}
		}
	}

	sendBtn, err := findButtonByText(ctx, session, "Send")
	if err != nil {
		return err
	}
	if err := session.ClickElementWithRetry(ctx, sendBtn, "send_invite"); err != nil {
		return err
	}

	c.store.MarkRequestSent(profileURL, note)
	return nil
}

func findConnectButton(ctx context.Context, session *browser.Session) (*rod.Element, error) {
	buttons, err := session.ElementsWithRetry(ctx, "button", 8*time.Second)
	if err != nil {
		return nil, fmt.Errorf("locate connect button: %w", err)
	}
	for _, btn := range buttons {
		text, err := btn.Text()
		if err != nil {
			continue
		}
		if strings.Contains(strings.ToLower(text), "connect") {
			return btn, nil
		}
	}
	return nil, fmt.Errorf("connect button not found")
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

func truncate(value string, limit int) string {
	value = strings.TrimSpace(value)
	if len(value) <= limit {
		return value
	}
	return value[:limit]
}
