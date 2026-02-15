package logger

import (
	"context"
	"log/slog"
	"os"
)

var defaultLogger *slog.Logger

func init() {
	defaultLogger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(defaultLogger)
}

// Init initializes the logger with the specified options
func Init(opts ...Option) {
	cfg := &config{
		level:  slog.LevelInfo,
		format: "text",
		output: os.Stderr,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	var handler slog.Handler
	handlerOpts := &slog.HandlerOptions{
		Level: cfg.level,
	}

	switch cfg.format {
	case "json":
		handler = slog.NewJSONHandler(cfg.output, handlerOpts)
	default:
		handler = slog.NewTextHandler(cfg.output, handlerOpts)
	}

	defaultLogger = slog.New(handler)
	slog.SetDefault(defaultLogger)
}

// config holds logger configuration
type config struct {
	level  slog.Level
	format string
	output *os.File
}

// Option configures the logger
type Option func(*config)

// WithLevel sets the log level
func WithLevel(level slog.Level) Option {
	return func(c *config) {
		c.level = level
	}
}

// WithFormat sets the output format ("text" or "json")
func WithFormat(format string) Option {
	return func(c *config) {
		c.format = format
	}
}

// WithOutput sets the output destination
func WithOutput(output *os.File) Option {
	return func(c *config) {
		c.output = output
	}
}

// Debug logs at debug level
func Debug(msg string, args ...any) {
	defaultLogger.Debug(msg, args...)
}

// Info logs at info level
func Info(msg string, args ...any) {
	defaultLogger.Info(msg, args...)
}

// Warn logs at warn level
func Warn(msg string, args ...any) {
	defaultLogger.Warn(msg, args...)
}

// Error logs at error level
func Error(msg string, args ...any) {
	defaultLogger.Error(msg, args...)
}

// Fatal logs at error level and exits
func Fatal(msg string, args ...any) {
	defaultLogger.Error(msg, args...)
	os.Exit(1)
}

// With returns a logger with additional attributes
func With(args ...any) *slog.Logger {
	return defaultLogger.With(args...)
}

// WithContext returns a logger from context or the default logger
func WithContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(loggerKey{}).(*slog.Logger); ok {
		return l
	}
	return defaultLogger
}

// NewContext returns a context with the logger attached
func NewContext(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, l)
}

type loggerKey struct{}

// Default returns the default logger
func Default() *slog.Logger {
	return defaultLogger
}
