package logger

import (
	"context"
	"fmt"
	"os"
	"sync"
)

// globalLogger is designed as a global logger in current process.
var global = &loggerAppliance{}

// loggerAppliance is the proxy of `Logger` to
// make logger change will affect all sub-logger.
type loggerAppliance struct {
	lock sync.Mutex
	Logger
}

func init() {
	global.SetLogger(DefaultLogger)
}

func (a *loggerAppliance) SetLogger(in Logger) {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.Logger = in
}

// SetLogger should be called before any other log call.
// And it is NOT THREAD SAFE.
func SetLogger(logger Logger) {
	global.SetLogger(logger)
}

// GetLogger returns global logger appliance as logger in current process.
func GetLogger() Logger {
	return global
}

// Log Print log by level and keyvals.
func Log(ctx context.Context, level Level, keyvals ...interface{}) {
	_ = global.Log(ctx, level, keyvals...)
}

// Debug logs a message at debug level.
func Debug(ctx context.Context, a ...interface{}) {
	_ = global.Log(ctx, LevelDebug, DefaultMessageKey, fmt.Sprint(a...))
}

// Debugf logs a message at debug level.
func Debugf(ctx context.Context, format string, a ...interface{}) {
	_ = global.Log(ctx, LevelDebug, DefaultMessageKey, fmt.Sprintf(format, a...))
}

// Debugw logs a message at debug level.
func Debugw(ctx context.Context, keyvals ...interface{}) {
	_ = global.Log(ctx, LevelDebug, keyvals...)
}

// Info logs a message at info level.
func Info(ctx context.Context, a ...interface{}) {
	_ = global.Log(ctx, LevelInfo, DefaultMessageKey, fmt.Sprint(a...))
}

// Infof logs a message at info level.
func Infof(ctx context.Context, format string, a ...interface{}) {
	_ = global.Log(ctx, LevelInfo, DefaultMessageKey, fmt.Sprintf(format, a...))
}

// Infow logs a message at info level.
func Infow(ctx context.Context, keyvals ...interface{}) {
	_ = global.Log(ctx, LevelInfo, keyvals...)
}

// Warn logs a message at warn level.
func Warn(ctx context.Context, a ...interface{}) {
	_ = global.Log(ctx, LevelWarn, DefaultMessageKey, fmt.Sprint(a...))
}

// Warnf logs a message at warnf level.
func Warnf(ctx context.Context, format string, a ...interface{}) {
	_ = global.Log(ctx, LevelWarn, DefaultMessageKey, fmt.Sprintf(format, a...))
}

// Warnw logs a message at warnf level.
func Warnw(ctx context.Context, keyvals ...interface{}) {
	_ = global.Log(ctx, LevelWarn, keyvals...)
}

// Error logs a message at error level.
func Error(ctx context.Context, a ...interface{}) {
	_ = global.Log(ctx, LevelError, DefaultMessageKey, fmt.Sprint(a...))
}

// Errorf logs a message at error level.
func Errorf(ctx context.Context, format string, a ...interface{}) {
	_ = global.Log(ctx, LevelError, DefaultMessageKey, fmt.Sprintf(format, a...))
}

// Errorw logs a message at error level.
func Errorw(ctx context.Context, keyvals ...interface{}) {
	_ = global.Log(ctx, LevelError, keyvals...)
}

// Fatal logs a message at fatal level.
func Fatal(ctx context.Context, a ...interface{}) {
	_ = global.Log(ctx, LevelFatal, DefaultMessageKey, fmt.Sprint(a...))
	os.Exit(1)
}

// Fatalf logs a message at fatal level.
func Fatalf(ctx context.Context, format string, a ...interface{}) {
	_ = global.Log(ctx, LevelFatal, DefaultMessageKey, fmt.Sprintf(format, a...))
	os.Exit(1)
}

// Fatalw logs a message at fatal level.
func Fatalw(ctx context.Context, keyvals ...interface{}) {
	_ = global.Log(ctx, LevelFatal, keyvals...)
	os.Exit(1)
}
