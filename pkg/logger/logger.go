// Package logger provides a thin wrapper around uber-go/zap with helpers
// for carrying contextual values (request_id, user_id, ...) through context.Context.
package logger

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap.Logger with extra helpers.
type Logger struct {
	z *zap.Logger
}

// New builds a Logger from a level ("debug"|"info"|"warn"|"error") and a
// format ("json"|"console"). JSON is intended for production; console for dev.
func New(level, format string) (*Logger, error) {
	var cfg zap.Config
	if format == "json" {
		cfg = zap.NewProductionConfig()
	} else {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	if err := cfg.Level.UnmarshalText([]byte(level)); err != nil {
		return nil, fmt.Errorf("invalid log level %q: %w", level, err)
	}
	cfg.DisableStacktrace = false
	z, err := cfg.Build(zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))
	if err != nil {
		return nil, fmt.Errorf("build zap logger: %w", err)
	}
	return &Logger{z: z}, nil
}

// With returns a child logger with extra fields attached.
func (l *Logger) With(fields ...zap.Field) *Logger {
	return &Logger{z: l.z.With(fields...)}
}

// Sync flushes any buffered log entries. Call before process exit.
func (l *Logger) Sync() error { return l.z.Sync() }

// Info logs an info-level message.
func (l *Logger) Info(msg string, fields ...zap.Field) { l.z.Info(msg, fields...) }

// Warn logs a warn-level message.
func (l *Logger) Warn(msg string, fields ...zap.Field) { l.z.Warn(msg, fields...) }

// Error logs an error-level message.
func (l *Logger) Error(msg string, fields ...zap.Field) { l.z.Error(msg, fields...) }

// Debug logs a debug-level message.
func (l *Logger) Debug(msg string, fields ...zap.Field) { l.z.Debug(msg, fields...) }

// Fatal logs at fatal level and exits the process with status 1.
func (l *Logger) Fatal(msg string, fields ...zap.Field) { l.z.Fatal(msg, fields...) }

// Zap returns the underlying *zap.Logger for advanced use.
func (l *Logger) Zap() *zap.Logger { return l.z }

// ----------------------------------------------------------------------
// Context helpers (for request_id, user_id propagation)
// ----------------------------------------------------------------------

type ctxKey int

const (
	ctxKeyRequestID ctxKey = iota
	ctxKeyUserID
)

// WithRequestID returns a copy of ctx that carries the request_id value.
func WithRequestID(ctx context.Context, rid string) context.Context {
	return context.WithValue(ctx, ctxKeyRequestID, rid)
}

// RequestIDFromContext returns the request_id stored in ctx, or "" if absent.
func RequestIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(ctxKeyRequestID).(string)
	return v
}

// WithUserID returns a copy of ctx that carries the authenticated user_id.
func WithUserID(ctx context.Context, uid int64) context.Context {
	return context.WithValue(ctx, ctxKeyUserID, uid)
}

// UserIDFromContext returns the user_id stored in ctx, or 0 if absent.
func UserIDFromContext(ctx context.Context) int64 {
	v, _ := ctx.Value(ctxKeyUserID).(int64)
	return v
}

// FieldsFromContext extracts known keys from ctx and returns them as zap fields.
func FieldsFromContext(ctx context.Context) []zap.Field {
	fields := make([]zap.Field, 0, 2)
	if rid := RequestIDFromContext(ctx); rid != "" {
		fields = append(fields, zap.String("request_id", rid))
	}
	if uid := UserIDFromContext(ctx); uid != 0 {
		fields = append(fields, zap.Int64("user_id", uid))
	}
	return fields
}
