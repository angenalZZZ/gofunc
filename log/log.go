package log

import (
	"fmt"
	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"strings"
	"time"
)

type (
	Logger = *zerolog.Logger
	Level  = zerolog.Level
)

// Log default logger.
var Log Logger

const (
	// DebugLevel defines debug log level.
	DebugLevel = zerolog.DebugLevel
	// InfoLevel defines info log level.
	InfoLevel = zerolog.InfoLevel
	// WarnLevel defines warn log level.
	WarnLevel = zerolog.WarnLevel
	// ErrorLevel defines error log level.
	ErrorLevel = zerolog.ErrorLevel
	// FatalLevel defines fatal log level.
	FatalLevel = zerolog.FatalLevel
	// PanicLevel defines panic log level.
	PanicLevel = zerolog.PanicLevel
	// NoLevel defines an absent log level.
	NoLevel = zerolog.NoLevel
	// Disabled disables the logger.
	Disabled = zerolog.Disabled

	// TraceLevel defines trace log level.
	TraceLevel Level = -1
)

// Init zero logger.
func Init(c Config) Logger {
	zerolog.TimeFieldFormat = "2006-01-02 15:04:05.000"
	zerolog.DurationFieldInteger = true
	zerolog.TimestampFieldName = "timestamp"
	zerolog.DurationFieldUnit = time.Millisecond

	var writers []io.Writer
	if strings.Contains(c.Writers, "file") {
		//if _, err := os.Stat(c.Filename); err != nil && os.IsNotExist(err) {
		//	if err = os.MkdirAll(c.Filename, 0644); err != nil {
		//		panic(err)
		//	}
		//}
		w := &lumberjack.Logger{
			Filename:   c.Filename,
			MaxSize:    c.MaxSize,
			MaxAge:     c.MaxAge,
			MaxBackups: c.MaxBackups,
			LocalTime:  c.LocalTime,
			Compress:   c.Compress,
		}
		writers = append(writers, w)
	}
	if strings.Contains(c.Writers, "stdout") {
		w := newConsole()
		writers = append(writers, w)
	}

	w := io.MultiWriter(writers...)
	z := zerolog.New(w).With().Timestamp()
	l := z.Logger()
	return &l
}

// InitConsole zero console logger.
func InitConsole() Logger {
	w := newConsole()
	z := zerolog.New(w).With().Timestamp()
	l := z.Logger()
	return &l
}

func newConsole() io.Writer {
	w := zerolog.NewConsoleWriter()
	w.TimeFormat = "01-02 15:04:05"
	w.NoColor = true
	// FormatLevel https://github.com/rs/zerolog/blob/master/console.go#L315
	w.FormatLevel = func(i interface{}) string {
		if s, ok := i.(string); ok {
			return s
		}
		if s, ok := i.(fmt.Stringer); ok {
			return s.String()
		}
		return fmt.Sprint(i)
	}
	return w
}
