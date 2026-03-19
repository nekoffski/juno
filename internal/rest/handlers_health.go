package rest

import (
	"context"
	"time"

	"github.com/nekoffski/juno/internal/bus"
	"github.com/nekoffski/juno/internal/core"
)

type HealthHandlers struct {
	sender *bus.Sender
}

func (h *HealthHandlers) GetHealth(
	_ context.Context,
	_ GetHealthRequestObject,
) (GetHealthResponseObject, error) {
	return GetHealth200JSONResponse{
		Status: "ok",
	}, nil
}

func (h *HealthHandlers) GetDeviceServiceHealth(
	ctx context.Context,
	_ GetDeviceServiceHealthRequestObject,
) (GetDeviceServiceHealthResponseObject, error) {
	f, err := h.sender.Request("device-service", core.HeartbeatRequest{Magic: "ping"})
	r, err := bus.AwaitFor[core.HeartbeatResponse](ctx, f, time.Second)

	if err != nil {
		return GetDeviceServiceHealth200JSONResponse{
			Status: "unhealthy",
		}, err
	}

	status := "ok"
	if !r.Healthy {
		status = "unhealthy"
	}

	return GetDeviceServiceHealth200JSONResponse{
		Status: status,
	}, nil
}
