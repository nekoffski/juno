package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	jmcp "github.com/nekoffski/juno/internal/mcp"
)

func main() {
	cfg, err := jmcp.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	if err := jmcp.Start(ctx, cfg); err != nil {
		log.Printf("MCP server stopped: %v", err)
		os.Exit(1)
	}
}
