package initializers

import (
	"log"
	"os"
	"strconv"

	"github.com/gobuffalo/envy"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
)

const (
	// LogLevelEnv is an environment variable name for LOG_LEVEL
	LogLevelEnv = "LOG_LEVEL"
	// EnableJSONLogsEnv is an environment variable name for ENABLE_JSON_LOGS
	EnableJSONLogsEnv = "ENABLE_JSON_LOGS"
	// DefaultLogLevel is a default LOG_LEVEL value
	DefaultLogLevel = 0
)

// InitializeLogs setups zerolog logger with consistent JSON output
func InitializeLogs() error {
	logLevel, err := strconv.Atoi(envy.Get(LogLevelEnv, "0"))
	if err != nil {
		logLevel = DefaultLogLevel
	}

	// Configure logger with UTC timestamps and caller info
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.Level(logLevel))
	
	// Always use JSON format in production
	if envy.Get("APP_ENV", "development") == "production" ||
	   envy.Get(EnableJSONLogsEnv, "true") == "true" {
		zlog.Logger = zerolog.New(os.Stderr).
			With().
			Timestamp().
			Caller().
			Logger()
	} else {
		// Development - use colored console output
		output := zerolog.ConsoleWriter{
			Out: os.Stderr,
			TimeFormat: "15:04:05",
		}
		zlog.Logger = zlog.Output(output).With().Caller().Logger()
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
