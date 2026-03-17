package rest

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/nekoffski/juno/internal/bus"
	"github.com/nekoffski/juno/internal/core"
)

const version = "1.0.0"

type RestService struct {
	*HealthHandlers
	port int
}

func NewRestService(cfg *core.Config) *RestService {
	return &RestService{port: cfg.RestPort}
}

func (s *RestService) Name() string {
	return "rest"
}

func (s *RestService) Init(ctx context.Context, mb *bus.MessageBus) error {
	s.HealthHandlers = &HealthHandlers{sender: mb.NewSender()}
	return nil
}

func (s *RestService) Run(ctx context.Context) error {
	e := echo.New()
	e.HideBanner = true
	RegisterHandlers(e, NewStrictHandler(s, nil))

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = e.Shutdown(shutdownCtx)
	}()

	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("Starting REST api on %s", addr)

	if err := e.Start(addr); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
