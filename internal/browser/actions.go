package browser

import (
	"context"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"

	"linkedin-automation-poc/internal/retry"
)

func (s *Session) NavigateWithRetry(ctx context.Context, url string) error {
	policy := retry.DefaultPolicy()
	policy.Retryable = func(err error) bool {
		msg := err.Error()
		if strings.Contains(msg, "ERR_INVALID_AUTH_CREDENTIALS") {
			return false
		}
		return true
	}
	policy.OnRetry = func(attempt int, delay time.Duration, err error) {
		if s.log != nil {
			s.log.Debug("retrying navigation", map[string]any{
				"url":           url,
				"attempt":       attempt,
				"next_delay_ms": delay.Milliseconds(),
				"error":         err.Error(),
			})
		}
	}
	return retry.Do(ctx, policy, func(_ int) error {
		return s.Page.Navigate(url)
	})
}

func (s *Session) ElementWithRetry(ctx context.Context, selector string, timeout time.Duration) (*rod.Element, error) {
	var el *rod.Element
	policy := retry.DefaultPolicy()
	policy.Attempts = 3
	policy.BaseDelay = 350 * time.Millisecond
	policy.MaxDelay = 2 * time.Second
	policy.OnRetry = func(attempt int, delay time.Duration, err error) {
		if s.log != nil {
			s.log.Debug("retrying element lookup", map[string]any{
				"selector":      selector,
				"attempt":       attempt,
				"next_delay_ms": delay.Milliseconds(),
				"error":         err.Error(),
			})
		}
	}
	err := retry.Do(ctx, policy, func(_ int) error {
		found, err := s.Page.Timeout(timeout).Element(selector)
		if err != nil {
			return err
		}
		el = found
		return nil
	})
	return el, err
}

func (s *Session) ElementsWithRetry(ctx context.Context, selector string, timeout time.Duration) (rod.Elements, error) {
	var els rod.Elements
	policy := retry.DefaultPolicy()
	policy.Attempts = 3
	policy.BaseDelay = 350 * time.Millisecond
	policy.MaxDelay = 2 * time.Second
	policy.OnRetry = func(attempt int, delay time.Duration, err error) {
		if s.log != nil {
			s.log.Debug("retrying elements lookup", map[string]any{
				"selector":      selector,
				"attempt":       attempt,
				"next_delay_ms": delay.Milliseconds(),
				"error":         err.Error(),
			})
		}
	}
	err := retry.Do(ctx, policy, func(_ int) error {
		found, err := s.Page.Timeout(timeout).Elements(selector)
		if err != nil {
			return err
		}
		els = found
		return nil
	})
	return els, err
}

func (s *Session) ClickElementWithRetry(ctx context.Context, el *rod.Element, label string) error {
	policy := retry.DefaultPolicy()
	policy.Attempts = 3
	policy.BaseDelay = 300 * time.Millisecond
	policy.MaxDelay = 3 * time.Second
	policy.OnRetry = func(attempt int, delay time.Duration, err error) {
		if s.log != nil {
			s.log.Debug("retrying click", map[string]any{
				"label":         label,
				"attempt":       attempt,
				"next_delay_ms": delay.Milliseconds(),
				"error":         err.Error(),
			})
		}
	}
	return retry.Do(ctx, policy, func(_ int) error {
		_ = s.HoverElement(ctx, el)
		return el.Click(proto.InputMouseButtonLeft, 1)
	})
}
