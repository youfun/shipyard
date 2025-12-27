package cliutils

import (
	"log"
	"os"
	"strings"
)

// DetectRuntime attempts to detect the application's runtime based on common project files.
func DetectRuntime() string {
	if _, err := os.Stat("mix.exs"); err == nil {
		// Check if it's a Phoenix project by looking for :phoenix dependency
		content, readErr := os.ReadFile("mix.exs")
		if readErr == nil && strings.Contains(string(content), ":phoenix") {
			log.Println("Detected Phoenix project (mix.exs with :phoenix).")
			return "phoenix"
		}
		// Pure Elixir project (no Phoenix)
		log.Println("Detected Elixir project (mix.exs without Phoenix).")
		return "elixir"
	}
	if _, err := os.Stat("package.json"); err == nil {
		log.Println("Detected Node.js project (package.json).")
		return "node"
	}
	if _, err := os.Stat("go.mod"); err == nil {
		log.Println("Detected Go project (go.mod).")
		return "golang"
	}
	if _, err := os.Stat("Dockerfile"); err == nil {
		log.Println("Detected Dockerfile.")
		return "docker"
	}
	log.Println("Could not automatically detect runtime; defaulting to 'elixir'.")
	return "elixir" // Default to elixir if nothing detected
}
