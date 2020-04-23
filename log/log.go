package log

import (
	"fmt"
	"github.com/angenalZZZ/gofunc/f"
	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"strings"
	"time"
)

// Log default logger, or use log of import "github.com/rs/zerolog/log"
var Log *Logger

// Logger *zerolog.Logger
type Logger = zerolog.Logger

// Init the global zero logger.
func init() {
	zerolog.CallerFieldName = "c"     // default: caller
	zerolog.ErrorFieldName = "e"      // default: error
	zerolog.ErrorStackFieldName = "s" // default: stack
	zerolog.LevelFieldName = "l"      // default: level
	zerolog.MessageFieldName = "m"    // default: message
	zerolog.TimestampFieldName = "t"  // default: time

	zerolog.DurationFieldInteger = true
	zerolog.DurationFieldUnit = time.Millisecond
	//zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	//zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack // Enable stack trace
}

// Init the zero logger.
func Init(c *Config) *Logger {
	// sets the global override for log level and time format.
	//if level, err := zerolog.ParseLevel(c.Level); err == nil {
	//	zerolog.SetGlobalLevel(level)
	//}
	if c.TimeFormat != "" {
		zerolog.TimeFieldFormat = c.TimeFormat
	}
	// configs writers
	var writers []io.Writer
	if strings.Contains(c.Writers, "file") {
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
		w := newConsole(zerolog.TimeFieldFormat, false)
		writers = append(writers, w)
	}
	w := io.MultiWriter(writers...)
	z := zerolog.New(w).With().Timestamp()
	l := z.Logger()
	if level, err := zerolog.ParseLevel(c.Level); err == nil {
		l.Level(level)
	}
	return &l
}

// InitConsole zero console logger.
func InitConsole(timeFormat string, jsonWriter bool) *Logger {
	w := newConsole(timeFormat, jsonWriter)
	l := zerolog.New(w).With().Timestamp().Logger()
	return &l
}

func newConsole(timeFormat string, jsonWriter bool) io.Writer {
	w := zerolog.NewConsoleWriter()
	w.TimeFormat = timeFormat
	w.NoColor = true // NoColor to Improve efficiency
	if jsonWriter {
		return w
	}
	setConsoleWriterFormat(&w)
	return &w
}

func newFileWriter(writer io.Writer, timeFormat string) io.Writer {
	w := &zerolog.ConsoleWriter{Out: writer}
	setConsoleWriterFormat(w)
	return w
}

func setConsoleWriterFormat(w *zerolog.ConsoleWriter) {
	// https://github.com/rs/zerolog/blob/master/console.go#L315
	w.FormatLevel = func(i interface{}) string {
		return fmt.Sprintf(`[%s]`, i)
	}
	w.FormatFieldName = func(i interface{}) string {
		return fmt.Sprintf(`"%s":`, i)
	}
	w.FormatFieldValue = f.ToString
	w.FormatMessage = f.ToString
}
