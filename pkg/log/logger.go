package log

import (
	"io"
	"log"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Pattern sourced from here: https://stackoverflow.com/questions/30257622/golang-zap-how-to-do-a-centralized-configuration
var (
	output io.Writer = os.Stdout
	logger *zap.Logger
	envLogLevel = "LOG_LEVEL"
	defaultLogLevel = zap.WarnLevel
	atomicLevel zap.AtomicLevel
)

// Level of logging
type Level = zapcore.Level

const (
	// DebugLevel logs are typically voluminous, and are usually disabled in
	// production.
	DebugLevel Level = zapcore.DebugLevel
	// InfoLevel is the default logging priority.
	InfoLevel Level = zapcore.InfoLevel
	// WarnLevel logs are more important than Info, but don't need individual
	// human review.
	WarnLevel Level = zapcore.WarnLevel
	// ErrorLevel logs are high-priority. If an application is running smoothly,
	// it shouldn't generate any error-level logs.
	ErrorLevel Level = zapcore.ErrorLevel
	// DPanicLevel logs are particularly important errors. In development the
	// logger panics after writing the message.
	DPanicLevel Level = zapcore.DPanicLevel
	// PanicLevel logs a message, then panics.
	PanicLevel Level = zapcore.PanicLevel
	// FatalLevel logs a message, then calls os.Exit(1).
	FatalLevel Level = zapcore.FatalLevel

	_minLevel = DebugLevel
	_maxLevel = FatalLevel
)

// Field represents the key-value pair for logging
type Field = zap.Field
type WriteSyncer = zapcore.WriteSyncer

func init() {
	var (
		envLevel string
		level zapcore.Level
		ok bool
		err error
	)

	if envLevel, ok = os.LookupEnv(envLogLevel); !ok {
		level = defaultLogLevel
	} else if err = level.UnmarshalText([]byte(envLevel)); err != nil {
		log.Fatalf("parsing error %s %w", envLogLevel, err)
		level = defaultLogLevel
	}

	atomicLevel = zap.NewAtomicLevelAt(level)

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = ""

	logger = zap.New(
		zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderCfg),
			zapcore.Lock(zapcore.AddSync(output)),
			atomicLevel,
		),
		zap.AddCaller(),
		zap.AddCallerSkip(3),
	)
}

func SetWriteSyncer(out io.Writer) {
	output = out
}

func SetLevel(level Level) {
	atomicLevel.SetLevel(level)
}

func Sync() {
	logger.Sync()
}

func Debug(msg string, fields ...Field) {
	logger.Debug(msg, fields...)
}

func Info(msg string, fields ...Field) {
	logger.Info(msg, fields...)
}

func Warn(msg string, fields ...Field) {
	logger.Warn(msg, fields...)
}

func Fatal(msg string, fields ...Field) {
	logger.Fatal(msg, fields...)
}

func Panic(msg string, fields ...Field) {
	logger.Panic(msg, fields...)
}

// writeSyncer decorates an io.Writer with a no-op Sync() function.
type writeSyncer struct {
	io.Writer
}

// Sync does nothing since all output was written to the writer immediately.
func (ws writeSyncer) Sync() error {
	return nil
}
