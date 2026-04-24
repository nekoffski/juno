package web

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

type Device struct {
	Id           int                    `json:"id"`
	Name         string                 `json:"name"`
	Vendor       interface{}            `json:"vendor"`
	Status       interface{}            `json:"status"`
	Capabilities []string               `json:"capabilities"`
	Properties   map[string]interface{} `json:"properties"`
}

type Handlers struct {
	restBase string
	client   *http.Client
	tmpl     *template.Template
}

func NewHandlers(restBase string, tmpl *template.Template) *Handlers {
	return &Handlers{
		restBase: restBase,
		client:   &http.Client{},
		tmpl:     tmpl,
	}
}

func (h *Handlers) Dashboard(c echo.Context) error {
	c.Response().Header().Set("Content-Type", "text/html; charset=utf-8")
	return h.tmpl.ExecuteTemplate(c.Response().Writer, "layout.html", nil)
}

func (h *Handlers) DevicesTab(c echo.Context) error {
	resp, err := h.client.Get(h.restBase + "/device")
	if err != nil {
		return fmt.Errorf("could not get devices: %w", err)
	}
	defer resp.Body.Close()

	var devices []Device
	if err := json.NewDecoder(resp.Body).Decode(&devices); err != nil {
		return fmt.Errorf("could not decode devices: %w", err)
	}

	c.Response().Header().Set("Content-Type", "text/html; charset=utf-8")
	return h.tmpl.ExecuteTemplate(c.Response().Writer, "devices.html", devices)
}

func (h *Handlers) Discover(c echo.Context) error {
	_, err := h.client.Post(h.restBase+"/device/discover", "application/json", nil)
	if err != nil {
		return fmt.Errorf("could not trigger discovery: %w", err)
	}
	return c.NoContent(http.StatusAccepted)
}

func (h *Handlers) MetricsTab(c echo.Context) error {
	c.Response().Header().Set("Content-Type", "text/html; charset=utf-8")
	return h.tmpl.ExecuteTemplate(c.Response().Writer, "metrics.html", nil)
}

func (h *Handlers) EventsTab(c echo.Context) error {
	c.Response().Header().Set("Content-Type", "text/html; charset=utf-8")
	return h.tmpl.ExecuteTemplate(c.Response().Writer, "events.html", nil)
}

func (h *Handlers) PerformAction(c echo.Context) error {
	id := c.Param("id")
	action := c.Param("action")

	body, err := buildActionBody(action, c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	url := fmt.Sprintf("%s/device/id/%s/action/%s", h.restBase, id, action)
	resp, err := h.client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("could not perform action: %w", err)
	}
	defer resp.Body.Close()

	deviceResp, err := h.client.Get(fmt.Sprintf("%s/device/id/%s", h.restBase, id))
	if err != nil {
		return fmt.Errorf("could not get device: %w", err)
	}
	defer deviceResp.Body.Close()

	var d Device
	if err := json.NewDecoder(deviceResp.Body).Decode(&d); err != nil {
		return fmt.Errorf("could not decode device: %w", err)
	}

	c.Response().Header().Set("Content-Type", "text/html; charset=utf-8")
	return h.tmpl.ExecuteTemplate(c.Response().Writer, "device_widget.html", d)
}

func buildActionBody(action string, c echo.Context) ([]byte, error) {
	var params map[string]interface{}
	switch action {
	case "toggle":
		params = nil
	case "brightness":
		v, err := strconv.Atoi(c.FormValue("brightness"))
		if err != nil {
			return nil, fmt.Errorf("invalid brightness value")
		}
		params = map[string]interface{}{"brightness": float64(v)}
	case "ct":
		v, err := strconv.Atoi(c.FormValue("ct"))
		if err != nil {
			return nil, fmt.Errorf("invalid ct value")
		}
		params = map[string]interface{}{"ct": float64(v)}
	case "rgb":
		hex := c.FormValue("color")
		r, g, b := hexToRGB(hex)
		params = map[string]interface{}{
			"color": map[string]interface{}{
				"r": float64(r),
				"g": float64(g),
				"b": float64(b),
			},
		}
	default:
		params = nil
	}

	body := map[string]interface{}{}
	if params != nil {
		body["params"] = params
	}
	return json.Marshal(body)
}

func hexToRGB(hex string) (r, g, b int) {
	if len(hex) > 0 && hex[0] == '#' {
		hex = hex[1:]
	}
	if len(hex) != 6 {
		return 0, 0, 0
	}
	val, err := strconv.ParseInt(hex, 16, 32)
	if err != nil {
		return 0, 0, 0
	}
	return int(val >> 16 & 0xff), int(val >> 8 & 0xff), int(val & 0xff)
}

func (h *Handlers) SSE(c echo.Context) error {
	restResp, err := h.client.Get(h.restBase + "/events")
	if err != nil {
		return fmt.Errorf("could not connect to REST SSE: %w", err)
	}
	defer restResp.Body.Close()

	w := c.Response().Writer
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	flusher, ok := w.(http.Flusher)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "streaming not supported")
	}
	flusher.Flush()

	ctx := c.Request().Context()
	scanner := bufio.NewScanner(restResp.Body)
	var eventType, dataLine string

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "event:"):
			eventType = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		case strings.HasPrefix(line, "data:"):
			dataLine = strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		case line == "" && eventType != "" && dataLine != "":
			switch eventType {
			case "device.updated":
				var d Device
				if err := json.Unmarshal([]byte(dataLine), &d); err == nil {
					var buf bytes.Buffer
					if err := h.tmpl.ExecuteTemplate(&buf, "device_widget.html", d); err == nil {
						html := buf.String()
						fmt.Fprintf(w, "event: device-%d\n", d.Id)
						for _, l := range strings.Split(html, "\n") {
							fmt.Fprintf(w, "data: %s\n", l)
						}
						fmt.Fprint(w, "\n")
						flusher.Flush()
					}
				}
			case "device.added":
				var d Device
				if err := json.Unmarshal([]byte(dataLine), &d); err == nil {
					var buf bytes.Buffer
					if err := h.tmpl.ExecuteTemplate(&buf, "device_widget.html", d); err == nil {
						html := buf.String()
						fmt.Fprint(w, "event: device.added\n")
						for _, l := range strings.Split(html, "\n") {
							fmt.Fprintf(w, "data: %s\n", l)
						}
						fmt.Fprint(w, "\n")
						flusher.Flush()
					}
				}
			}
			eventType = ""
			dataLine = ""
		}
	}

	return nil
}
