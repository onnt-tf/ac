package logger

import (
	"fmt"
	"os"
	"sync"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Level string

const (
	LevelDebug Level = "debug"
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"

	defaultLogDirectory = "./log"
)

var (
	sugaredLogger *zap.SugaredLogger
	once          sync.Once
)

// NewLogger initializes the logger with the provided configuration
func InitLogger() error {
	var initErr error
	once.Do(func() {
		if err := os.MkdirAll(defaultLogDirectory, 0744); err != nil {
			initErr = fmt.Errorf("failed to create log directory, err: %w", err)
			return
		}

		encoderConfig := zap.NewProductionEncoderConfig()
		encoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder

		fileEncoder := zapcore.NewJSONEncoder(encoderConfig)
		consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)

		outputLevel := zapcore.InfoLevel

		lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= outputLevel && lvl <= zapcore.InfoLevel
		})

		highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= outputLevel && lvl > zapcore.InfoLevel
		})

		cores := []zapcore.Core{
			zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), lowPriority),
			zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), highPriority),
			zapcore.NewCore(fileEncoder, zapcore.AddSync(&lumberjack.Logger{
				Filename:   fmt.Sprintf("%s/low.log", defaultLogDirectory),
				MaxSize:    500,
				MaxBackups: 3,
				MaxAge:     28,
			}), lowPriority),
			zapcore.NewCore(fileEncoder, zapcore.AddSync(&lumberjack.Logger{
				Filename:   fmt.Sprintf("%s/high.log", defaultLogDirectory),
				MaxSize:    500,
				MaxBackups: 3,
				MaxAge:     7,
			}), highPriority),
		}

		core := zapcore.NewTee(cores...)
		logger := zap.New(core)
		defer logger.Sync()

		sugaredLogger = logger.Sugar()
	})

	return initErr
}

func Get() *zap.SugaredLogger {
	return sugaredLogger
}

// LogWith records a log message with request context information
// Parameters:
//   - ctx: Echo context containing the HTTP request/response data
//   - level: Logging level (debug/info/warn/error)
//   - msg: Message to be logged
//   - kv: Additional key-value pairs to include in log
func LogWith(ctx echo.Context, level Level, msg string, kv map[string]interface{}) {
	// Create context fields map
	baseKv := map[string]interface{}{
		"method":     ctx.Request().Method,
		"uri":        ctx.Request().RequestURI,
		"request_id": ctx.Response().Header().Get(echo.HeaderXRequestID),
	}

	// Merge with external kv map, letting external values take precedence
	for k, v := range kv {
		baseKv[k] = v
	}

	// Add all fields to logger
	fields := make([]interface{}, 0, len(baseKv)*2)
	for k, v := range baseKv {
		fields = append(fields, k, v)
	}

	// Log message at appropriate level
	switch level {
	case LevelDebug:
		sugaredLogger.With(fields...).Debug(msg)
	case LevelInfo:
		sugaredLogger.With(fields...).Info(msg)
	case LevelWarn:
		sugaredLogger.With(fields...).Warn(msg)
	case LevelError:
		sugaredLogger.With(fields...).Error(msg)
	default:
		sugaredLogger.With(fields...).Info(msg)
	}
}

// The following functions wrap logWithContext providing a convenient way
// to log messages with different log levels: Info, Debug, Warn, and Error.
func Infof(ctx echo.Context, msg string, v ...interface{}) {
	LogWith(ctx, LevelInfo, fmt.Sprintf(msg, v...), nil)
}

func Debugf(ctx echo.Context, msg string, v ...interface{}) {
	LogWith(ctx, LevelDebug, fmt.Sprintf(msg, v...), nil)
}

func Warnf(ctx echo.Context, msg string, v ...interface{}) {
	LogWith(ctx, LevelWarn, fmt.Sprintf(msg, v...), nil)
}

func Errorf(ctx echo.Context, msg string, v ...interface{}) {
	LogWith(ctx, LevelError, fmt.Sprintf(msg, v...), nil)
}
