package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"linkedin-automation-poc/internal/config"
)

type Level int

const (
	Debug Level = iota
	Info
	Warn
	Error
)

type Logger struct {
	level  Level
	format string
	out    io.Writer
	mu     *sync.Mutex
	fields map[string]any
}

func New(cfg config.LoggingConfig) (*Logger, error) {
	level := parseLevel(cfg.Level)
	format := strings.ToLower(cfg.Format)
	var out io.Writer = os.Stdout
	if cfg.Path != "" {
		file, err := os.OpenFile(cfg.Path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("open log file: %w", err)
		}
		out = file
	}
	return &Logger{
		level:  level,
		format: format,
		out:    out,
		fields: map[string]any{},
		mu:     &sync.Mutex{},
	}, nil
}

func (l *Logger) Debug(msg string, fields map[string]any) {
	l.log(Debug, msg, fields)
}

func (l *Logger) Info(msg string, fields map[string]any) {
	l.log(Info, msg, fields)
}

func (l *Logger) Warn(msg string, fields map[string]any) {
	l.log(Warn, msg, fields)
}

func (l *Logger) Error(msg string, fields map[string]any) {
	l.log(Error, msg, fields)
}

func (l *Logger) log(level Level, msg string, fields map[string]any) {
	if level < l.level {
		return
	}
	if fields == nil {
		fields = map[string]any{}
	}
	entry := map[string]any{
		"time":  time.Now().Format(time.RFC3339Nano),
		"level": level.String(),
		"msg":   msg,
	}
	for k, v := range l.fields {
		entry[k] = v
	}
	for k, v := range fields {
		entry[k] = v
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	switch l.format {
	case "json":
		enc := json.NewEncoder(l.out)
		_ = enc.Encode(entry)
	default:
		fmt.Fprintln(l.out, formatText(entry))
	}
}

func (l *Logger) With(fields map[string]any) *Logger {
	if len(fields) == 0 {
		return l
	}
	merged := make(map[string]any, len(l.fields)+len(fields))
	for k, v := range l.fields {
		merged[k] = v
	}
	for k, v := range fields {
		merged[k] = v
	}
	return &Logger{
		level:  l.level,
		format: l.format,
		out:    l.out,
		fields: merged,
		mu:     l.mu,
	}
}

func formatText(entry map[string]any) string {
	parts := make([]string, 0, len(entry))
	for _, key := range []string{"time", "level", "msg"} {
		if value, ok := entry[key]; ok {
			parts = append(parts, fmt.Sprintf("%s=%v", key, value))
		}
	}
	for key, value := range entry {
		if key == "time" || key == "level" || key == "msg" {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s=%v", key, value))
	}
	return strings.Join(parts, " ")
}

func parseLevel(raw string) Level {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "debug":
		return Debug
	case "warn", "warning":
		return Warn
	case "error":
		return Error
	default:
		return Info
	}
}

func (l Level) String() string {
	switch l {
	case Debug:
		return "debug"
	case Warn:
		return "warn"
	case Error:
		return "error"
	default:
		return "info"
	}
}
