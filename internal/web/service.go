package web

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/nekoffski/juno/internal/bus"
	"github.com/nekoffski/juno/internal/core"
)

type templateRenderer struct {
	templates *template.Template
}

func (t *templateRenderer) Render(w io.Writer, name string, data any, _ echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

type WebService struct {
	handlers *Handlers
	port     int
}

func NewWebService(cfg *core.Config) *WebService {
	return &WebService{port: cfg.WebPort}
}

func (s *WebService) Name() string {
	return "web"
}

func (s *WebService) Init(_ context.Context, mb *bus.MessageBus) error {
	s.handlers = &Handlers{sender: mb.NewSender()}
	return nil
}

func (s *WebService) Run(ctx context.Context) error {
	tmpl := template.Must(template.ParseFS(templateFS, "templates/*.html"))

	e := echo.New()
	e.HideBanner = true
	e.Renderer = &templateRenderer{templates: tmpl}

	e.GET("/", s.handlers.Dashboard)
	e.GET("/tabs/devices", s.handlers.DevicesTab)
	e.GET("/tabs/metrics", s.handlers.MetricsTab)
	e.GET("/tabs/events", s.handlers.EventsTab)
	e.POST("/device/:id/action/:action", s.handlers.PerformAction)

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = e.Shutdown(shutdownCtx)
	}()

	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("Starting Web UI on %s", addr)

	if err := e.Start(addr); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
