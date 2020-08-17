package log

import (
	"io"
	"os"
	"github.com/sirupsen/logrus"
)

// Pattern sourced from here: https://stackoverflow.com/questions/30257622/golang-logrus-how-to-do-a-centralized-configuration
var (
	logger *logrus.Logger
	envLogLevel = "LOG_LEVEL"
	defaultLogLevel = logrus.WarnLevel
)

type Level logrus.Level

// pulled from logrus
// These are the different logging levels. You can set the logging level to log
// on your instance of logger, obtained with `logrus.New()`.
const (
	// PanicLevel level, highest level of severity. Logs and then calls panic with the
	// message passed to Debug, Info, ...
	PanicLevel Level = iota
	// FatalLevel level. Logs and then calls `logger.Exit(1)`. It will exit even if the
	// logging level is set to Panic.
	FatalLevel
	// ErrorLevel level. Logs. Used for errors that should definitely be noted.
	// Commonly used for hooks to send errors to an error tracking service.
	ErrorLevel
	// WarnLevel level. Non-critical entries that deserve eyes.
	WarnLevel
	// InfoLevel level. General operational entries about what's going on inside the
	// application.
	InfoLevel
	// DebugLevel level. Usually only enabled when debugging. Very verbose logging.
	DebugLevel
	// TraceLevel level. Designates finer-grained informational events than the Debug.
	TraceLevel
)

func init() {
	var (
		envLevel string
		level logrus.Level
		ok bool
		err error
	)

	logger = logrus.New()

	if envLevel, ok = os.LookupEnv(envLogLevel); !ok {
		level = defaultLogLevel
	} else if level, err = logrus.ParseLevel(envLevel); err != nil {
		logrus.Errorf("error parsing %s: %v", envLogLevel, err)
		level = defaultLogLevel
	}

	logger.SetLevel(level)
	// logger.SetReportCaller(true)
	logger.SetFormatter(&logrus.TextFormatter{
		DisableColors: false,
		FullTimestamp: false,
	})

	logger.Info("set log level to ", level)
}

func SetLevel(level Level) {
	logger.SetLevel(logrus.Level(level))
}

func SetOutput(writer io.Writer) {
	logger.SetOutput(writer)
}

func Trace(args ...interface{}) {
	logger.Trace(args...)
}

func Debug(args ...interface{}) {
	logger.Debug(args...)
}

func Info(args ...interface{}) {
	logger.Info(args...)
}

func Warn(args ...interface{}) {
	logger.Warn(args...)
}

func Fatal(args ...interface{}) {
	logger.Fatal(args...)
}

func Panic(args ...interface{}) {
	logger.Panic(args...)
}

func WithField(name, value string) *logrus.Entry {
	return logger.WithField(name, value)
}

func WithFields(fields logrus.Fields) *logrus.Entry {
	return logger.WithFields(fields)
}
