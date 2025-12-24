package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Auth      AuthConfig      `json:"auth" yaml:"auth"`
	Search    SearchConfig    `json:"search" yaml:"search"`
	Connect   ConnectConfig   `json:"connect" yaml:"connect"`
	Messaging MessagingConfig `json:"messaging" yaml:"messaging"`
	Stealth   StealthConfig   `json:"stealth" yaml:"stealth"`
	Schedule  ScheduleConfig  `json:"schedule" yaml:"schedule"`
	Limits    LimitsConfig    `json:"limits" yaml:"limits"`
	Storage   StorageConfig   `json:"storage" yaml:"storage"`
	Logging   LoggingConfig   `json:"logging" yaml:"logging"`
	Browser   BrowserConfig   `json:"browser" yaml:"browser"`
}

type AuthConfig struct {
	LoginURL        string `json:"login_url" yaml:"login_url"`
	CookiePath      string `json:"cookie_path" yaml:"cookie_path"`
	ReuseCookies    bool   `json:"reuse_cookies" yaml:"reuse_cookies"`
	SaveCookies     bool   `json:"save_cookies" yaml:"save_cookies"`
	CheckpointPause bool   `json:"checkpoint_pause" yaml:"checkpoint_pause"`
}

type SearchConfig struct {
	Title       string   `json:"title" yaml:"title"`
	Company     string   `json:"company" yaml:"company"`
	Companies   []string `json:"companies" yaml:"companies"`
	CompanyOnly bool     `json:"company_only" yaml:"company_only"`
	LeadsFile   string   `json:"leads_file" yaml:"leads_file"`
	Location    string   `json:"location" yaml:"location"`
	Keywords    []string `json:"keywords" yaml:"keywords"`
	MaxPages    int      `json:"max_pages" yaml:"max_pages"`
	MaxLeads    int      `json:"max_leads" yaml:"max_leads"`
	SortBy      string   `json:"sort_by" yaml:"sort_by"`
}

type ConnectConfig struct {
	DailyLimit   int    `json:"daily_limit" yaml:"daily_limit"`
	MaxPerRun    int    `json:"max_per_run" yaml:"max_per_run"`
	NoteTemplate string `json:"note_template" yaml:"note_template"`
	SkipIfSent   bool   `json:"skip_if_sent" yaml:"skip_if_sent"`
}

type MessagingConfig struct {
	DailyLimit      int    `json:"daily_limit" yaml:"daily_limit"`
	MaxPerRun       int    `json:"max_per_run" yaml:"max_per_run"`
	Template        string `json:"template" yaml:"template"`
	FollowUpDelayHr int    `json:"follow_up_delay_hr" yaml:"follow_up_delay_hr"`
}

type StealthConfig struct {
	UserAgent           string `json:"user_agent" yaml:"user_agent"`
	Locale              string `json:"locale" yaml:"locale"`
	Timezone            string `json:"timezone" yaml:"timezone"`
	ViewportWidth       int    `json:"viewport_width" yaml:"viewport_width"`
	ViewportHeight      int    `json:"viewport_height" yaml:"viewport_height"`
	Headless            bool   `json:"headless" yaml:"headless"`
	MouseMoveJitter     int    `json:"mouse_move_jitter" yaml:"mouse_move_jitter"`
	ThinkTimeMinMs      int    `json:"think_time_min_ms" yaml:"think_time_min_ms"`
	ThinkTimeMaxMs      int    `json:"think_time_max_ms" yaml:"think_time_max_ms"`
	TypingDelayMinMs    int    `json:"typing_delay_min_ms" yaml:"typing_delay_min_ms"`
	TypingDelayMaxMs    int    `json:"typing_delay_max_ms" yaml:"typing_delay_max_ms"`
	ScrollStepMin       int    `json:"scroll_step_min" yaml:"scroll_step_min"`
	ScrollStepMax       int    `json:"scroll_step_max" yaml:"scroll_step_max"`
	EnableRodStealth    bool   `json:"enable_rod_stealth" yaml:"enable_rod_stealth"`
	RandomizeViewport   bool   `json:"randomize_viewport" yaml:"randomize_viewport"`
	ViewportVariancePx  int    `json:"viewport_variance_px" yaml:"viewport_variance_px"`
	ActionIntervalMinMs int    `json:"action_interval_min_ms" yaml:"action_interval_min_ms"`
	ActionIntervalMaxMs int    `json:"action_interval_max_ms" yaml:"action_interval_max_ms"`
	ScrollBackChance    int    `json:"scroll_back_chance" yaml:"scroll_back_chance"`
	TypoChance          int    `json:"typo_chance" yaml:"typo_chance"`
	MouseWanderChance   int    `json:"mouse_wander_chance" yaml:"mouse_wander_chance"`
	HoverPauseMinMs     int    `json:"hover_pause_min_ms" yaml:"hover_pause_min_ms"`
	HoverPauseMaxMs     int    `json:"hover_pause_max_ms" yaml:"hover_pause_max_ms"`
}

