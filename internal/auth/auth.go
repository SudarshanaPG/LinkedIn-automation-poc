package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"

	"linkedin-automation-poc/internal/browser"
	"linkedin-automation-poc/internal/config"
	"linkedin-automation-poc/internal/logger"
)

type Authenticator struct {
	cfg config.AuthConfig
	log *logger.Logger
}

func New(cfg config.AuthConfig, log *logger.Logger) *Authenticator {
	return &Authenticator{cfg: cfg, log: log}
}

func CredentialsFromEnv() (string, string, error) {
	email := strings.TrimSpace(os.Getenv("LINKEDIN_EMAIL"))
	password := strings.TrimSpace(os.Getenv("LINKEDIN_PASSWORD"))
	if email == "" || password == "" {
		return "", "", fmt.Errorf("LINKEDIN_EMAIL and LINKEDIN_PASSWORD are required")
	}
	return email, password, nil
}

func (a *Authenticator) Login(ctx context.Context, session *browser.Session) error {
	baseURL := session.BaseURL()
	if a.cfg.ReuseCookies {
		if ok, err := a.loadCookies(session.Page); err != nil {
			a.log.Warn("cookie load failed", map[string]any{"error": err.Error()})
		} else if ok {
			if err := session.NavigateWithRetry(ctx, baseURL+"/feed/"); err != nil {
				return fmt.Errorf("navigate feed: %w", err)
			}
			session.Stealth.Think(ctx)
			if loggedIn, _ := a.IsLoggedIn(session.Page); loggedIn {
				a.log.Info("reused session cookies", nil)
				return nil
			}
		}
	}

	email, password, err := CredentialsFromEnv()
	if err != nil {
		return err
	}

	if err := session.NavigateWithRetry(ctx, a.cfg.LoginURL); err != nil {
		return fmt.Errorf("navigate login: %w", err)
	}
	session.Stealth.Think(ctx)

	emailField, err := session.ElementWithRetry(ctx, "input#username", 7*time.Second)
	if err != nil {
		return fmt.Errorf("locate email field: %w", err)
	}
	if err := session.ClickElementWithRetry(ctx, emailField, "login_email"); err != nil {
		return err
	}
	if err := session.Stealth.TypeHuman(ctx, session.Page, email); err != nil {
		return err
	}
	session.Stealth.ActionPause(ctx)

	passwordField, err := session.ElementWithRetry(ctx, "input#password", 7*time.Second)
	if err != nil {
		return fmt.Errorf("locate password field: %w", err)
	}
	if err := session.ClickElementWithRetry(ctx, passwordField, "login_password"); err != nil {
		return err
	}
	if err := session.Stealth.TypeHuman(ctx, session.Page, password); err != nil {
		return err
	}
	session.Stealth.ActionPause(ctx)

	submit, err := session.ElementWithRetry(ctx, "button[type='submit']", 7*time.Second)
	if err != nil {
		return fmt.Errorf("locate submit button: %w", err)
	}
	if err := session.ClickElementWithRetry(ctx, submit, "login_submit"); err != nil {
		return err
	}

	session.Stealth.Think(ctx)
	if checkpoint, reason := a.DetectCheckpoint(session.Page); checkpoint {
		if a.cfg.CheckpointPause {
			a.log.Warn("checkpoint detected, waiting for manual resolution", map[string]any{"reason": reason})
			if err := waitForCheckpointClear(ctx, session.Page); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("checkpoint detected: %s", reason)
		}
	}

	loggedIn, err := a.IsLoggedIn(session.Page)
	if err != nil {
		return err
	}
	if !loggedIn {
		return errors.New("login failed")
	}

	if a.cfg.SaveCookies {
		if err := a.saveCookies(session.Page, baseURL); err != nil {
			a.log.Warn("cookie save failed", map[string]any{"error": err.Error()})
		}
	}
	return nil
}

func (a *Authenticator) IsLoggedIn(page *rod.Page) (bool, error) {
	info, err := page.Info()
	if err != nil {
		return false, err
	}
	if strings.Contains(info.URL, "/feed") || strings.Contains(info.URL, "/mynetwork") {
		return true, nil
	}
	if _, err := page.Timeout(3 * time.Second).Element("a[href*='mynetwork']"); err == nil {
		return true, nil
	}
	return false, nil
}

func (a *Authenticator) DetectCheckpoint(page *rod.Page) (bool, string) {
	info, err := page.Info()
	if err == nil {
		if strings.Contains(info.URL, "checkpoint") || strings.Contains(info.URL, "challenge") || strings.Contains(info.URL, "captcha") {
			return true, "checkpoint url"
		}
	}
	if _, err := page.Timeout(2 * time.Second).Element("input[name='pin']"); err == nil {
		return true, "2fa prompt"
	}
	if _, err := page.Timeout(2 * time.Second).Element("iframe[src*='captcha']"); err == nil {
		return true, "captcha iframe"
	}
	if _, err := page.Timeout(2 * time.Second).Element("div.challenge-dialog"); err == nil {
		return true, "challenge dialog"
	}
	return false, ""
}

func waitForCheckpointClear(ctx context.Context, page *rod.Page) error {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			info, err := page.Info()
			if err == nil && strings.Contains(info.URL, "/feed") {
				return nil
			}
		}
	}
}

func (a *Authenticator) loadCookies(page *rod.Page) (bool, error) {
	data, err := os.ReadFile(a.cfg.CookiePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	var payload cookieJar
	if err := json.Unmarshal(data, &payload); err != nil {
		return false, err
	}
	if len(payload.Cookies) == 0 {
		return false, nil
	}
	return true, (proto.NetworkSetCookies{Cookies: payload.Cookies}).Call(page)
}

func (a *Authenticator) saveCookies(page *rod.Page, baseURL string) error {
	cookies, err := page.Cookies([]string{baseURL})
	if err != nil {
		return err
	}
	payload := cookieJar{Cookies: toCookieParams(cookies)}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(a.cfg.CookiePath), 0755); err != nil {
		return err
	}
	return os.WriteFile(a.cfg.CookiePath, data, 0644)
}

type cookieJar struct {
	Cookies []*proto.NetworkCookieParam `json:"cookies"`
}

func toCookieParams(cookies []*proto.NetworkCookie) []*proto.NetworkCookieParam {
	params := make([]*proto.NetworkCookieParam, 0, len(cookies))
	for _, cookie := range cookies {
		params = append(params, &proto.NetworkCookieParam{
			Name:     cookie.Name,
			Value:    cookie.Value,
			Domain:   cookie.Domain,
			Path:     cookie.Path,
			Expires:  cookie.Expires,
			HTTPOnly: cookie.HTTPOnly,
			Secure:   cookie.Secure,
			SameSite: cookie.SameSite,
		})
	}
	return params
}
