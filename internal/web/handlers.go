package web

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/nekoffski/juno/internal/bus"
	"github.com/nekoffski/juno/internal/core"
	"github.com/nekoffski/juno/internal/device"
)

type Handlers struct {
	sender *bus.Sender
	mb     *bus.MessageBus
}

func (h *Handlers) Dashboard(c echo.Context) error {
	return c.Render(http.StatusOK, "layout.html", nil)
}

func (h *Handlers) DevicesTab(c echo.Context) error {
	f, err := h.sender.Request("device-service", device.GetDevicesRequest{})
	if err != nil {
		return fmt.Errorf("could not send get devices request: %w", err)
	}

	r, err := bus.AwaitFor[device.GetDevicesResponse](c.Request().Context(), f, bus.DefaultRequestTimeout)
	if err != nil {
		return fmt.Errorf("could not await get devices response: %w", err)
	}

	return c.Render(http.StatusOK, "devices.html", r.Devices)
}

func (h *Handlers) MetricsTab(c echo.Context) error {
	return c.Render(http.StatusOK, "metrics.html", nil)
}

func (h *Handlers) EventsTab(c echo.Context) error {
	return c.Render(http.StatusOK, "events.html", nil)
}

func (h *Handlers) PerformAction(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid device id")
	}
	action := c.Param("action")

	params := buildActionParams(action, c)

	f, err := h.sender.Request("device-service", device.PerformDeviceActionRequest{
		Id:     id,
		Action: action,
		Params: params,
	})
	if err != nil {
		return fmt.Errorf("could not send perform device action request: %w", err)
	}

	_, err = bus.AwaitFor[device.AckResponse](c.Request().Context(), f, bus.DefaultRequestTimeout)
	if err != nil {
		if err == core.ErrDeviceNotFound {
			return c.Render(http.StatusNotFound, "error.html", "Device not found")
		}
		return fmt.Errorf("could not await perform device action response: %w", err)
	}

	// Re-fetch the device and render the updated widget
	f2, err := h.sender.Request("device-service", device.GetDeviceByIdRequest{Id: id})
	if err != nil {
		return fmt.Errorf("could not send get device by id request: %w", err)
	}

	r, err := bus.AwaitFor[device.GetDeviceByIdResponse](c.Request().Context(), f2, bus.DefaultRequestTimeout)
	if err != nil {
		return fmt.Errorf("could not await get device by id response: %w", err)
	}

	return c.Render(http.StatusOK, "device_widget.html", r.Device)
}

func buildActionParams(action string, c echo.Context) map[string]any {
	switch action {
	case "brightness":
		v, err := strconv.Atoi(c.FormValue("brightness"))
		if err != nil {
			return nil
		}
		return map[string]any{"brightness": float64(v)}
	case "rgb":
		hex := c.FormValue("color")
		r, g, b := hexToRGB(hex)
		return map[string]any{
			"color": map[string]any{
				"r": float64(r),
				"g": float64(g),
				"b": float64(b),
			},
		}
	default:
		return nil
	}
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

func (h *Handlers) SSE(tmpl *template.Template) echo.HandlerFunc {
	return func(c echo.Context) error {
		sub := h.mb.NewSubscriber()
		if err := sub.Subscribe("device.events"); err != nil {
			return err
		}
		defer sub.Close()

		w := c.Response()
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.WriteHeader(http.StatusOK)

		flusher, ok := w.Writer.(http.Flusher)
		if !ok {
			return echo.NewHTTPError(http.StatusInternalServerError, "streaming not supported")
		}

		ctx := c.Request().Context()
		for {
			select {
			case ev, ok := <-sub.Events():
				if !ok {
					return nil
				}
				e, ok := ev.(device.DeviceUpdatedEvent)
				if !ok {
					continue
				}
				var buf bytes.Buffer
				if err := tmpl.ExecuteTemplate(&buf, "device_widget.html", e.Device); err != nil {
					continue
				}
				fmt.Fprintf(w, "event: device-%d\n", e.Device.Id)
				for _, line := range strings.Split(buf.String(), "\n") {
					fmt.Fprintf(w, "data: %s\n", line)
				}
				fmt.Fprint(w, "\n")
				flusher.Flush()
			case <-ctx.Done():
				return nil
			}
		}
	}
}
