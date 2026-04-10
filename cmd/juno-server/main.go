package main

import (
	"context"

	"github.com/nekoffski/juno/internal/core"
	"github.com/nekoffski/juno/internal/db"
	"github.com/nekoffski/juno/internal/device"
	"github.com/nekoffski/juno/internal/logger"
	"github.com/nekoffski/juno/internal/rest"
	"github.com/nekoffski/juno/internal/supervisor"
	"github.com/nekoffski/juno/internal/yeelight"
	"github.com/rs/zerolog/log"
)

func main() {
	logger.Init("juno-server")

	cfg, err := core.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	pool, err := db.Open(context.Background(), db.Config{
		Host:     cfg.DB.Host,
		Port:     cfg.DB.Port,
		User:     cfg.DB.User,
		Password: cfg.DB.Password,
		Name:     cfg.DB.Name,
	})

	if err != nil {
		log.Fatal().Err(err).Msg("failed to open database")
	}
	defer pool.Close()

	s := supervisor.NewSupervisor(
		device.NewDeviceService(
			device.NewPgxRepository(pool),
			map[device.DeviceVendor]device.VendorAdapter{
				device.DeviceVendorYeelight: yeelight.NewAdapter(cfg.YeelightSsdpAddr, cfg.LanAgentURL),
			},
		),
		rest.NewRestService(cfg),
	)

	if err := s.Run(); err != nil {
		log.Fatal().Err(err).Msg("supervisor failed")
	}
}
