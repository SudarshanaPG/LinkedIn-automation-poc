package browser

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func FindBrowserBinary() string {
	if fromEnv := strings.TrimSpace(os.Getenv("LINKEDIN_BROWSER_PATH")); fromEnv != "" {
		if resolved := resolvePathOrCommand(fromEnv); resolved != "" {
			return resolved
		}
	}

	candidates := []string{}

	if runtime.GOOS == "windows" {
		programFiles := os.Getenv("PROGRAMFILES")
		programFilesX86 := os.Getenv("PROGRAMFILES(X86)")
		localAppData := os.Getenv("LOCALAPPDATA")

		candidates = append(candidates,
			filepath.Join(programFiles, "Google", "Chrome", "Application", "chrome.exe"),
			filepath.Join(programFilesX86, "Google", "Chrome", "Application", "chrome.exe"),
			filepath.Join(programFiles, "Microsoft", "Edge", "Application", "msedge.exe"),
			filepath.Join(programFilesX86, "Microsoft", "Edge", "Application", "msedge.exe"),
			filepath.Join(programFiles, "BraveSoftware", "Brave-Browser", "Application", "brave.exe"),
			filepath.Join(programFilesX86, "BraveSoftware", "Brave-Browser", "Application", "brave.exe"),
			filepath.Join(localAppData, "Google", "Chrome", "Application", "chrome.exe"),
			filepath.Join(localAppData, "Microsoft", "Edge", "Application", "msedge.exe"),
			filepath.Join(localAppData, "BraveSoftware", "Brave-Browser", "Application", "brave.exe"),
		)
	} else {
		candidates = append(candidates,
			"google-chrome",
			"chrome",
			"chromium",
			"chromium-browser",
			"brave",
			"microsoft-edge",
		)
	}

	for _, candidate := range candidates {
		if resolved := resolvePathOrCommand(candidate); resolved != "" {
			return resolved
		}
	}

	return ""
}

func resolvePathOrCommand(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if fileExists(value) {
		return value
	}
	if resolved, err := exec.LookPath(value); err == nil && resolved != "" {
		return resolved
	}
	return ""
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
