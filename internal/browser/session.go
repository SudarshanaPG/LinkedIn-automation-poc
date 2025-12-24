package browser

import (
	"context"
	"fmt"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"

	"linkedin-automation-poc/internal/config"
	"linkedin-automation-poc/internal/logger"
	"linkedin-automation-poc/internal/stealth"
)

type Session struct {
	Browser   *rod.Browser
	Page      *rod.Page
	MousePos  stealth.Point
	Stealth   *stealth.Controller
	cfg       config.Config
	log       *logger.Logger
	viewportW int
	viewportH int
}

func NewSession(ctx context.Context, cfg config.Config, log *logger.Logger, stealthCtrl *stealth.Controller) (*Session, error) {
	launch := launcher.New().
		Leakless(false).
		Headless(cfg.Stealth.Headless).
		Set("disable-blink-features", "AutomationControlled").
		Set("disable-infobars").
		Set("no-first-run").
		Set("disable-default-apps").
		Set("disable-dev-shm-usage").
		Set("disable-popup-blocking")

	binPath := cfg.Browser.ExecutablePath
	if binPath == "" {
		binPath = FindBrowserBinary()
	}
	if binPath != "" {
		launch = launch.Bin(binPath)
	}

	launchURL, err := launch.Launch()
	if err != nil {
		return nil, fmt.Errorf("launch browser: %w", err)
	}

	browser := rod.New().ControlURL(launchURL)
	if err := browser.Connect(); err != nil {
		return nil, fmt.Errorf("connect browser: %w", err)
	}

	page, err := browser.Page(proto.TargetCreateTarget{URL: "about:blank"})
	if err != nil {
		return nil, fmt.Errorf("create page: %w", err)
	}

	width, height := cfg.Stealth.ViewportWidth, cfg.Stealth.ViewportHeight
	width, height = stealthCtrl.RandomizeViewport(width, height)
	if err := (proto.EmulationSetDeviceMetricsOverride{
		Width:             width,
		Height:            height,
		DeviceScaleFactor: 1,
		Mobile:            false,
	}).Call(page); err != nil {
		return nil, fmt.Errorf("set viewport: %w", err)
	}

	if err := (proto.EmulationSetTimezoneOverride{TimezoneID: cfg.Stealth.Timezone}).Call(page); err != nil {
		return nil, fmt.Errorf("set timezone: %w", err)
	}
	if err := (proto.EmulationSetLocaleOverride{Locale: cfg.Stealth.Locale}).Call(page); err != nil {
		return nil, fmt.Errorf("set locale: %w", err)
	}
	if err := (proto.NetworkSetUserAgentOverride{
		UserAgent:      cfg.Stealth.UserAgent,
		AcceptLanguage: cfg.Stealth.Locale,
		Platform:       "Win32",
	}).Call(page); err != nil {
		return nil, fmt.Errorf("set user agent: %w", err)
	}
	if cfg.Stealth.EnableRodStealth {
		if _, err := page.EvalOnNewDocument(stealth.FingerprintScript(cfg.Stealth)); err != nil {
			return nil, fmt.Errorf("inject fingerprint: %w", err)
		}
	}

	return &Session{
		Browser:   browser,
		Page:      page,
		MousePos:  stealth.Point{X: float64(width / 2), Y: float64(height / 2)},
		Stealth:   stealthCtrl,
		cfg:       cfg,
		log:       log,
		viewportW: width,
		viewportH: height,
	}, nil
}

func (s *Session) Close() error {
	if s.Browser == nil {
		return nil
	}
	return s.Browser.Close()
}

func (s *Session) BaseURL() string {
	return s.cfg.BaseURL()
}

func (s *Session) MoveMouseTo(ctx context.Context, x, y float64) error {
	next, err := s.Stealth.MoveMouseHuman(ctx, s.Page, s.MousePos, stealth.Point{X: x, Y: y})
	if err != nil {
		return err
	}
	s.MousePos = next
	return nil
}

func (s *Session) HoverElement(ctx context.Context, el *rod.Element) error {
	next, err := s.Stealth.HoverElement(ctx, s.Page, s.MousePos, el)
	if err != nil {
		return err
	}
	s.MousePos = next
	return nil
}

func (s *Session) MaybeWander(ctx context.Context) {
	if !s.Stealth.ShouldWanderMouse() {
		return
	}
	next, err := s.Stealth.WanderMouse(ctx, s.Page, s.MousePos, s.viewportW, s.viewportH)
	if err == nil {
		s.MousePos = next
	}
}
