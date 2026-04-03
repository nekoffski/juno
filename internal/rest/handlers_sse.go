package rest

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/nekoffski/juno/internal/bus"
	"github.com/nekoffski/juno/internal/device"
)

func newSSEHandler(mb *bus.MessageBus) echo.HandlerFunc {
	return func(c echo.Context) error {
		sub := mb.NewSubscriber()
		if err := sub.Subscribe("device.events"); err != nil {
			return err
		}
		defer sub.Close()

		w := c.Response()
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusOK)

		flusher, ok := w.Writer.(http.Flusher)
		if !ok {
			return echo.NewHTTPError(http.StatusInternalServerError, "streaming not supported")
		}
		flusher.Flush()

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
				d := convertDeviceModel(&e.Device)
				data, err := json.Marshal(d)
				if err != nil {
					continue
				}
				fmt.Fprintf(w, "event: device.updated\ndata: %s\n\n", data)
				flusher.Flush()
			case <-ctx.Done():
				return nil
			}
		}
	}
}
