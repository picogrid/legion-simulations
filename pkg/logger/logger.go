package logger

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

// Level represents the severity of a log message
type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

// Color codes for different log levels
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[90m"
	colorBold   = "\033[1m"
)

// Logger is the main logger interface
type Logger interface {
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	WithField(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger
	WithPrefix(prefix string) Logger
}

// logger implements the Logger interface
type logger struct {
	mu       sync.Mutex
	level    Level
	writer   io.Writer
	fields   map[string]interface{}
	prefix   string
	noColor  bool
	showTime bool
}

// Default logger instance
var defaultLogger = New()

// Config holds logger configuration
type Config struct {
	Level    Level
	Writer   io.Writer
	NoColor  bool
	ShowTime bool
}

// New creates a new logger with default configuration
func New() Logger {
	return NewWithConfig(Config{
		Level:    InfoLevel,
		Writer:   os.Stdout,
		NoColor:  false,
		ShowTime: true,
	})
}

// NewWithConfig creates a new logger with custom configuration
func NewWithConfig(cfg Config) Logger {
	return &logger{
		level:    cfg.Level,
		writer:   cfg.Writer,
		fields:   make(map[string]interface{}),
		noColor:  cfg.NoColor,
		showTime: cfg.ShowTime,
	}
}

// SetLevel sets the global log level
func SetLevel(level Level) {
	if l, ok := defaultLogger.(*logger); ok {
		l.mu.Lock()
		l.level = level
		l.mu.Unlock()
	}
}

// SetNoColor disables color output
func SetNoColor(noColor bool) {
	if l, ok := defaultLogger.(*logger); ok {
		l.mu.Lock()
		l.noColor = noColor
		l.mu.Unlock()
	}
}

// Helper methods for the default logger
func Debug(args ...interface{})                       { defaultLogger.Debug(args...) }
func Debugf(format string, args ...interface{})       { defaultLogger.Debugf(format, args...) }
func Info(args ...interface{})                        { defaultLogger.Info(args...) }
func Infof(format string, args ...interface{})        { defaultLogger.Infof(format, args...) }
func Warn(args ...interface{})                        { defaultLogger.Warn(args...) }
func Warnf(format string, args ...interface{})        { defaultLogger.Warnf(format, args...) }
func Error(args ...interface{})                       { defaultLogger.Error(args...) }
func Errorf(format string, args ...interface{})       { defaultLogger.Errorf(format, args...) }
func Fatal(args ...interface{})                       { defaultLogger.Fatal(args...) }
func Fatalf(format string, args ...interface{})       { defaultLogger.Fatalf(format, args...) }
func WithField(key string, value interface{}) Logger  { return defaultLogger.WithField(key, value) }
func WithFields(fields map[string]interface{}) Logger { return defaultLogger.WithFields(fields) }
func WithPrefix(prefix string) Logger                 { return defaultLogger.WithPrefix(prefix) }

// Implementation of logger methods

func (l *logger) log(level Level, args ...interface{}) {
	if level < l.level {
		return
	}

	l.mu.Lock()

	// Build the log message
	var parts []string

	// Add timestamp if enabled
	if l.showTime {
		timestamp := time.Now().Format("15:04:05")
		if l.noColor {
			parts = append(parts, timestamp)
		} else {
			parts = append(parts, colorGray+timestamp+colorReset)
		}
	}

	// Add level
	levelStr, levelColor := l.getLevelString(level)
	if l.noColor {
		parts = append(parts, levelStr)
	} else {
		parts = append(parts, levelColor+levelStr+colorReset)
	}

	// Add prefix if set
	if l.prefix != "" {
		if l.noColor {
			parts = append(parts, "["+l.prefix+"]")
		} else {
			parts = append(parts, colorCyan+"["+l.prefix+"]"+colorReset)
		}
	}

	// Add fields if any
	if len(l.fields) > 0 {
		var fieldParts []string
		for k, v := range l.fields {
			fieldParts = append(fieldParts, fmt.Sprintf("%s=%v", k, v))
		}
		fieldsStr := strings.Join(fieldParts, " ")
		if l.noColor {
			parts = append(parts, fieldsStr)
		} else {
			parts = append(parts, colorGray+fieldsStr+colorReset)
		}
	}

	// Add message
	message := fmt.Sprint(args...)
	parts = append(parts, message)

	// Write to output
	_, _ = fmt.Fprintln(l.writer, strings.Join(parts, " "))

	l.mu.Unlock()

	// Exit on fatal (after unlocking mutex)
	if level == FatalLevel {
		os.Exit(1)
	}
}

func (l *logger) logf(level Level, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	l.log(level, message)
}

func (l *logger) getLevelString(level Level) (string, string) {
	switch level {
	case DebugLevel:
		return "DEBUG", colorGray
	case InfoLevel:
		return "INFO ", colorGreen
	case WarnLevel:
		return "WARN ", colorYellow
	case ErrorLevel:
		return "ERROR", colorRed
	case FatalLevel:
		return "FATAL", colorRed + colorBold
	default:
		return "UNKNOWN", colorReset
	}
}

// Logger interface implementation

func (l *logger) Debug(args ...interface{}) {
	l.log(DebugLevel, args...)
}

func (l *logger) Debugf(format string, args ...interface{}) {
	l.logf(DebugLevel, format, args...)
}

func (l *logger) Info(args ...interface{}) {
	l.log(InfoLevel, args...)
}

func (l *logger) Infof(format string, args ...interface{}) {
	l.logf(InfoLevel, format, args...)
}

func (l *logger) Warn(args ...interface{}) {
	l.log(WarnLevel, args...)
}

func (l *logger) Warnf(format string, args ...interface{}) {
	l.logf(WarnLevel, format, args...)
}

func (l *logger) Error(args ...interface{}) {
	l.log(ErrorLevel, args...)
}

func (l *logger) Errorf(format string, args ...interface{}) {
	l.logf(ErrorLevel, format, args...)
}

func (l *logger) Fatal(args ...interface{}) {
	l.log(FatalLevel, args...)
}

func (l *logger) Fatalf(format string, args ...interface{}) {
	l.logf(FatalLevel, format, args...)
}

func (l *logger) WithField(key string, value interface{}) Logger {
	newLogger := &logger{
		level:    l.level,
		writer:   l.writer,
		fields:   make(map[string]interface{}),
		prefix:   l.prefix,
		noColor:  l.noColor,
		showTime: l.showTime,
	}

	// Copy existing fields
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}

	// Add new field
	newLogger.fields[key] = value

	return newLogger
}

func (l *logger) WithFields(fields map[string]interface{}) Logger {
	newLogger := &logger{
		level:    l.level,
		writer:   l.writer,
		fields:   make(map[string]interface{}),
		prefix:   l.prefix,
		noColor:  l.noColor,
		showTime: l.showTime,
	}

	// Copy existing fields
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}

	// Add new fields
	for k, v := range fields {
		newLogger.fields[k] = v
	}

	return newLogger
}

func (l *logger) WithPrefix(prefix string) Logger {
	newLogger := &logger{
		level:    l.level,
		writer:   l.writer,
		fields:   make(map[string]interface{}),
		prefix:   prefix,
		noColor:  l.noColor,
		showTime: l.showTime,
	}

	// Copy existing fields
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}

	return newLogger
}

// ParseLevel parses a string log level
func ParseLevel(level string) Level {
	switch strings.ToLower(level) {
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	case "warn", "warning":
		return WarnLevel
	case "error":
		return ErrorLevel
	case "fatal":
		return FatalLevel
	default:
		return InfoLevel
	}
}
