package stealth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"linkedin-automation-poc/internal/config"
	"linkedin-automation-poc/internal/logger"
)

type Scheduler struct {
	cfg    config.ScheduleConfig
	loc    *time.Location
	breaks []timeRange
}

type timeRange struct {
	Start time.Duration
	End   time.Duration
}

func NewScheduler(cfg config.ScheduleConfig) (*Scheduler, error) {
	loc, err := time.LoadLocation(cfg.Timezone)
	if err != nil {
		return nil, fmt.Errorf("load schedule timezone: %w", err)
	}
	breaks := parseBreaks(cfg.Breaks)
	return &Scheduler{cfg: cfg, loc: loc, breaks: breaks}, nil
}

func (s *Scheduler) Enforce(ctx context.Context, log *logger.Logger) error {
	if !s.cfg.Enabled {
		return nil
	}
	now := time.Now().In(s.loc)
	if s.IsWithinWindow(now) {
		return nil
	}
	next := s.NextWindow(now)
	if s.cfg.Enforce {
		return fmt.Errorf("outside activity window, next slot: %s", next.Format(time.RFC3339))
	}
	if log != nil {
		log.Warn("outside activity window, continuing for demo", map[string]any{"next": next.Format(time.RFC3339)})
	}
	return nil
}

func (s *Scheduler) WaitForWindow(ctx context.Context) error {
	if !s.cfg.Enabled {
		return nil
	}
	now := time.Now().In(s.loc)
	if s.IsWithinWindow(now) {
		return nil
	}
	next := s.NextWindow(now)
	timer := time.NewTimer(time.Until(next))
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (s *Scheduler) IsWithinWindow(now time.Time) bool {
	if !containsInt(s.cfg.BusinessDays, int(now.Weekday())) {
		return false
	}
	tod := timeOfDay(now)
	start := time.Duration(s.cfg.StartHour) * time.Hour
	end := time.Duration(s.cfg.EndHour) * time.Hour
	if tod < start || tod > end {
		return false
	}
	for _, br := range s.breaks {
		if tod >= br.Start && tod <= br.End {
			return false
		}
	}
	return true
}

func (s *Scheduler) NextWindow(now time.Time) time.Time {
	current := now
	for i := 0; i < 8; i++ {
		day := current
		if containsInt(s.cfg.BusinessDays, int(day.Weekday())) {
			start := time.Date(day.Year(), day.Month(), day.Day(), s.cfg.StartHour, 0, 0, 0, s.loc)
			end := time.Date(day.Year(), day.Month(), day.Day(), s.cfg.EndHour, 0, 0, 0, s.loc)
			if current.Before(start) {
				return start
			}
			if current.After(end) {
				current = start.Add(24 * time.Hour)
				continue
			}
			for _, br := range s.breaks {
				brStart := start.Add(br.Start)
				brEnd := start.Add(br.End)
				if current.After(brStart) && current.Before(brEnd) {
					return brEnd
				}
			}
			return current
		}
		current = time.Date(day.Year(), day.Month(), day.Day(), s.cfg.StartHour, 0, 0, 0, s.loc).Add(24 * time.Hour)
	}
	return now.Add(24 * time.Hour)
}

func parseBreaks(entries []string) []timeRange {
	breaks := make([]timeRange, 0, len(entries))
	for _, entry := range entries {
		parts := strings.Split(entry, "-")
		if len(parts) != 2 {
			continue
		}
		start := parseTimeOfDay(parts[0])
		end := parseTimeOfDay(parts[1])
		if start == 0 && end == 0 {
			continue
		}
		breaks = append(breaks, timeRange{Start: start, End: end})
	}
	return breaks
}

func parseTimeOfDay(raw string) time.Duration {
	parts := strings.Split(strings.TrimSpace(raw), ":")
	if len(parts) != 2 {
		return 0
	}
	hour, err := time.Parse("15:04", fmt.Sprintf("%s:%s", parts[0], parts[1]))
	if err != nil {
		return 0
	}
	return time.Duration(hour.Hour())*time.Hour + time.Duration(hour.Minute())*time.Minute
}

func timeOfDay(t time.Time) time.Duration {
	return time.Duration(t.Hour())*time.Hour + time.Duration(t.Minute())*time.Minute
}

func containsInt(values []int, target int) bool {
	for _, v := range values {
		if v == target {
			return true
		}
	}
	return false
}
