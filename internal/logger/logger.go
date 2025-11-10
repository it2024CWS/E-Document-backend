package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var Logger zerolog.Logger

// LogLevel represents the logging level
type LogLevel string

const (
	DebugLevel LogLevel = "debug"
	InfoLevel  LogLevel = "info"
	WarnLevel  LogLevel = "warn"
	ErrorLevel LogLevel = "error"
)

// Config holds logger configuration
type Config struct {
	Level      LogLevel
	Pretty     bool   // Pretty print for development
	TimeFormat string // Time format for logs
}

// Init initializes the global logger
func Init(cfg Config) {
	// Set log level
	level := parseLogLevel(cfg.Level)
	zerolog.SetGlobalLevel(level)

	// Configure output
	var output io.Writer
	if cfg.Pretty {
		// Pretty console output for development
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: cfg.TimeFormat,
			FormatLevel: func(i interface{}) string {
				return formatLevel(i)
			},
		}
	} else {
		// JSON output for production
		output = os.Stdout
	}

	// Create logger
	Logger = zerolog.New(output).With().Timestamp().Caller().Logger()
	log.Logger = Logger
}

// parseLogLevel converts string to zerolog level
func parseLogLevel(level LogLevel) zerolog.Level {
	switch level {
	case DebugLevel:
		return zerolog.DebugLevel
	case InfoLevel:
		return zerolog.InfoLevel
	case WarnLevel:
		return zerolog.WarnLevel
	case ErrorLevel:
		return zerolog.ErrorLevel
	default:
		return zerolog.InfoLevel
	}
}

// formatLevel adds colors to log levels
func formatLevel(i interface{}) string {
	var l string
	if ll, ok := i.(string); ok {
		switch ll {
		case "debug":
			l = "\033[36mDEBUG\033[0m" // Cyan
		case "info":
			l = "\033[32mINFO\033[0m" // Green
		case "warn":
			l = "\033[33mWARN\033[0m" // Yellow
		case "error":
			l = "\033[31mERROR\033[0m" // Red
		case "fatal":
			l = "\033[35mFATAL\033[0m" // Magenta
		default:
			l = ll
		}
	}
	return l
}

// Debug logs a debug message
func Debug(msg string) {
	Logger.Debug().Msg(msg)
}

// Debugf logs a formatted debug message
func Debugf(format string, args ...interface{}) {
	Logger.Debug().Msgf(format, args...)
}

// Info logs an info message
func Info(msg string) {
	Logger.Info().Msg(msg)
}

// Infof logs a formatted info message
func Infof(format string, args ...interface{}) {
	Logger.Info().Msgf(format, args...)
}

// Warn logs a warning message
func Warn(msg string) {
	Logger.Warn().Msg(msg)
}

// Warnf logs a formatted warning message
func Warnf(format string, args ...interface{}) {
	Logger.Warn().Msgf(format, args...)
}

// Error logs an error message
func Error(msg string) {
	Logger.Error().Msg(msg)
}

// Errorf logs a formatted error message
func Errorf(format string, args ...interface{}) {
	Logger.Error().Msgf(format, args...)
}

// ErrorWithErr logs an error message with error details
func ErrorWithErr(msg string, err error) {
	Logger.Error().Err(err).Msg(msg)
}

// Fatal logs a fatal error and exits
func Fatal(msg string) {
	Logger.Fatal().Msg(msg)
}

// Fatalf logs a formatted fatal error and exits
func Fatalf(format string, args ...interface{}) {
	Logger.Fatal().Msgf(format, args...)
}

// FatalWithErr logs a fatal error with error details and exits
func FatalWithErr(msg string, err error) {
	Logger.Fatal().Err(err).Msg(msg)
}

// WithField adds a field to the log context
func WithField(key string, value interface{}) *zerolog.Event {
	return Logger.Info().Interface(key, value)
}

// WithFields adds multiple fields to the log context
func WithFields(fields map[string]interface{}) *zerolog.Event {
	event := Logger.Info()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	return event
}

// NewDefaultConfig returns default logger configuration
func NewDefaultConfig() Config {
	return Config{
		Level:      InfoLevel,
		Pretty:     true,
		TimeFormat: time.RFC3339,
	}
}
