package stealth

import (
	"fmt"
	"strings"

	"linkedin-automation-poc/internal/config"
)

func FingerprintScript(cfg config.StealthConfig) string {
	languages := []string{cfg.Locale, strings.Split(cfg.Locale, "-")[0]}
	languageArray := fmt.Sprintf("['%s','%s']", languages[0], languages[1])
	return fmt.Sprintf(`() => {
    Object.defineProperty(navigator, 'webdriver', { get: () => undefined });
    Object.defineProperty(navigator, 'languages', { get: () => %s });
    Object.defineProperty(navigator, 'plugins', { get: () => [1, 2, 3, 4, 5] });
    Object.defineProperty(navigator, 'platform', { get: () => 'Win32' });
    window.chrome = { runtime: {} };
    const originalQuery = window.navigator.permissions.query;
    window.navigator.permissions.query = (parameters) => (
        parameters && parameters.name === 'notifications'
            ? Promise.resolve({ state: Notification.permission })
            : originalQuery(parameters)
    );
    const getParameter = WebGLRenderingContext.prototype.getParameter;
    WebGLRenderingContext.prototype.getParameter = function(parameter) {
        if (parameter === 37445) return 'Intel Inc.';
        if (parameter === 37446) return 'Intel Iris OpenGL Engine';
        return getParameter.apply(this, [parameter]);
    };
}`, languageArray)
}
