package rest

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/nekoffski/juno/internal/bus"
	"github.com/nekoffski/juno/internal/core"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
)

type RestService struct {
	*HealthHandlers
	*DeviceHandlers
	mb          *bus.MessageBus
	port        int
	metricsPort int
	registry    *prometheus.Registry
}

func NewRestService(cfg *core.Config, registry *prometheus.Registry) *RestService {
	return &RestService{
		port:        cfg.RestPort,
		metricsPort: cfg.MetricsPort,
		registry:    registry,
	}
}

func (s *RestService) Name() string {
	return "rest"
}

func (s *RestService) Init(ctx context.Context, mb *bus.MessageBus) error {
	s.mb = mb
	s.HealthHandlers = &HealthHandlers{sender: mb.NewSender()}
	s.DeviceHandlers = &DeviceHandlers{sender: mb.NewSender()}
	return nil
}

func (s *RestService) Run(ctx context.Context) error {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Logger.SetOutput(io.Discard)
	e.Use(echoprometheus.NewMiddlewareWithConfig(echoprometheus.MiddlewareConfig{
		Registerer: s.registry,
		Subsystem:  "juno",
	}))
	RegisterHandlers(e, NewStrictHandler(s, nil))
	e.GET("/events", newSSEHandler(s.mb))

	metrics := echo.New()
	metrics.HideBanner = true
	metrics.HidePort = true
	metrics.Logger.SetOutput(io.Discard)
	metrics.GET("/metrics", echoprometheus.NewHandlerWithConfig(echoprometheus.HandlerConfig{
		Gatherer: s.registry,
	}))

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = e.Shutdown(shutdownCtx)
		_ = metrics.Shutdown(shutdownCtx)
	}()

	metricsAddr := fmt.Sprintf(":%d", s.metricsPort)
	log.Info().Str("addr", metricsAddr).Msg("starting metrics server")
	go func() {
		if err := metrics.Start(metricsAddr); !errors.Is(err, http.ErrServerClosed) {
			log.Error().Err(err).Msg("metrics server error")
		}
	}()

	addr := fmt.Sprintf(":%d", s.port)
	log.Info().Str("addr", addr).Msg("starting REST API")

	if err := e.Start(addr); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
