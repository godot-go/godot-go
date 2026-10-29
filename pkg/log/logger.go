package log

import (
	"fmt"
	"io"
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Pattern sourced from here: https://stackoverflow.com/questions/30257622/golang-zap-how-to-do-a-centralized-configuration
// swapWriter is an io.Writer that forwards to a replaceable target. zap
// captures the writer inside the core when the logger is built (in init), so
// reassigning the package `output` variable would leave the running logger
// writing to the original stdout. Routing the core through this indirection
// lets SetWriteSyncer point entries at a capture buffer (and restore stdout
// afterwards) at runtime, which the diagnostic-capture tests depend on.
//
// Zap already serialises writes with the outer zapcore.Lock, so the only state
// here that needs guarding is the target pointer itself: Swap rewrites it while
// concurrent writers read it.
type swapWriter struct {
	mu     sync.RWMutex
	target io.Writer
}

func (w *swapWriter) Write(p []byte) (int, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.target.Write(p)
}

// Swap replaces the write target and returns the previous one so a caller can
// restore it.
func (w *swapWriter) Swap(next io.Writer) io.Writer {
	w.mu.Lock()
	defer w.mu.Unlock()
	prev := w.target
	w.target = next
	return prev
}

// Sync forwards to the target when it supports syncing, preserving the
// behaviour AddSync gave the original *os.File target.
func (w *swapWriter) Sync() error {
	w.mu.RLock()
	defer w.mu.RUnlock()
	if s, ok := w.target.(interface{ Sync() error }); ok {
		return s.Sync()
	}
	return nil
}

var (
	// stdoutSink is the writer installed in the logger core. SetWriteSyncer
	// swaps its target rather than rebinding the variable.
	stdoutSink                = &swapWriter{target: os.Stdout}
	output          io.Writer = stdoutSink
	logger          *zap.Logger
	envLogLevel     = "LOG_LEVEL"
	defaultLogLevel = zap.WarnLevel
	atomicLevel     zap.AtomicLevel
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

// WriteSyncer represents a writer than support syncing.
type WriteSyncer = zapcore.WriteSyncer

func init() {
	var (
		envLevel string
		level    zapcore.Level
		ok       bool
		err      error
	)

	if envLevel, ok = os.LookupEnv(envLogLevel); !ok {
		level = defaultLogLevel
	} else if err = level.UnmarshalText([]byte(envLevel)); err != nil {
		_ = fmt.Errorf("parsing error %s %w", envLogLevel, err)
		level = defaultLogLevel
	}

	atomicLevel = zap.NewAtomicLevelAt(level)

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	encoderCfg.TimeKey = ""

	options := []zap.Option{
		zap.AddCaller(),
		zap.AddCallerSkip(1),
	}

	// show stacktrace
	options = append(options, zap.AddStacktrace(zap.WarnLevel))

	logger = zap.New(
		zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderCfg),
			zapcore.Lock(zapcore.AddSync(output)),
			atomicLevel,
		),
		options...,
	)
}

// SetWriteSyncer redirects subsequent entries to out and returns the previous
// writer so the caller can restore it. The swap takes effect for the already
// built logger: the core writes through stdoutSink, whose target changes here.
func SetWriteSyncer(out io.Writer) io.Writer {
	return stdoutSink.Swap(out)
}

// SetLevel calls SetLevel on the underlying logger
func SetLevel(level Level) {
	atomicLevel.SetLevel(level)
}

func GetLevel() Level {
	return atomicLevel.Level()
}

// Sync calls Sync on the underlying logger
func Sync() {
	logger.Sync()
}

// Debug logs error messages
func Debug(msg string, fields ...Field) {
	logger.Debug(msg, fields...)
}

// Info logs error messages
func Info(msg string, fields ...Field) {
	logger.Info(msg, fields...)
}

// Warn logs error messages
func Warn(msg string, fields ...Field) {
	logger.Warn(msg, fields...)
}

// Error logs error messages
func Error(msg string, fields ...Field) {
	logger.Error(msg, fields...)
}

// Fatal logs error messages
func Fatal(msg string, fields ...Field) {
	logger.Fatal(msg, fields...)
}

// Panic logs error messages
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
