package handlers

import (
	"os"
	"strconv"
)

// Handlers holds all handler instances with their dependencies
type Handlers struct {
	Repo       DatabaseRepository
	SystemPort int
}

// GlobalSystemPort stores the port the server is running on
var GlobalSystemPort int = 15678

// SetSystemPort sets the global system port
func SetSystemPort(port int) {
	if port > 0 {
		GlobalSystemPort = port
		return
	}

	if envPort := os.Getenv("PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil && p > 0 {
			GlobalSystemPort = p
			return
		}
	}

	GlobalSystemPort = 15678
}

// NewHandlers creates a new Handlers instance with the given repository
func NewHandlers(repo DatabaseRepository) *Handlers {
	return &Handlers{
		Repo: repo,
	}
}

// NewDefaultHandlers creates a new Handlers instance with the default repository implementation
func NewDefaultHandlers() *Handlers {
	return NewHandlers(&DefaultRepository{})
}
