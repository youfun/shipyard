package logs

import (
	"encoding/json"
	"regexp"
	"strings"
)

// ANSI color codes
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorGray   = "\033[37m"
)

// Precompiled regex (performance optimization)
var (
	// Log levels
	regexError = regexp.MustCompile(`(?i)\b(error|err|fatal|fail|failed|failure)\b`)
	regexWarn  = regexp.MustCompile(`(?i)\b(warn|warning|caution)\b`)
	regexInfo  = regexp.MustCompile(`(?i)\b(info|information)\b`)
	regexDebug = regexp.MustCompile(`(?i)\b(debug|trace)\b`)

	// Special patterns
	regexIP        = regexp.MustCompile(`\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\b`)
	regexURL       = regexp.MustCompile(`https?://[^\s]+`)
	regexTimestamp = regexp.MustCompile(`\d{4}-\d{2}-\d{2}[T ]\d{2}:\d{2}:\d{2}`)
	regexNumber    = regexp.MustCompile(`\b\d+\b`)
)

// ParseAndColorLine parses a single log line and adds ANSI colors.
// Supports:
//   - JSON format log auto-beautification
//   - Keyword highlighting (ERROR, WARN, INFO)
//   - IP address, URL, timestamp coloring
func ParseAndColorLine(line string, colorEnabled bool) string {
	if !colorEnabled {
		return line
	}

	// Detect if it's JSON format
	if strings.HasPrefix(strings.TrimSpace(line), "{") {
		if coloredJSON := tryFormatJSON(line); coloredJSON != "" {
			return coloredJSON
		}
	}

	// Colorize by priority
	colored := line

	// 1. Log level (Highest priority)
	colored = regexError.ReplaceAllStringFunc(colored, func(match string) string {
		return ColorRed + match + ColorReset
	})
	colored = regexWarn.ReplaceAllStringFunc(colored, func(match string) string {
		return ColorYellow + match + ColorReset
	})
	colored = regexInfo.ReplaceAllStringFunc(colored, func(match string) string {
		return ColorGreen + match + ColorReset
	})
	colored = regexDebug.ReplaceAllStringFunc(colored, func(match string) string {
		return ColorGray + match + ColorReset
	})

	// 2. URL (Before IP to avoid conflict)
	colored = regexURL.ReplaceAllStringFunc(colored, func(match string) string {
		return ColorBlue + match + ColorReset
	})

	// 3. IP Address
	colored = regexIP.ReplaceAllStringFunc(colored, func(match string) string {
		return ColorCyan + match + ColorReset
	})

	// 4. Timestamp
	colored = regexTimestamp.ReplaceAllStringFunc(colored, func(match string) string {
		return ColorGray + match + ColorReset
	})

	return colored
}

// tryFormatJSON tries to parse the string as JSON and beautify the output
func tryFormatJSON(line string) string {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(line), &data); err != nil {
		return "" // Not valid JSON
	}

	// Beautify JSON
	formatted, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return ""
	}

	// Colorize JSON
	result := string(formatted)

	// Colorize JSON keys (Use Cyan)
	result = regexp.MustCompile(`"([^"]+)":`).ReplaceAllStringFunc(result, func(match string) string {
		return ColorCyan + match + ColorReset
	})

	// Colorize special values
	if level, ok := data["level"]; ok {
		levelStr := strings.ToLower(level.(string))
		switch {
		case strings.Contains(levelStr, "error") || strings.Contains(levelStr, "fatal"):
			result = strings.Replace(result, `"`+level.(string)+`"`, ColorRed+`"`+level.(string)+`"`+ColorReset, 1)
		case strings.Contains(levelStr, "warn"):
			result = strings.Replace(result, `"`+level.(string)+`"`, ColorYellow+`"`+level.(string)+`"`+ColorReset, 1)
		case strings.Contains(levelStr, "info"):
			result = strings.Replace(result, `"`+level.(string)+`"`, ColorGreen+`"`+level.(string)+`"`+ColorReset, 1)
		}
	}

	return result
}

// ColorizeLogLevel adds color to log level strings
func ColorizeLogLevel(level string) string {
	switch strings.ToUpper(level) {
	case "ERROR", "FATAL", "FAIL":
		return ColorRed + level + ColorReset
	case "WARN", "WARNING":
		return ColorYellow + level + ColorReset
	case "INFO":
		return ColorGreen + level + ColorReset
	case "DEBUG", "TRACE":
		return ColorGray + level + ColorReset
	default:
		return level
	}
}

// ColorizeStatus adds color to deployment status strings
func ColorizeStatus(status string) string {
	switch strings.ToLower(status) {
	case "running", "active":
		return ColorGreen + status + ColorReset
	case "stopped", "inactive":
		return ColorGray + status + ColorReset
	case "failed", "error":
		return ColorRed + status + ColorReset
	case "pending", "standby":
		return ColorYellow + status + ColorReset
	default:
		return status
	}
}
