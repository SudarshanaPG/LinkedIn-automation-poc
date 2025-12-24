package profile

import "strings"

func ApplyTemplate(template string, profile Profile) string {
	if template == "" {
		return ""
	}
	replacements := profile.Variables()
	replacer := strings.NewReplacer(
		"{{FullName}}", replacements["FullName"],
		"{{FirstName}}", replacements["FirstName"],
		"{{LastName}}", replacements["LastName"],
		"{{Headline}}", replacements["Headline"],
		"{{Company}}", replacements["Company"],
		"{{Industry}}", replacements["Industry"],
	)
	return strings.TrimSpace(replacer.Replace(template))
}
