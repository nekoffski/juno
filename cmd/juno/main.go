package main

import (
	"log"

	"github.com/nekoffski/juno/internal/core"
	"github.com/nekoffski/juno/internal/device"
	"github.com/nekoffski/juno/internal/rest"
	"github.com/nekoffski/juno/internal/supervisor"
)

func main() {
	cfg, err := core.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	s := supervisor.NewSupervisor(
		rest.NewRestService(cfg),
		device.NewDeviceService(),
	)
	if err := s.Run(); err != nil {
		log.Fatalf("Failed to start supervisor: %v", err)
	}
}
