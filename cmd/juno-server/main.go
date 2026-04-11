package main

import (
	"context"
	"fmt"

	"github.com/nekoffski/juno/internal/core"
	"github.com/nekoffski/juno/internal/db"
	"github.com/nekoffski/juno/internal/device"
	"github.com/nekoffski/juno/internal/logger"
	"github.com/nekoffski/juno/internal/rest"
	"github.com/nekoffski/juno/internal/supervisor"
	"github.com/nekoffski/juno/internal/yeelight"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
)

func main() {
	logger.Init("juno-server")

	cfg, err := core.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	registry := prometheus.NewRegistry()
	tracer := db.NewQueryTracer(registry)

	pool, err := db.Open(context.Background(), db.Config{
		Host:     cfg.DB.Host,
		Port:     cfg.DB.Port,
		User:     cfg.DB.User,
		Password: cfg.DB.Password,
		Name:     cfg.DB.Name,
	}, tracer)

	if err != nil {
		log.Fatal().Err(err).Msg("failed to open database")
	}
	defer pool.Close()

	db.NewPoolCollector(pool, registry)

	s := supervisor.NewSupervisor(
		device.NewDeviceService(
			device.NewPgxRepository(pool),
			map[device.DeviceVendor]device.VendorAdapter{
				device.DeviceVendorYeelight: yeelight.NewAdapter(yeelightSsdpAddr(cfg), lanAgentURL(cfg)),
			},
		),
		rest.NewRestService(cfg, registry),
	)

	if err := s.Run(); err != nil {
		log.Fatal().Err(err).Msg("supervisor failed")
	}
}

func yeelightSsdpAddr(cfg *core.Config) string {
	return fmt.Sprintf("%s:%d", cfg.YeelightSsdpAddr, cfg.YeelightSsdpPort)
}

func lanAgentURL(cfg *core.Config) string {
	if cfg.LanAgentAddr == "" {
		return ""
	}
	return fmt.Sprintf("http://%s:%d", cfg.LanAgentAddr, cfg.LanAgentPort)
}
