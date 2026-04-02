package main

import (
	"context"
	"log"

	"github.com/nekoffski/juno/internal/core"
	"github.com/nekoffski/juno/internal/db"
	"github.com/nekoffski/juno/internal/device"
	"github.com/nekoffski/juno/internal/rest"
	"github.com/nekoffski/juno/internal/supervisor"
	"github.com/nekoffski/juno/internal/yeelight"
)

func main() {
	cfg, err := core.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	pool, err := db.Open(context.Background(), db.Config{
		Host:     cfg.DB.Host,
		Port:     cfg.DB.Port,
		User:     cfg.DB.User,
		Password: cfg.DB.Password,
		Name:     cfg.DB.Name,
	})

	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer pool.Close()

	s := supervisor.NewSupervisor(
		device.NewDeviceService(
			pool,
			map[device.DeviceVendor]device.VendorAdapter{
				device.DeviceVendorYeelight: yeelight.NewAdapter(),
			},
		),
		rest.NewRestService(cfg),
	)

	if err := s.Run(); err != nil {
		log.Fatalf("Failed to start supervisor: %v", err)
	}
}
