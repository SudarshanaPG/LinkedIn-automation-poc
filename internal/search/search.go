package search

import (
	"bufio"
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"linkedin-automation-poc/internal/browser"
	"linkedin-automation-poc/internal/config"
	"linkedin-automation-poc/internal/logger"
)

type Finder struct {
	cfg config.SearchConfig
	log *logger.Logger
}

func New(cfg config.SearchConfig, log *logger.Logger) *Finder {
	return &Finder{cfg: cfg, log: log}
}

func (f *Finder) Search(ctx context.Context, session *browser.Session) ([]string, error) {
	baseURL := session.BaseURL()
	results := make([]string, 0, f.cfg.MaxLeads)
	seen := map[string]struct{}{}

	leadsFile := strings.TrimSpace(f.cfg.LeadsFile)
	if leadsFile != "" {
		links, err := readLeadsFile(leadsFile, baseURL)
		if err != nil {
			return nil, err
		}
		for _, link := range links {
			if _, ok := seen[link]; ok {
				continue
			}
			seen[link] = struct{}{}
			results = append(results, link)
			if len(results) >= f.cfg.MaxLeads {
				break
			}
		}
		if f.log != nil {
			f.log.Info("loaded leads from file", map[string]any{"path": leadsFile, "leads": len(results)})
		}
		return results, nil
	}

	queries, err := buildQueries(f.cfg)
	if err != nil {
		return nil, err
	}

	for _, query := range queries {
		if err := ctx.Err(); err != nil {
			return results, err
		}
		if f.cfg.CompanyOnly {
			f.log.Debug("searching company", map[string]any{"company": query})
		}
		escaped := url.QueryEscape(query)
		for page := 1; page <= f.cfg.MaxPages; page++ {
			if err := ctx.Err(); err != nil {
				return results, err
			}
			searchURL := fmt.Sprintf("%s/search/results/people/?keywords=%s&page=%d", baseURL, escaped, page)
			if f.cfg.SortBy != "" {
				searchURL += "&sortBy=" + url.QueryEscape(f.cfg.SortBy)
			}
			if err := session.NavigateWithRetry(ctx, searchURL); err != nil {
				return results, fmt.Errorf("navigate search page: %w", err)
			}
			session.Stealth.Think(ctx)

			_ = session.Stealth.ScrollHuman(ctx, session.Page, session.Stealth.ScrollStep()*2)

			links, err := f.extractProfileLinks(ctx, session)
			if err != nil {
				return results, err
			}
			for _, link := range links {
				link = sanitizeProfileURL(link, baseURL)
				if link == "" {
					continue
				}
				if _, ok := seen[link]; ok {
					continue
				}
				seen[link] = struct{}{}
				results = append(results, link)
				if len(results) >= f.cfg.MaxLeads {
					return results, nil
				}
			}
			session.Stealth.ActionPause(ctx)
		}
	}
	return results, nil
}

func (f *Finder) extractProfileLinks(ctx context.Context, session *browser.Session) ([]string, error) {
	anchors, err := session.ElementsWithRetry(ctx, "a.app-aware-link", 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("collect search links: %w", err)
	}
	links := make([]string, 0, len(anchors))
	for _, anchor := range anchors {
		href, err := anchor.Attribute("href")
		if err != nil || href == nil {
			continue
		}
		links = append(links, *href)
	}
	return links, nil
}

func buildKeywords(cfg config.SearchConfig) string {
	parts := []string{}
	if cfg.Title != "" {
		parts = append(parts, cfg.Title)
	}
	if cfg.Company != "" {
		parts = append(parts, cfg.Company)
	}
	if cfg.Location != "" {
		parts = append(parts, cfg.Location)
	}
	for _, word := range cfg.Keywords {
		if strings.TrimSpace(word) != "" {
			parts = append(parts, word)
		}
	}
	return strings.Join(parts, " ")
}

func buildQueries(cfg config.SearchConfig) ([]string, error) {
	if cfg.CompanyOnly {
		companies := normalizeCompanies(cfg)
		if len(companies) == 0 {
			return nil, fmt.Errorf("no companies configured")
		}
		return companies, nil
	}
	keywords := buildKeywords(cfg)
	if keywords == "" {
		return nil, fmt.Errorf("no search keywords configured")
	}
	return []string{keywords}, nil
}

func normalizeCompanies(cfg config.SearchConfig) []string {
	raw := cfg.Companies
	if len(raw) == 0 && strings.TrimSpace(cfg.Company) != "" {
		raw = []string{cfg.Company}
	}
	out := make([]string, 0, len(raw))
	seen := map[string]struct{}{}
	for _, company := range raw {
		company = strings.TrimSpace(company)
		if company == "" {
			continue
		}
		key := strings.ToLower(company)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, company)
	}
	return out
}

func sanitizeProfileURL(raw string, baseURL string) string {
	if !strings.Contains(raw, "/in/") {
		return ""
	}
	if strings.Contains(raw, "miniProfile") {
		return ""
	}
	trimmed := strings.Split(raw, "?")[0]
	if !strings.HasPrefix(trimmed, "http") {
		if baseURL == "" {
			baseURL = "https://www.linkedin.com"
		}
		trimmed = strings.TrimRight(baseURL, "/") + trimmed
	}
	return trimmed
}

func readLeadsFile(path string, baseURL string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open leads file: %w", err)
	}
	defer file.Close()

	out := make([]string, 0, 128)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "/") {
			line = strings.TrimRight(baseURL, "/") + line
		}
		line = sanitizeProfileURL(line, baseURL)
		if line == "" {
			continue
		}
		out = append(out, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read leads file: %w", err)
	}
	return out, nil
}
