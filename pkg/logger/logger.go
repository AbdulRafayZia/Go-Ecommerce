package logger

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	// CorrelationIDKey is the context key for correlation/trace ID
	CorrelationIDKey contextKey = "correlation_id"
)

// Logger wraps zerolog.Logger with additional functionality
type Logger struct {
	logger zerolog.Logger
}

// Config holds logger configuration
type Config struct {
	Level      string
	Output     io.Writer
	PrettyPrint bool
}

// New creates a new Logger instance
func New(cfg Config) *Logger {
	// Set log level
	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	// Configure output
	output := cfg.Output
	if output == nil {
		output = os.Stdout
	}

	// Use pretty logging for development
	if cfg.PrettyPrint {
		output = zerolog.ConsoleWriter{
			Out:        output,
			TimeFormat: time.RFC3339,
		}
	}

	// Create logger with common fields
	log := zerolog.New(output).
		With().
		Timestamp().
		Caller().
		Logger()

	return &Logger{logger: log}
}

// NewDefault creates a logger with default configuration
func NewDefault() *Logger {
	return New(Config{
		Level:       "info",
		Output:      os.Stdout,
		PrettyPrint: false,
	})
}

// WithCorrelationID adds correlation ID to the logger from context
func (l *Logger) WithCorrelationID(ctx context.Context) *Logger {
	if correlationID, ok := ctx.Value(CorrelationIDKey).(string); ok && correlationID != "" {
		return &Logger{
			logger: l.logger.With().Str("correlation_id", correlationID).Logger(),
		}
	}
	return l
}

// WithField adds a field to the logger
func (l *Logger) WithField(key string, value interface{}) *Logger {
	return &Logger{
		logger: l.logger.With().Interface(key, value).Logger(),
	}
}

// WithFields adds multiple fields to the logger
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	ctx := l.logger.With()
	for k, v := range fields {
		ctx = ctx.Interface(k, v)
	}
	return &Logger{logger: ctx.Logger()}
}

// Debug logs a debug message
func (l *Logger) Debug(msg string) {
	l.logger.Debug().Msg(msg)
}

// Debugf logs a formatted debug message
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.logger.Debug().Msgf(format, args...)
}

// Info logs an info message
func (l *Logger) Info(msg string) {
	l.logger.Info().Msg(msg)
}

// Infof logs a formatted info message
func (l *Logger) Infof(format string, args ...interface{}) {
	l.logger.Info().Msgf(format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string) {
	l.logger.Warn().Msg(msg)
}

// Warnf logs a formatted warning message
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.logger.Warn().Msgf(format, args...)
}

// Error logs an error message
func (l *Logger) Error(msg string) {
	l.logger.Error().Msg(msg)
}

// Errorf logs a formatted error message
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.logger.Error().Msgf(format, args...)
}

// ErrorWithErr logs an error message with an error object
func (l *Logger) ErrorWithErr(err error, msg string) {
	l.logger.Error().Err(err).Msg(msg)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string) {
	l.logger.Fatal().Msg(msg)
}

// Fatalf logs a formatted fatal message and exits
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.logger.Fatal().Msgf(format, args...)
}

// FatalWithErr logs a fatal message with an error object and exits
func (l *Logger) FatalWithErr(err error, msg string) {
	l.logger.Fatal().Err(err).Msg(msg)
}

// GetZerolog returns the underlying zerolog.Logger for advanced usage
func (l *Logger) GetZerolog() *zerolog.Logger {
	return &l.logger
}
