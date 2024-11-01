package logger

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// DefaultLogger is default logger.
var DefaultLogger = NewStdLogger(log.Writer())

// Logger is a logger interface.
type Logger interface {
	Log(ctx context.Context, level Level, keyvals ...interface{}) error
}

type LogEncoding string

const (
	LogEncodingConsole = LogEncoding("console")
	LogEncodingJSON    = LogEncoding("json")
)

type LogConfigOption func(*LogConfig)

type LogConfig struct {
	Level    string      `json:"level" yaml:"level" env:"DEFAULT_LOG_LEVEL"`
	Encoding LogEncoding `json:"encoding" yaml:"encoding" env:"DEFAULT_LOG_ENCODING"`

	// 日志存储配置
	Filepath string `json:"filepath" yaml:"filepath" env:"DEFAULT_LOG_FILEPATH"`
	// MaxSize is the maximum size in megabytes of the log file before it gets
	// rotated. It defaults to 100 megabytes.
	MaxSize int `json:"maxSize" yaml:"maxSize" env:"DEFAULT_LOG_MAX_SIZE"`
	// MaxAge is the maximum number of days to retain old log files based on the
	// timestamp encoded in their filename.  Note that a day is defined as 24
	// hours and may not exactly correspond to calendar days due to daylight
	// savings, leap seconds, etc. The default is not to remove old log files
	// based on age.
	MaxAge int `json:"maxAge" yaml:"maxAge" env:"DEFAULT_LOG_MAX_AGE"`
	// MaxBackups is the maximum number of old log files to retain.  The default
	// is to retain all old log files (though MaxAge may still cause them to get
	// deleted.)
	MaxBackups int `json:"maxBackups" yaml:"maxBackups" env:"DEFAULT_LOG_MAX_BACKUPS"`

	TraceProvider trace.TracerProvider `json:"-"`
}

func defaultLogConfig() *LogConfig {
	emptyOr := func(str string, defaultVal string) string {
		str = strings.TrimSpace(str)
		if str == "" {
			return defaultVal
		}
		return str
	}
	level := emptyOr(os.Getenv("DEFAULT_LOG_LEVEL"), "INFO")
	logEncoding := emptyOr(os.Getenv("DEFAULT_LOG_ENCODING"), string(LogEncodingConsole))
	logFilepath := emptyOr(os.Getenv("DEFAULT_LOG_FILEPATH"), "")
	maxSize := emptyOr(os.Getenv("DEFAULT_LOG_MAX_SIZE"), "0")
	maxAge := emptyOr(os.Getenv("DEFAULT_LOG_MAX_AGE"), "0")
	maxBackups := emptyOr(os.Getenv("DEFAULT_LOG_MAX_BACKUPS"), "0")

	i64MaxSize, _ := strconv.Atoi(maxSize)
	i64MaxAge, _ := strconv.Atoi(maxAge)
	i64MaxBackups, _ := strconv.Atoi(maxBackups)

	return &LogConfig{
		Level:         level,
		Encoding:      LogEncoding(logEncoding),
		Filepath:      logFilepath,
		MaxSize:       i64MaxSize,
		MaxAge:        i64MaxAge,
		MaxBackups:    i64MaxBackups,
		TraceProvider: otel.GetTracerProvider(),
	}
}

type LogWrapper func(ctx context.Context, level Level, keyvals ...interface{}) error

func (f LogWrapper) Log(ctx context.Context, level Level, keyvals ...interface{}) error {
	return f(ctx, level, keyvals...)
}

func NewLogger(opts ...LogConfigOption) (Logger, error) {
	c := defaultLogConfig()
	for _, opt := range opts {
		opt(c)
	}

	logfile := c.Filepath

	// 命令行输出一份
	syncers := []zapcore.WriteSyncer{
		zapcore.AddSync(os.Stdout),
	}
	if logfile != "" {
		// 配置lumberjack日志轮转
		w := zapcore.AddSync(&lumberjack.Logger{
			Filename:   logfile,
			MaxSize:    c.MaxSize,
			MaxBackups: c.MaxBackups,
			MaxAge:     c.MaxAge,
			Compress:   true,
		})

		// 创建自定义的WriteSyncer，结合控制台输出和文件输出
		syncers = append(syncers, w)
	}

	// 编码配置
	encoderConfig := zap.NewProductionEncoderConfig()
	var encoder zapcore.Encoder
	switch c.Encoding {
	case LogEncodingConsole:
		encoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	case LogEncodingJSON:
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	default:
		encoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// 使用自定义的WriteSyncer构建core
	core := zapcore.NewCore(
		encoder,
		zapcore.NewMultiWriteSyncer(syncers...),
		convZapLevel(ParseLevel(c.Level)),
	)

	// 使用core创建logger
	logger := zap.New(core)

	return LogWrapper(func(ctx context.Context, level Level, keyvals ...interface{}) error {
		span := trace.SpanContextFromContext(ctx)
		zapLevel := convZapLevel(level)
		fields := []zap.Field{
			zap.String("trace_id", span.TraceID().String()),
			zap.String("span_id", span.SpanID().String()),
		}

		for i := 0; i < len(keyvals); i += 2 {
			fields = append(fields, zap.String(fmt.Sprintf("%v", keyvals[i]), fmt.Sprintf("%v", keyvals[i+1])))
		}
		logger.Log(zapLevel, "", fields...)
		return nil
	}), nil
}

func convZapLevel(level Level) zapcore.Level {
	zapLevel := zap.DebugLevel
	switch level {
	case LevelDebug:
		zapLevel = zap.DebugLevel
	case LevelInfo:
		zapLevel = zap.InfoLevel
	case LevelWarn:
		zapLevel = zap.WarnLevel
	case LevelError:
		zapLevel = zap.ErrorLevel
	case LevelFatal:
		zapLevel = zap.FatalLevel
	}
	return zapLevel
}
