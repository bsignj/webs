package log

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var ZLogger *zap.Logger

// NewLogger creates and configures a new zap.Logger based on the provided parameters.
// logLevel: specifies the verbosity of the log outputs (e.g., DEBUG, INFO, WARN, ERROR).
// outputPath: determines where the log outputs are written (file path or "stdout").
// errOutputPath: determines where error logs are written (typically "stderr").
func NewLogger(logLevel, outputPath, errOutputPath string) *zap.Logger {
	var zapLogLevel zap.AtomicLevel

	// Set the corresponding zap log level based on the input string
	switch logLevel {
	case "DEBUG":
		zapLogLevel = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "INFO":
		zapLogLevel = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "WARN":
		zapLogLevel = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "ERROR":
		zapLogLevel = zap.NewAtomicLevelAt(zap.ErrorLevel)
	case "PANIC":
		zapLogLevel = zap.NewAtomicLevelAt(zap.PanicLevel)
	case "FATAL":
		zapLogLevel = zap.NewAtomicLevelAt(zap.FatalLevel)
	default:
		zapLogLevel = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	// Set default output paths if none are provided
	if outputPath == "" {
		outputPath = "stdout"
	}
	if errOutputPath == "" {
		errOutputPath = "stderr"
	}

	// Use date-formatted log files if output paths are set to "file"
	if outputPath == "file" {
		outputPath = time.Now().Format("2006-01-02") + ".log"
	}
	if errOutputPath == "file" {
		errOutputPath = time.Now().Format("2006-01-02") + "_error.log"
	}

	// Create zap configuration
	var cfg = zap.Config{
		Level:       zapLogLevel,
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 1000,
		},
		Encoding:         "console",
		EncoderConfig:    getCustomEncoderConfig(),
		OutputPaths:      []string{outputPath},
		ErrorOutputPaths: []string{errOutputPath},
	}

	// Build the logger based on the configuration
	zapLogger, err := cfg.Build()
	if err != nil {
		os.Stderr.WriteString("Error creating logger: " + err.Error() + "\n")
		os.Stderr.WriteString("Fehler beim Erstellen des Loggers: " + err.Error() + "\n")
		return nil
	}

	ZLogger = zapLogger
	return ZLogger
}

// getCustomEncoderConfig creates a custom encoder configuration for the logger.
func getCustomEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.RFC3339TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	}
}
