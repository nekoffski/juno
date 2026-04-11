package main

import (
	"context"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/nekoffski/juno/internal/lan"
	"github.com/nekoffski/juno/internal/logger"
	"github.com/rs/zerolog/log"
)

func main() {
	logger.Init("juno-lan-agent")

	addr := "0.0.0.0"
	port := 7000

	if v := os.Getenv("JUNO_LAN_AGENT_PORT"); v != "" {
		p, err := strconv.Atoi(v)
		if err != nil {
			log.Fatal().Err(err).Msg("invalid JUNO_LAN_AGENT_PORT")
		}
		port = p
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	svc := lan.NewService(lan.Config{Addr: addr, Port: port})

	log.Info().Str("addr", addr).Int("port", port).Msg("juno-lan-agent starting")
	if err := svc.Run(ctx); err != nil {
		log.Fatal().Err(err).Msg("juno-lan-agent exited with error")
	}
}
