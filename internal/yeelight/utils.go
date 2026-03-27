package yeelight

import (
	"strconv"

	"github.com/nekoffski/juno/internal/core"
	"github.com/nekoffski/juno/internal/device"
)

func unpackRgb(color int) device.ColorRGB {
	r := (color >> 16) & 0xFF
	g := (color >> 8) & 0xFF
	b := color & 0xFF
	return device.ColorRGB{R: r, G: g, B: b}
}

func packRgb(color device.ColorRGB) int {
	return (color.R << 16) | (color.G << 8) | color.B
}

func mapPropertyName(yeelightProp string) string {
	switch yeelightProp {
	case "bright":
		return "brightness"
	default:
		return yeelightProp
	}
}

func mapPropertyValue(prop string, value any) any {
	switch prop {
	case "rgb":
		switch v := value.(type) {
		case float64:
			return unpackRgb(int(v))
		case string:
			if n, err := strconv.Atoi(v); err == nil {
				return unpackRgb(n)
			}
		}
	case "bright":
		switch v := value.(type) {
		case float64:
			return int(v)
		case string:
			if n, err := strconv.Atoi(v); err == nil {
				return n
			}
		}
	}
	return value
}

func mapProperty(prop string, value any) (string, any) {
	return mapPropertyName(prop), mapPropertyValue(prop, value)
}

func toYeelightAction(action device.Action) (string, []any, error) {
	switch action.Method {
	case "on", "off", "toggle":
		return action.Method, []any{}, nil

	case "rgb":
		color, ok := action.Params.(device.ColorRGB)
		if !ok {
			return "", nil, core.ErrInvalidArguments
		}
		return "set_rgb", []any{packRgb(color)}, nil

	case "brightness":
		brightness, ok := action.Params.(int)
		if !ok {
			return "", nil, core.ErrInvalidArguments
		}
		return "bright", []any{brightness}, nil

	default:
		return "", nil, core.ErrDeviceNotCapable
	}
}