type ScheduleConfig struct {
	Enabled      bool     `json:"enabled" yaml:"enabled"`
	Timezone     string   `json:"timezone" yaml:"timezone"`
	BusinessDays []int    `json:"business_days" yaml:"business_days"`
	StartHour    int      `json:"start_hour" yaml:"start_hour"`
	EndHour      int      `json:"end_hour" yaml:"end_hour"`
	Breaks       []string `json:"breaks" yaml:"breaks"`
	Enforce      bool     `json:"enforce" yaml:"enforce"`
}

type LimitsConfig struct {
	ConnectionPerHour int `json:"connection_per_hour" yaml:"connection_per_hour"`
	MessagePerHour    int `json:"message_per_hour" yaml:"message_per_hour"`
}

type StorageConfig struct {
	Path                 string `json:"path" yaml:"path"`
	FlushIntervalSeconds int    `json:"flush_interval_seconds" yaml:"flush_interval_seconds"`
}

type LoggingConfig struct {
	Level  string `json:"level" yaml:"level"`
	Format string `json:"format" yaml:"format"`
	Path   string `json:"path" yaml:"path"`
}

type BrowserConfig struct {
	ExecutablePath string `json:"executable_path" yaml:"executable_path"`
}

func DefaultConfig() Config {
	return Config{
		Auth: AuthConfig{
			LoginURL:        "https://www.linkedin.com/login",
			CookiePath:      "data/cookies.json",
			ReuseCookies:    true,
			SaveCookies:     true,
			CheckpointPause: true,
		},
		Search: SearchConfig{
			MaxPages: 5,
			MaxLeads: 50,
			SortBy:   "relevance",
		},
		Connect: ConnectConfig{
			DailyLimit:   25,
			MaxPerRun:    10,
			NoteTemplate: "Hi {{FirstName}}, I enjoyed your recent work at {{Company}} and would love to connect.",
			SkipIfSent:   true,
		},
		Messaging: MessagingConfig{
			DailyLimit:      25,
			MaxPerRun:       10,
			Template:        "Thanks for connecting, {{FirstName}}. Curious about your work in {{Industry}}.",
			FollowUpDelayHr: 24,
		},
		Stealth: StealthConfig{
			UserAgent:           "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			Locale:              "en-US",
			Timezone:            "America/New_York",
			ViewportWidth:       1366,
			ViewportHeight:      768,
			Headless:            false,
			MouseMoveJitter:     3,
			ThinkTimeMinMs:      600,
			ThinkTimeMaxMs:      1800,
			TypingDelayMinMs:    45,
			TypingDelayMaxMs:    190,
			ScrollStepMin:       280,
			ScrollStepMax:       820,
			EnableRodStealth:    true,
			RandomizeViewport:   true,
			ViewportVariancePx:  80,
			ActionIntervalMinMs: 350,
			ActionIntervalMaxMs: 1200,
			ScrollBackChance:    15,
			TypoChance:          4,
			MouseWanderChance:   20,
			HoverPauseMinMs:     200,
			HoverPauseMaxMs:     800,
		},
		Schedule: ScheduleConfig{
			Enabled:      true,
			Timezone:     "America/New_York",
			BusinessDays: []int{1, 2, 3, 4, 5},
			StartHour:    9,
			EndHour:      17,
			Breaks:       []string{"12:00-12:45", "15:15-15:30"},
			Enforce:      false,
		},
		Limits: LimitsConfig{
			ConnectionPerHour: 8,
			MessagePerHour:    8,
		},
		Storage: StorageConfig{
			Path:                 "data/state.json",
			FlushIntervalSeconds: 15,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "text",
			Path:   "",
		},
		Browser: BrowserConfig{
			ExecutablePath: "",
		},
	}
}

