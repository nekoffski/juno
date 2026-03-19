package rest

import (
	"context"
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/nekoffski/juno/internal/bus"
	"github.com/nekoffski/juno/internal/core"
	"github.com/nekoffski/juno/internal/device"
)

type DeviceHandlers struct {
	sender *bus.Sender
}

func convertDeviceModel(d *device.DeviceModel) Device {
	return Device{
		Id:           d.Id,
		Name:         d.Name,
		Status:       d.Status,
		Capabilities: d.Capabilities,
		Vendor:       d.Vendor,
	}
}

func (h *DeviceHandlers) GetDevices(
	ctx context.Context,
	_ GetDevicesRequestObject,
) (GetDevicesResponseObject, error) {
	f, err := h.sender.Request("device-service", device.GetDevicesRequest{})
	if err != nil {
		log.Errorf("Could not send get devices request: %v", err)
		return nil, echo.NewHTTPError(500, fmt.Sprintf("Failed to send get devices request: %v", err))
	}

	r, err := bus.AwaitFor[device.GetDevicesResponse](ctx, f, bus.DefaultRequestTimeout)
	if err != nil {
		log.Errorf("Failed to await get devices response: %v", err)
		return nil, echo.NewHTTPError(500, fmt.Sprintf("Failed to await get devices response: %v", err))
	}

	res := GetDevices200JSONResponse{}
	for _, d := range r.Devices {
		res = append(res, convertDeviceModel(d))
	}

	return res, nil

}

func (h *DeviceHandlers) GetDeviceById(
	ctx context.Context,
	req GetDeviceByIdRequestObject,
) (GetDeviceByIdResponseObject, error) {
	f, err := h.sender.Request("device-service", device.GetDeviceByIdRequest{Id: req.Id})
	if err != nil {
		log.Errorf("Could not send get device by id request: %v", err)
		return nil, echo.NewHTTPError(500, fmt.Sprintf("Failed to send get device by id request: %v", err))
	}

	r, err := bus.AwaitFor[device.GetDeviceByIdResponse](ctx, f, bus.DefaultRequestTimeout)

	if err != nil {
		if err == core.ErrDeviceNotFound {
			return GetDeviceById404Response{}, nil
		}
		log.Errorf("Failed to await get device by id response: %v", err)
		return nil, echo.NewHTTPError(500, fmt.Sprintf("Failed to await get device by id response: %v", err))
	}

	res := GetDeviceById200JSONResponse{
		Id:           r.Device.Id,
		Name:         r.Device.Name,
		Status:       r.Device.Status,
		Capabilities: r.Device.Capabilities,
		Vendor:       r.Device.Vendor,
	}
	return res, nil
}

func (h *DeviceHandlers) DiscoverDevices(
	_ context.Context,
	_ DiscoverDevicesRequestObject,
) (DiscoverDevicesResponseObject, error) {
	if err := h.sender.Send("device-service", device.DiscoverDevicesRequest{}); err != nil {
		log.Errorf("Could not send discover request: %v", err)
		return nil, echo.NewHTTPError(500, fmt.Sprintf("Failed to send discover request: %v", err))
	}
	return DiscoverDevices202Response{}, nil
}
