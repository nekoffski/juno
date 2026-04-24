package main

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"os"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/nekoffski/juno/internal/logger"
	"github.com/nekoffski/juno/internal/web"
	"github.com/rs/zerolog/log"
)

func main() {
	logger.Init("juno-web")

	restPort := envIntOr("JUNO_REST_PORT", 6001)
	restBase := fmt.Sprintf("http://127.0.0.1:%d", restPort)
	webPort := envIntOr("JUNO_WEB_PORT", 6002)

	tmpl, err := template.New("").Funcs(template.FuncMap{
		"rgbHex": func(v interface{}) string {
			m, ok := v.(map[string]interface{})
			if !ok {
				return "#000000"
			}
			r := int(toFloat(m["r"]))
			g := int(toFloat(m["g"]))
			b := int(toFloat(m["b"]))
			return fmt.Sprintf("#%02x%02x%02x", r, g, b)
		},
		"propInt": func(v interface{}) int {
			return int(toFloat(v))
		},
	}).ParseFS(web.TemplateFS, "templates/*.html")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to parse templates")
	}

	h := web.NewHandlers(restBase, tmpl)

	staticSub, err := fs.Sub(web.TemplateFS, "static")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get static sub-fs")
	}

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Logger.SetOutput(io.Discard)
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogMethod: true,
		LogURI:    true,
		LogStatus: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			log.Info().Str("method", v.Method).Str("uri", v.URI).Int("status", v.Status).Msg("request")
			return nil
		},
	}))
	e.Use(middleware.Recover())

	e.GET("/static/*", echo.WrapHandler(
		http.StripPrefix("/static/", http.FileServer(http.FS(staticSub))),
	))

	e.GET("/", h.Dashboard)
	e.GET("/tabs/devices", h.DevicesTab)
	e.GET("/tabs/metrics", h.MetricsTab)
	e.GET("/tabs/events", h.EventsTab)
	e.POST("/device/:id/action/:action", h.PerformAction)
	e.POST("/discover", h.Discover)
	e.GET("/sse", h.SSE)

	log.Fatal().Err(e.Start(fmt.Sprintf(":%d", webPort))).Msg("web server stopped")
}

func envIntOr(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func toFloat(v interface{}) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	case int64:
		return float64(n)
	default:
		return 0
	}
}
