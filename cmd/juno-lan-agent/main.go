package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/nekoffski/juno/internal/lan"
)

func main() {
	addr := os.Getenv("JUNO_LAN_AGENT_ADDR")
	if addr == "" {
		addr = "0.0.0.0"
	}

	port := 7000
	if v := os.Getenv("JUNO_LAN_AGENT_PORT"); v != "" {
		p, err := strconv.Atoi(v)
		if err != nil {
			log.Fatalf("invalid JUNO_LAN_AGENT_PORT: %v", err)
		}
		port = p
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	svc := lan.NewService(lan.Config{Addr: addr, Port: port})

	log.Printf("juno-lan-agent starting on %s:%d", addr, port)
	if err := svc.Run(ctx); err != nil {
		log.Fatalf("juno-lan-agent exited with error: %v", err)
	}
}
