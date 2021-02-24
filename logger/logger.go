package logger

// Uses a singleton design patten, has package state but easy to access anywhere

import (
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
)

var logger zerolog.Logger

// Init should be called when the app is initilised
func Init() {
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC1123}

	logger = zerolog.New(output).With().Timestamp().Logger()
}

// Debug prints at debug level
func Debug(format string, v ...interface{}) {
	logger.Debug().Msg(fmt.Sprintf(format, v...))
}

// Info prints at info level
func Info(format string, v ...interface{}) {
	logger.Info().Msg(fmt.Sprintf(format, v...))
}

// Warn prints at warn level
func Warn(format string, v ...interface{}) {
	logger.Warn().Msg(fmt.Sprintf(format, v...))
}

// Error prints at error level
func Error(format string, v ...interface{}) {
	logger.Error().Msg(fmt.Sprintf(format, v...))
}

// Fatal prints at fatal level and calls os.Exit(1) and stops execution
// Use sparingly
func Fatal(format string, v ...interface{}) {
	logger.Fatal().Msg(fmt.Sprintf(format, v...))
}
