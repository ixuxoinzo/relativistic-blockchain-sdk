package utils

import (
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LoggerConfig struct {
	Level    string
	Format   string
	Output   string
	FilePath string
}

type Logger struct {
	*zap.Logger
	config *LoggerConfig
}

func NewLogger(config *LoggerConfig) (*Logger, error) {
	var zapConfig zap.Config

	if config.Format == "json" {
		zapConfig = zap.NewProductionConfig()
	} else {
		zapConfig = zap.NewDevelopmentConfig()
	}

	level, err := zap.ParseAtomicLevel(config.Level)
	if err != nil {
		return nil, err
	}
	zapConfig.Level = level

	var outputPaths []string
	switch config.Output {
	case "file":
		if config.FilePath == "" {
			config.FilePath = "logs/app.log"
		}
		outputPaths = []string{config.FilePath}
	case "both":
		if config.FilePath == "" {
			config.FilePath = "logs/app.log"
		}
		outputPaths = []string{config.FilePath, "stdout"}
	default:
		outputPaths = []string{"stdout"}
	}
	zapConfig.OutputPaths = outputPaths
	zapConfig.ErrorOutputPaths = outputPaths

	zapConfig.EncoderConfig.TimeKey = "timestamp"
	zapConfig.EncoderConfig.LevelKey = "level"
	zapConfig.EncoderConfig.NameKey = "logger"
	zapConfig.EncoderConfig.CallerKey = "caller"
	zapConfig.EncoderConfig.MessageKey = "message"
	zapConfig.EncoderConfig.StacktraceKey = "stacktrace"
	zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	zapConfig.EncoderConfig.EncodeDuration = zapcore.StringDurationEncoder
	zapConfig.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	if config.Output == "file" || config.Output == "both" {
		dir := filepath.Dir(config.FilePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}
	}

	logger, err := zapConfig.Build()
	if err != nil {
		return nil, err
	}

	return &Logger{
		Logger: logger,
		config: config,
	}, nil
}

func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.Logger.Debug(msg, fields...)
}

func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.Logger.Info(msg, fields...)
}

func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.Logger.Warn(msg, fields...)
}

func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.Logger.Error(msg, fields...)
}

func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	l.Logger.Fatal(msg, fields...)
}

func (l *Logger) With(fields ...zap.Field) *Logger {
	return &Logger{
		Logger: l.Logger.With(fields...),
		config: l.config,
	}
}

func (l *Logger) Sync() error {
	return l.Logger.Sync()
}

func (l *Logger) LogHTTPRequest(method, path, query, ip, userAgent string, status int, duration time.Duration) {
	fields := []zap.Field{
		zap.String("method", method),
		zap.String("path", path),
		zap.String("query", query),
		zap.String("ip", ip),
		zap.String("user_agent", userAgent),
		zap.Int("status", status),
		zap.Duration("duration", duration),
	}

	if status >= 500 {
		l.Error("HTTP request", fields...)
	} else if status >= 400 {
		l.Warn("HTTP request", fields...)
	} else {
		l.Info("HTTP request", fields...)
	}
}

func (l *Logger) LogNodeEvent(eventType, nodeID string, fields ...zap.Field) {
	allFields := []zap.Field{
		zap.String("event_type", eventType),
		zap.String("node_id", nodeID),
	}
	allFields = append(allFields, fields...)
	l.Info("Node event", allFields...)
}

func (l *Logger) LogValidationResult(valid bool, confidence float64, reason string, fields ...zap.Field) {
	allFields := []zap.Field{
		zap.Bool("valid", valid),
		zap.Float64("confidence", confidence),
		zap.String("reason", reason),
	}
	allFields = append(allFields, fields...)

	if valid {
		l.Info("Validation result", allFields...)
	} else {
		l.Warn("Validation result", allFields...)
	}
}

func (l *Logger) LogMetric(name string, value float64, fields ...zap.Field) {
	allFields := []zap.Field{
		zap.String("metric", name),
		zap.Float64("value", value),
	}
	allFields = append(allFields, fields...)
	l.Info("Metric", allFields...)
}

func (l *Logger) LogError(err error, message string, fields ...zap.Field) {
	allFields := []zap.Field{
		zap.Error(err),
	}
	allFields = append(allFields, fields...)
	l.Error(message, allFields...)
}
