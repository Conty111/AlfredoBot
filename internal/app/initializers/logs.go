package initializers

import (
	"log"
	"os"
	"strconv"

	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"

	"github.com/Conty111/AlfredoBot/internal/configs"
)

const (
	// DefaultLogLevel is a default LOG_LEVEL value
	DefaultLogLevel = 0
)

// InitializeLogs setups zerolog logger with consistent JSON output
func InitializeLogs(cfg configs.App) error {
	logLevel, err := strconv.Atoi(cfg.LogLevel)
	if err != nil {
		logLevel = DefaultLogLevel
	}

	// Configure logger with UTC timestamps and caller info
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.Level(logLevel))
	
	if cfg.Environment == "production" ||
	   cfg.JSONLogs || cfg.LogFile != "" {
		output := os.Stdout
		if cfg.LogFile != "" {
			file, err := os.OpenFile(cfg.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return err
			}
			output = file
		}
		zlog.Logger = zerolog.New(output).
			With().
			Timestamp().
			Caller().
			Logger()
	} else {
		zlog.Logger = zlog.
		Output(zerolog.ConsoleWriter{Out: os.Stdout}).
		With().
		Caller().
		Logger()
	}

	zerolog.DefaultContextLogger = &zlog.Logger
	log.SetOutput(&zerologWriter{logger: zlog.Logger})

	return nil
}

type zerologWriter struct {
	logger zerolog.Logger
}

func (w *zerologWriter) Write(p []byte) (n int, err error) {
	w.logger.Info().Msg(string(p))
	return len(p), nil
}
