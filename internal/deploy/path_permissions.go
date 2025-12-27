package deploy

import (
	"fmt"
	"log" // Note: Use path instead of path/filepath, as remote servers are Linux and require / separators
	"path"
	"strings"
)

// pathSuffixes are common suffixes that indicate path-related variables
var pathSuffixes = []string{"_PATH", "_DIR", "_FILE", "_DB", "_DATABASE"}

// shellEscape escapes a string for safe use in shell commands
// This prevents command injection vulnerabilities
func shellEscape(s string) string {
	// Use single quotes and escape any single quotes in the string
	// This is the safest way to pass arbitrary strings to shell
	escaped := strings.ReplaceAll(s, "'", "'\"'\"'")
	return fmt.Sprintf("'%s'", escaped)
}

// isLocalPath determines if a value represents a local filesystem path
// rather than a URL or other non-path value.
// It returns true if the value looks like a local path (starts with / or ./ or ../)
// and false if it looks like a URL or other value.
func isLocalPath(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return false
	}

	// URLs typically have scheme://
	if strings.Contains(value, "://") {
		return false
	}

	// Check if it starts with path indicators
	if strings.HasPrefix(value, "/") {
		return true
	}
	if strings.HasPrefix(value, "./") || strings.HasPrefix(value, "../") {
		return true
	}

	// Check if it looks like a path with directory separators
	// But avoid false positives like "key=value" or "localhost:5432"
	if strings.Contains(value, "/") && !strings.Contains(value, ":") {
		return true
	}

	return false
}

// extractPathsFromEnvVars scans environment variables and extracts those
// that appear to be local filesystem paths.
// It returns a map of variable names to their path values.
func extractPathsFromEnvVars(envs map[string]interface{}) map[string]string {
	paths := make(map[string]string)

	for key, value := range envs {
		// Convert value to string
		valueStr, ok := value.(string)
		if !ok {
			continue
		}

		// Check if variable name suggests it's a path
		keyUpper := strings.ToUpper(key)
		isPathVar := false
		for _, suffix := range pathSuffixes {
			if strings.HasSuffix(keyUpper, suffix) {
				isPathVar = true
				break
			}
		}

		// If it looks like a path variable and the value is a local path
		if isPathVar && isLocalPath(valueStr) {
			paths[key] = valueStr
		}
	}

	return paths
}

// ensurePathPermissions ensures that directories specified in environment
// variables exist with proper permissions for the application to create files.
// This is called before starting the application to prevent "readonly database"
// or similar permission errors.
//
// Note: We only create directories, NOT files. Applications like Phoenix/SQLite
// will create their own database files. We just need to ensure the parent
// directories exist and have correct ownership/permissions.
func (d *Deployer) ensurePathPermissions(envs map[string]interface{}) error {
	paths := extractPathsFromEnvVars(envs)

	if len(paths) == 0 {
		log.Println("No path-related environment variables detected, skipping path permission setup")
		return nil
	}

	log.Printf("Found %d path-related environment variables, ensuring directories and permissions...", len(paths))

	// Determine the user:group for file ownership
	// Default to phoenix:phoenix, but could be made configurable in the future
	ownerUser := "phoenix"
	ownerGroup := "phoenix"

	for varName, pathValue := range paths {
		log.Printf("  [DEBUG] Processing env var: %s=%s", varName, pathValue)

		// Get the directory path
		// If pathValue looks like a file (has extension), get its parent directory
		// Otherwise, treat the whole path as a directory
		dir := pathValue
		if path.Ext(pathValue) != "" {
			dir = path.Dir(pathValue)
			log.Printf("  [DEBUG] Path has extension, using parent directory: %s", dir)
		} else {
			log.Printf("  [DEBUG] Path has no extension, treating as directory: %s", dir)
		}

		// Step 1: Create directory with sudo (system directories require root)
		// Use mkdir -p to create parent directories
		createDirCmd := fmt.Sprintf("sudo mkdir -p %s", shellEscape(dir))
		log.Printf("  [DEBUG] Executing: %s", createDirCmd)
		if err := d.executeRemoteCommand(createDirCmd, true); err != nil {
			log.Printf("  ❌ [ERROR] Failed to create directory %s", dir)
			log.Printf("  ❌ [ERROR] Command: %s", createDirCmd)
			log.Printf("  ❌ [ERROR] Error: %v", err)
			continue
		}
		log.Printf("  [DEBUG] mkdir command succeeded")

		// Verify directory was created
		verifyCmd := fmt.Sprintf("ls -ld %s", shellEscape(dir))
		log.Printf("  [DEBUG] Verifying directory exists: %s", verifyCmd)
		if err := d.executeRemoteCommand(verifyCmd, false); err != nil {
			log.Printf("  ❌ [ERROR] Directory still does not exist after mkdir: %s", dir)
		}

		// Step 2: Set ownership with sudo so phoenix user can write to it
		chownCmd := fmt.Sprintf("sudo chown -R %s:%s %s", ownerUser, ownerGroup, shellEscape(dir))
		log.Printf("  [DEBUG] Executing: %s", chownCmd)
		if err := d.executeRemoteCommand(chownCmd, false); err != nil {
			log.Printf("  ⚠️ [WARN] Failed to set ownership for %s: %v", dir, err)
		}

		// Step 3: Set directory permissions (775 so group can write)
		// This is important for directories containing databases
		chmodCmd := fmt.Sprintf("sudo chmod -R 775 %s", shellEscape(dir))
		log.Printf("  [DEBUG] Executing: %s", chmodCmd)
		if err := d.executeRemoteCommand(chmodCmd, false); err != nil {
			log.Printf("  ⚠️ [WARN] Failed to set permissions for %s: %v", dir, err)
		}

		// Note: We do NOT create the actual database file here.
		// Phoenix/SQLite will create it themselves when needed.
		// We just ensure the directory exists and is writable.

		log.Printf("  ✅ Ensured directory exists and has correct permissions: %s", dir)
	}

	log.Println("✅ Path permissions setup completed")
	return nil
}
