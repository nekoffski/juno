package logger

import (
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func Init(service string) {
	level := zerolog.InfoLevel
	if v := os.Getenv("JUNO_LOG_LEVEL"); v != "" {
		if l, err := zerolog.ParseLevel(strings.ToLower(v)); err == nil {
			level = l
		}
	}
	zerolog.SetGlobalLevel(level)
	log.Logger = zerolog.New(os.Stdout).
		With().
		Timestamp().
		Str("service", service).
		Logger()
}
