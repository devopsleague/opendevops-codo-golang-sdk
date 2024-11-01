package logger

import (
	"context"
	"fmt"
	"os"
)

// DefaultMessageKey default message key.
var DefaultMessageKey = "msg"

// Option is Helper option.
type Option func(*Helper)

// Helper is a logger helper.
type Helper struct {
	logger  Logger
	msgKey  string
	sprint  func(...interface{}) string
	sprintf func(format string, a ...interface{}) string
}

// WithMessageKey with message key.
func WithMessageKey(k string) Option {
	return func(opts *Helper) {
		opts.msgKey = k
	}
}

// WithSprint with sprint
func WithSprint(sprint func(...interface{}) string) Option {
	return func(opts *Helper) {
		opts.sprint = sprint
	}
}

// WithSprintf with sprintf
func WithSprintf(sprintf func(format string, a ...interface{}) string) Option {
	return func(opts *Helper) {
		opts.sprintf = sprintf
	}
}

// NewHelper new a logger helper.
func NewHelper(logger Logger, opts ...Option) *Helper {
	options := &Helper{
		msgKey:  DefaultMessageKey, // default message key
		logger:  logger,
		sprint:  fmt.Sprint,
		sprintf: fmt.Sprintf,
	}
	for _, o := range opts {
		o(options)
	}
	return options
}

// Debug logs a message at debug level.
func (h *Helper) Debug(ctx context.Context, a ...interface{}) {
	_ = h.logger.Log(ctx, LevelDebug, h.msgKey, h.sprint(a...))
}

// Debugf logs a message at debug level.
func (h *Helper) Debugf(ctx context.Context, format string, a ...interface{}) {
	_ = h.logger.Log(ctx, LevelDebug, h.msgKey, h.sprintf(format, a...))
}

// Debugw logs a message at debug level.
func (h *Helper) Debugw(ctx context.Context, keyvals ...interface{}) {
	_ = h.logger.Log(ctx, LevelDebug, keyvals...)
}

// Info logs a message at info level.
func (h *Helper) Info(ctx context.Context, a ...interface{}) {
	_ = h.logger.Log(ctx, LevelInfo, h.msgKey, h.sprint(a...))
}

// Infof logs a message at info level.
func (h *Helper) Infof(ctx context.Context, format string, a ...interface{}) {
	_ = h.logger.Log(ctx, LevelInfo, h.msgKey, h.sprintf(format, a...))
}

// Infow logs a message at info level.
func (h *Helper) Infow(ctx context.Context, keyvals ...interface{}) {
	_ = h.logger.Log(ctx, LevelInfo, keyvals...)
}

// Warn logs a message at warn level.
func (h *Helper) Warn(ctx context.Context, a ...interface{}) {
	_ = h.logger.Log(ctx, LevelWarn, h.msgKey, h.sprint(a...))
}

// Warnf logs a message at warnf level.
func (h *Helper) Warnf(ctx context.Context, format string, a ...interface{}) {
	_ = h.logger.Log(ctx, LevelWarn, h.msgKey, h.sprintf(format, a...))
}

// Warnw logs a message at warnf level.
func (h *Helper) Warnw(ctx context.Context, keyvals ...interface{}) {
	_ = h.logger.Log(ctx, LevelWarn, keyvals...)
}

// Error logs a message at error level.
func (h *Helper) Error(ctx context.Context, a ...interface{}) {
	_ = h.logger.Log(ctx, LevelError, h.msgKey, h.sprint(a...))
}

// Errorf logs a message at error level.
func (h *Helper) Errorf(ctx context.Context, format string, a ...interface{}) {
	_ = h.logger.Log(ctx, LevelError, h.msgKey, h.sprintf(format, a...))
}

// Errorw logs a message at error level.
func (h *Helper) Errorw(ctx context.Context, keyvals ...interface{}) {
	_ = h.logger.Log(ctx, LevelError, keyvals...)
}

// Fatal logs a message at fatal level.
func (h *Helper) Fatal(ctx context.Context, a ...interface{}) {
	_ = h.logger.Log(ctx, LevelFatal, h.msgKey, h.sprint(a...))
	os.Exit(1)
}

// Fatalf logs a message at fatal level.
func (h *Helper) Fatalf(ctx context.Context, format string, a ...interface{}) {
	_ = h.logger.Log(ctx, LevelFatal, h.msgKey, h.sprintf(format, a...))
	os.Exit(1)
}

// Fatalw logs a message at fatal level.
func (h *Helper) Fatalw(ctx context.Context, keyvals ...interface{}) {
	_ = h.logger.Log(ctx, LevelFatal, keyvals...)
	os.Exit(1)
}
