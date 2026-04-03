package device

import (
	"github.com/nekoffski/juno/internal/core"
)

type GetDevicesRequest struct{}

type GetDevicesResponse struct {
	Devices []*DeviceModel
}

type GetDeviceByIdRequest struct {
	Id int `json:"id"`
}

type GetDeviceByIdResponse struct {
	Device *DeviceModel `json:"device"`
}

type DiscoverDevicesRequest struct{}

type DeleteDeviceRequest struct {
	Id int `json:"id"`
}

type AckResponse struct{}

type GetDevicePropertiesRequest struct {
	Id         int      `json:"id"`
	Properties []string `json:"properties"`
}

type GetDevicePropertiesResponse struct {
	Properties map[string]any `json:"properties"`
}

type PerformDeviceActionRequest struct {
	Id     int            `json:"id"`
	Action string         `json:"action"`
	Params map[string]any `json:"params"`
}

type DeviceUpdatedEvent struct {
	Device DeviceModel
}

func parseActionParams(name string, payload any) (any, error) {
	m, _ := payload.(map[string]any)

	switch name {
	case "on", "off", "toggle":
		return nil, nil

	case "rgb":
		colorMap, ok := m["color"].(map[string]any)

		if !ok {
			return nil, core.ErrInvalidArguments
		}

		r, ok1 := colorMap["r"].(float64)
		g, ok2 := colorMap["g"].(float64)
		b, ok3 := colorMap["b"].(float64)

		if !ok1 || !ok2 || !ok3 {
			return nil, core.ErrInvalidArguments
		}

		return ColorRGB{R: int(r), G: int(g), B: int(b)}, nil

	case "brightness":
		v, ok := m["brightness"].(float64)
		if !ok {
			return nil, core.ErrInvalidArguments
		}
		return int(v), nil

	case "ct":
		v, ok := m["ct"].(float64)
		if !ok {
			return nil, core.ErrInvalidArguments
		}
		return int(v), nil

	default:
		return nil, core.ErrInvalidArguments
	}
}
