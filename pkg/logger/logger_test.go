package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name   string
		config Config
	}{
		{
			name: "default config",
			config: Config{
				Level:       "info",
				PrettyPrint: false,
			},
		},
		{
			name: "debug level",
			config: Config{
				Level:       "debug",
				PrettyPrint: false,
			},
		},
		{
			name: "invalid level defaults to info",
			config: Config{
				Level:       "invalid",
				PrettyPrint: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			tt.config.Output = buf
			logger := New(tt.config)
			if logger == nil {
				t.Fatal("expected logger to be non-nil")
			}
		})
	}
}

func TestLogger_WithCorrelationID(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(Config{
		Level:       "info",
		Output:      buf,
		PrettyPrint: false,
	})

	correlationID := "test-correlation-123"
	ctx := context.WithValue(context.Background(), CorrelationIDKey, correlationID)

	logger.WithCorrelationID(ctx).Info("test message")

	output := buf.String()
	if !strings.Contains(output, correlationID) {
		t.Errorf("expected log output to contain correlation ID %s, got %s", correlationID, output)
	}
}

func TestLogger_WithField(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(Config{
		Level:       "info",
		Output:      buf,
		PrettyPrint: false,
	})

	logger.WithField("user_id", "12345").Info("test message")

	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("failed to parse log output as JSON: %v", err)
	}

	if logEntry["user_id"] != "12345" {
		t.Errorf("expected user_id to be 12345, got %v", logEntry["user_id"])
	}
}

func TestLogger_WithFields(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(Config{
		Level:       "info",
		Output:      buf,
		PrettyPrint: false,
	})

	fields := map[string]interface{}{
		"user_id":  "12345",
		"order_id": "order-789",
	}

	logger.WithFields(fields).Info("test message")

	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("failed to parse log output as JSON: %v", err)
	}

	if logEntry["user_id"] != "12345" {
		t.Errorf("expected user_id to be 12345, got %v", logEntry["user_id"])
	}
	if logEntry["order_id"] != "order-789" {
		t.Errorf("expected order_id to be order-789, got %v", logEntry["order_id"])
	}
}

func TestLogger_LogLevels(t *testing.T) {
	tests := []struct {
		name     string
		logLevel string
		logFunc  func(*Logger)
		contains string
	}{
		{
			name:     "debug log",
			logLevel: "debug",
			logFunc:  func(l *Logger) { l.Debug("debug message") },
			contains: "debug message",
		},
		{
			name:     "info log",
			logLevel: "info",
			logFunc:  func(l *Logger) { l.Info("info message") },
			contains: "info message",
		},
		{
			name:     "warn log",
			logLevel: "warn",
			logFunc:  func(l *Logger) { l.Warn("warn message") },
			contains: "warn message",
		},
		{
			name:     "error log",
			logLevel: "error",
			logFunc:  func(l *Logger) { l.Error("error message") },
			contains: "error message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			logger := New(Config{
				Level:       tt.logLevel,
				Output:      buf,
				PrettyPrint: false,
			})

			tt.logFunc(logger)

			output := buf.String()
			if !strings.Contains(output, tt.contains) {
				t.Errorf("expected log output to contain %s, got %s", tt.contains, output)
			}
		})
	}
}
