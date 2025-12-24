package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"linkedin-automation-poc/internal/config"
	"linkedin-automation-poc/internal/logger"
)

type State struct {
	SentRequests map[string]RequestRecord `json:"sent_requests"`
	Accepted     map[string]AcceptRecord  `json:"accepted"`
	Messages     []MessageRecord          `json:"messages"`
	LastUpdated  time.Time                `json:"last_updated"`
}

type RequestRecord struct {
	ProfileURL string    `json:"profile_url"`
	Note       string    `json:"note"`
	SentAt     time.Time `json:"sent_at"`
}

type AcceptRecord struct {
	ProfileURL string    `json:"profile_url"`
	AcceptedAt time.Time `json:"accepted_at"`
}

type MessageRecord struct {
	ProfileURL string    `json:"profile_url"`
	Template   string    `json:"template"`
	Body       string    `json:"body"`
	SentAt     time.Time `json:"sent_at"`
}

type Storage struct {
	path   string
	state  State
	mu     sync.Mutex
	dirty  bool
	ticker *time.Ticker
	done   chan struct{}
	log    *logger.Logger
}

func New(cfg config.StorageConfig, log *logger.Logger) (*Storage, error) {
	store := &Storage{
		path: cfg.Path,
		state: State{
			SentRequests: map[string]RequestRecord{},
			Accepted:     map[string]AcceptRecord{},
			Messages:     []MessageRecord{},
			LastUpdated:  time.Now(),
		},
		done: make(chan struct{}),
		log:  log,
	}
	if err := store.load(); err != nil {
		return nil, err
	}
	if cfg.FlushIntervalSeconds > 0 {
		store.ticker = time.NewTicker(time.Duration(cfg.FlushIntervalSeconds) * time.Second)
		go store.autoFlush()
	}
	return store, nil
}

func (s *Storage) load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read state: %w", err)
	}
	if err := json.Unmarshal(data, &s.state); err != nil {
		return fmt.Errorf("parse state: %w", err)
	}
	if s.state.SentRequests == nil {
		s.state.SentRequests = map[string]RequestRecord{}
	}
	if s.state.Accepted == nil {
		s.state.Accepted = map[string]AcceptRecord{}
	}
	return nil
}

func (s *Storage) autoFlush() {
	for {
		select {
		case <-s.ticker.C:
			if err := s.Save(); err != nil && s.log != nil {
				s.log.Warn("storage autosave failed", map[string]any{"error": err.Error()})
			}
		case <-s.done:
			return
		}
	}
}

func (s *Storage) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.dirty {
		return nil
	}
	s.state.LastUpdated = time.Now()
	payload, err := json.MarshalIndent(s.state, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0755); err != nil {
		return fmt.Errorf("create state dir: %w", err)
	}
	if err := os.WriteFile(s.path, payload, 0644); err != nil {
		return fmt.Errorf("write state: %w", err)
	}
	s.dirty = false
	return nil
}

func (s *Storage) Close() error {
	if s.ticker != nil {
		s.ticker.Stop()
	}
	close(s.done)
	return s.Save()
}

func (s *Storage) MarkRequestSent(profileURL, note string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.SentRequests[profileURL] = RequestRecord{
		ProfileURL: profileURL,
		Note:       note,
		SentAt:     time.Now(),
	}
	s.dirty = true
}

func (s *Storage) MarkAccepted(profileURL string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Accepted[profileURL] = AcceptRecord{
		ProfileURL: profileURL,
		AcceptedAt: time.Now(),
	}
	s.dirty = true
}

func (s *Storage) AddMessage(profileURL, template, body string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Messages = append(s.state.Messages, MessageRecord{
		ProfileURL: profileURL,
		Template:   template,
		Body:       body,
		SentAt:     time.Now(),
	})
	s.dirty = true
}

func (s *Storage) HasSent(profileURL string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.state.SentRequests[profileURL]
	return ok
}

func (s *Storage) IsAccepted(profileURL string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.state.Accepted[profileURL]
	return ok
}

func (s *Storage) PendingRequests() []RequestRecord {
	s.mu.Lock()
	defer s.mu.Unlock()
	pending := make([]RequestRecord, 0, len(s.state.SentRequests))
	for url, record := range s.state.SentRequests {
		if _, ok := s.state.Accepted[url]; ok {
			continue
		}
		pending = append(pending, record)
	}
	return pending
}

func (s *Storage) AcceptedConnections() []AcceptRecord {
	s.mu.Lock()
	defer s.mu.Unlock()
	accepted := make([]AcceptRecord, 0, len(s.state.Accepted))
	for _, record := range s.state.Accepted {
		accepted = append(accepted, record)
	}
	return accepted
}

func (s *Storage) CountRequestsSince(since time.Time) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	count := 0
	for _, record := range s.state.SentRequests {
		if record.SentAt.After(since) {
			count++
		}
	}
	return count
}

func (s *Storage) CountMessagesSince(since time.Time) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	count := 0
	for _, record := range s.state.Messages {
		if record.SentAt.After(since) {
			count++
		}
	}
	return count
}

func (s *Storage) HasMessaged(profileURL string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, record := range s.state.Messages {
		if record.ProfileURL == profileURL {
			return true
		}
	}
	return false
}

func (s *Storage) LastMessageAt(profileURL string) time.Time {
	s.mu.Lock()
	defer s.mu.Unlock()
	var latest time.Time
	for _, record := range s.state.Messages {
		if record.ProfileURL == profileURL && record.SentAt.After(latest) {
			latest = record.SentAt
		}
	}
	return latest
}
