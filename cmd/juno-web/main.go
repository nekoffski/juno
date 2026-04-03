package main

import (
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/nekoffski/juno/internal/web"
)

func main() {
	restBase := envOr("JUNO_REST_BASE_URL", "http://localhost:6000")
	webPort := envIntOr("JUNO_WEB_PORT", 6001)

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
		log.Fatalf("failed to parse templates: %v", err)
	}

	h := web.NewHandlers(restBase, tmpl)

	staticSub, err := fs.Sub(web.TemplateFS, "static")
	if err != nil {
		log.Fatalf("failed to get static sub-fs: %v", err)
	}

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/static/*", echo.WrapHandler(
		http.StripPrefix("/static/", http.FileServer(http.FS(staticSub))),
	))

	e.GET("/", h.Dashboard)
	e.GET("/tabs/devices", h.DevicesTab)
	e.GET("/tabs/metrics", h.MetricsTab)
	e.GET("/tabs/events", h.EventsTab)
	e.POST("/device/:id/action/:action", h.PerformAction)
	e.GET("/sse", h.SSE)

	log.Fatal(e.Start(fmt.Sprintf(":%d", webPort)))
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
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