func Load(path string) (Config, error) {
	cfg := DefaultConfig()
	if path == "" {
		path = "config.yaml"
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return cfg, fmt.Errorf("read config: %w", err)
		}
	} else {
		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".json":
			if err := json.Unmarshal(data, &cfg); err != nil {
				return cfg, fmt.Errorf("parse json config: %w", err)
			}
		default:
			if err := yaml.Unmarshal(data, &cfg); err != nil {
				return cfg, fmt.Errorf("parse yaml config: %w", err)
			}
		}
	}

	ApplyEnvOverrides(&cfg)
	if err := cfg.Validate(); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func ApplyEnvOverrides(cfg *Config) {
	if v := os.Getenv("LINKEDIN_USER_AGENT"); v != "" {
		cfg.Stealth.UserAgent = v
	}
	if v := os.Getenv("LINKEDIN_BROWSER_PATH"); v != "" {
		cfg.Browser.ExecutablePath = v
	}
	if v := os.Getenv("LINKEDIN_HEADLESS"); v != "" {
		cfg.Stealth.Headless = parseBool(v, cfg.Stealth.Headless)
	}
	if v := os.Getenv("LINKEDIN_VIEWPORT_WIDTH"); v != "" {
		cfg.Stealth.ViewportWidth = parseInt(v, cfg.Stealth.ViewportWidth)
	}
	if v := os.Getenv("LINKEDIN_VIEWPORT_HEIGHT"); v != "" {
		cfg.Stealth.ViewportHeight = parseInt(v, cfg.Stealth.ViewportHeight)
	}
	if v := os.Getenv("LINKEDIN_STORAGE_PATH"); v != "" {
		cfg.Storage.Path = v
	}
	if v := os.Getenv("LINKEDIN_COOKIE_PATH"); v != "" {
		cfg.Auth.CookiePath = v
	}
	if v := os.Getenv("LINKEDIN_LOG_LEVEL"); v != "" {
		cfg.Logging.Level = v
	}
	if v := os.Getenv("LINKEDIN_LOG_FORMAT"); v != "" {
		cfg.Logging.Format = v
	}
	if v := os.Getenv("LINKEDIN_SCHEDULE_ENFORCE"); v != "" {
		cfg.Schedule.Enforce = parseBool(v, cfg.Schedule.Enforce)
	}
}

func (cfg Config) Validate() error {
	if cfg.Search.MaxPages <= 0 {
		return fmt.Errorf("search.max_pages must be > 0")
	}
	if cfg.Search.MaxLeads <= 0 {
		return fmt.Errorf("search.max_leads must be > 0")
	}
	if strings.TrimSpace(cfg.Search.LeadsFile) == "" {
		if cfg.Search.CompanyOnly {
			if strings.TrimSpace(cfg.Search.Company) == "" && len(cfg.Search.Companies) == 0 {
				return fmt.Errorf("search.company_only requires search.company or search.companies")
			}
		} else {
			if strings.TrimSpace(cfg.Search.Title) == "" && strings.TrimSpace(cfg.Search.Company) == "" && strings.TrimSpace(cfg.Search.Location) == "" && !hasNonEmpty(cfg.Search.Keywords) {
				return fmt.Errorf("at least one search filter is required (title/company/location/keywords)")
			}
		}
	}
	if cfg.Connect.DailyLimit <= 0 || cfg.Messaging.DailyLimit <= 0 {
		return fmt.Errorf("daily limits must be > 0")
	}
	if cfg.Stealth.ThinkTimeMinMs < 0 || cfg.Stealth.ThinkTimeMaxMs < cfg.Stealth.ThinkTimeMinMs {
		return fmt.Errorf("invalid stealth think time range")
	}
	if cfg.Stealth.ActionIntervalMinMs < 0 || cfg.Stealth.ActionIntervalMaxMs < cfg.Stealth.ActionIntervalMinMs {
		return fmt.Errorf("invalid stealth action interval range")
	}
	if cfg.Schedule.StartHour < 0 || cfg.Schedule.StartHour > 23 || cfg.Schedule.EndHour < 0 || cfg.Schedule.EndHour > 23 {
		return fmt.Errorf("invalid schedule hours")
	}
	if cfg.Storage.Path == "" {
		return fmt.Errorf("storage.path is required")
	}
	if cfg.Logging.Level == "" {
		cfg.Logging.Level = "info"
	}
	if cfg.Logging.Format == "" {
		cfg.Logging.Format = "text"
	}
	return nil
}

func hasNonEmpty(values []string) bool {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return true
		}
	}
	return false
}

func parseInt(raw string, fallback int) int {
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return fallback
	}
	return value
}

func parseBool(raw string, fallback bool) bool {
	value, err := strconv.ParseBool(strings.TrimSpace(raw))
	if err != nil {
		return fallback
	}
	return value
}

func (cfg Config) FollowUpDelay() time.Duration {
	return time.Duration(cfg.Messaging.FollowUpDelayHr) * time.Hour
}

func (cfg Config) BaseURL() string {
	parsed, err := url.Parse(cfg.Auth.LoginURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "https://www.linkedin.com"
	}
	return parsed.Scheme + "://" + parsed.Host
}
