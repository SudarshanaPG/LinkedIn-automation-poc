package profile

import (
	"strings"

	"github.com/go-rod/rod"
)

type Profile struct {
	FullName string
	First    string
	Last     string
	Headline string
	Company  string
	Industry string
}

func FromPage(page *rod.Page) (Profile, error) {
	name := textOrEmpty(page, "h1")
	headline := textOrEmpty(page, "div.text-body-medium")
	first, last := splitName(name)
	company := extractCompany(headline)
	return Profile{
		FullName: name,
		First:    first,
		Last:     last,
		Headline: headline,
		Company:  company,
	}, nil
}

func (p Profile) Variables() map[string]string {
	return map[string]string{
		"FullName":  p.FullName,
		"FirstName": p.First,
		"LastName":  p.Last,
		"Headline":  p.Headline,
		"Company":   p.Company,
		"Industry":  p.Industry,
	}
}

func textOrEmpty(page *rod.Page, selector string) string {
	el, err := page.Element(selector)
	if err != nil {
		return ""
	}
	text, err := el.Text()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(text)
}

func splitName(full string) (string, string) {
	parts := strings.Fields(full)
	if len(parts) == 0 {
		return "", ""
	}
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], parts[len(parts)-1]
}

func extractCompany(headline string) string {
	lower := strings.ToLower(headline)
	if idx := strings.Index(lower, " at "); idx != -1 {
		return strings.TrimSpace(headline[idx+4:])
	}
	return ""
}
