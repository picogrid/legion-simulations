package logger

import (
	"fmt"
	"strings"
)

// Icons and symbols for different log types
const (
	IconSuccess = "✅"
	IconError   = "❌"
	IconWarning = "⚠️"
	IconInfo    = "ℹ️"
	IconDebug   = "🔍"
	IconRocket  = "🚀"
	IconConfig  = "⚙️"
	IconNetwork = "🌐"
	IconTime    = "⏱️"
	IconLock    = "🔒"
	IconKey     = "🔑"
	IconUser    = "👤"
	IconFolder  = "📁"
	IconFile    = "📄"
	IconRefresh = "🔄"
	IconCheck   = "✓"
	IconCross   = "✗"
	IconDot     = "•"
	IconArrow   = "→"
)

// Success logs a success message with a green checkmark
func Success(args ...interface{}) {
	message := fmt.Sprint(args...)
	defaultLogger.Info(IconSuccess + " " + message)
}

// Successf logs a formatted success message
func Successf(format string, args ...interface{}) {
	Success(fmt.Sprintf(format, args...))
}

// Progress logs a progress message with a refresh icon
func Progress(args ...interface{}) {
	message := fmt.Sprint(args...)
	defaultLogger.Info(IconRefresh + " " + message)
}

// Progressf logs a formatted progress message
func Progressf(format string, args ...interface{}) {
	Progress(fmt.Sprintf(format, args...))
}

// Network logs a network-related message
func Network(args ...interface{}) {
	message := fmt.Sprint(args...)
	defaultLogger.Info(IconNetwork + " " + message)
}

// Networkf logs a formatted network message
func Networkf(format string, args ...interface{}) {
	Network(fmt.Sprintf(format, args...))
}

// LogSection creates a visual section separator
func LogSection(title string) {
	width := 50
	line := strings.Repeat("=", width)

	if l, ok := defaultLogger.(*logger); ok && !l.noColor {
		fmt.Println(colorCyan + line + colorReset)
		fmt.Println(colorCyan + colorBold + title + colorReset)
		fmt.Println(colorCyan + line + colorReset)
	} else {
		fmt.Println(line)
		fmt.Println(title)
		fmt.Println(line)
	}
}

// LogSubSection creates a visual subsection separator
func LogSubSection(title string) {
	width := 40
	line := strings.Repeat("-", width)

	if l, ok := defaultLogger.(*logger); ok && !l.noColor {
		fmt.Println(colorGray + line + colorReset)
		fmt.Println(colorGray + title + colorReset)
		fmt.Println(colorGray + line + colorReset)
	} else {
		fmt.Println(line)
		fmt.Println(title)
		fmt.Println(line)
	}
}

// LogList logs a list of items with bullets
func LogList(title string, items []string) {
	Info(title)
	for _, item := range items {
		fmt.Printf("  %s %s\n", IconDot, item)
	}
}

// LogKeyValue logs a key-value pair with nice formatting
func LogKeyValue(key string, value interface{}) {
	if l, ok := defaultLogger.(*logger); ok && !l.noColor {
		fmt.Printf("%s%s:%s %v\n", colorCyan, key, colorReset, value)
	} else {
		fmt.Printf("%s: %v\n", key, value)
	}
}

// LogKeyValues logs multiple key-value pairs
func LogKeyValues(pairs map[string]interface{}) {
	for k, v := range pairs {
		LogKeyValue(k, v)
	}
}

// Table represents a simple table for logging
type Table struct {
	headers []string
	rows    [][]string
}

// NewTable creates a new table
func NewTable(headers ...string) *Table {
	return &Table{
		headers: headers,
		rows:    [][]string{},
	}
}

// AddRow adds a row to the table
func (t *Table) AddRow(values ...string) {
	t.rows = append(t.rows, values)
}

// Print prints the table
func (t *Table) Print() {
	if len(t.headers) == 0 {
		return
	}

	// Calculate column widths
	widths := make([]int, len(t.headers))
	for i, h := range t.headers {
		widths[i] = len(h)
	}

	for _, row := range t.rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Print headers
	for i, h := range t.headers {
		fmt.Printf("%-*s  ", widths[i], h)
	}
	fmt.Println()

	// Print separator
	for i := range t.headers {
		fmt.Print(strings.Repeat("-", widths[i]) + "  ")
	}
	fmt.Println()

	// Print rows
	for _, row := range t.rows {
		for i, cell := range row {
			if i < len(widths) {
				fmt.Printf("%-*s  ", widths[i], cell)
			}
		}
		fmt.Println()
	}
}
