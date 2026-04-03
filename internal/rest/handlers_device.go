package rest

import (
	"context"
	"fmt"
	"maps"

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
		Properties:   d.Properties,
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
		Properties:   r.Device.Properties,
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

func (h *DeviceHandlers) GetDeviceProperties(
	ctx context.Context,
	request GetDevicePropertiesRequestObject,
) (GetDevicePropertiesResponseObject, error) {
	var fields []string
	if request.Params.Fields != nil {
		fields = *request.Params.Fields
	}
	f, err := h.sender.Request("device-service", device.GetDevicePropertiesRequest{
		Id:         request.Id,
		Properties: fields,
	})
	if err != nil {
		log.Errorf("Could not send get device properties request: %v", err)
		return nil, echo.NewHTTPError(500, fmt.Sprintf("Failed to send get device properties request: %v", err))
	}

	r, err := bus.AwaitFor[device.GetDevicePropertiesResponse](ctx, f, bus.DefaultRequestTimeout)
	if err != nil {
		if err == core.ErrDeviceNotFound {
			return GetDeviceProperties404Response{}, nil
		}
		log.Errorf("Failed to await get device properties response: %v", err)
		return nil, echo.NewHTTPError(500, fmt.Sprintf("Failed to await get device properties response: %v", err))
	}

	res := GetDeviceProperties200JSONResponse{}
	maps.Copy(res, r.Properties)
	return res, nil
}

func (h *DeviceHandlers) DeleteDevice(
	ctx context.Context,
	req DeleteDeviceRequestObject,
) (DeleteDeviceResponseObject, error) {
	f, err := h.sender.Request("device-service", device.DeleteDeviceRequest{Id: req.Id})
	if err != nil {
		log.Errorf("Could not send delete device request: %v", err)
		return nil, echo.NewHTTPError(500, fmt.Sprintf("Failed to send delete device request: %v", err))
	}

	_, err = bus.AwaitFor[device.AckResponse](ctx, f, bus.DefaultRequestTimeout)
	if err != nil {
		if err == core.ErrDeviceNotFound {
			return DeleteDevice404Response{}, nil
		}
		log.Errorf("Failed to await delete device response: %v", err)
		return nil, echo.NewHTTPError(500, fmt.Sprintf("Failed to await delete device response: %v", err))
	}
	return DeleteDevice200Response{}, nil
}

func (h *DeviceHandlers) PerformDeviceAction(
	ctx context.Context,
	request PerformDeviceActionRequestObject,
) (PerformDeviceActionResponseObject, error) {
	var params map[string]any
	if request.Body != nil && request.Body.Params != nil {
		params = *request.Body.Params
	}

	f, err := h.sender.Request("device-service", device.PerformDeviceActionRequest{
		Id:     request.Id,
		Action: request.Action,
		Params: params,
	})
	if err != nil {
		log.Errorf("Could not send perform device action request: %v", err)
		return nil, echo.NewHTTPError(500, fmt.Sprintf("Failed to send perform device action request: %v", err))
	}

	_, err = bus.AwaitFor[device.AckResponse](ctx, f, bus.DefaultRequestTimeout)
	if err != nil {
		switch err {
		case core.ErrDeviceNotFound:
			return PerformDeviceAction404Response{}, nil
		}
		log.Errorf("Failed to await perform device action response: %v", err)
		return nil, echo.NewHTTPError(500, fmt.Sprintf("Failed to await perform device action response: %v", err))
	}
	return PerformDeviceAction200Response{}, nil

}
