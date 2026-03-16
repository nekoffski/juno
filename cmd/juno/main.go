package main

import (
	"log"

	"github.com/nekoffski/juno/internal/device"
	"github.com/nekoffski/juno/internal/rest"
	"github.com/nekoffski/juno/internal/supervisor"
)

func main() {
	s := supervisor.NewSupervisor(
		rest.NewRestService(":6000"),
		device.NewDeviceService(),
	)
	if err := s.Run(); err != nil {
		log.Fatalf("Failed to start supervisor: %v", err)
	}
}
