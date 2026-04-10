package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/nekoffski/juno/internal/logger"
	jmcp "github.com/nekoffski/juno/internal/mcp"
	"github.com/rs/zerolog/log"
)

func main() {
	logger.Init("juno-mcp")

	cfg, err := jmcp.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	if err := jmcp.Start(ctx, cfg); err != nil {
		log.Error().Err(err).Msg("MCP server stopped")
		os.Exit(1)
	}
}
