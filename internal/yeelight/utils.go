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

func mapPropertyValue(prop string, value string) any {
	switch prop {
	case "rgb":
		if v, err := strconv.Atoi(value); err == nil {
			return unpackRgb(v)
		}
	case "bright":
		if v, err := strconv.Atoi(value); err == nil {
			return v
		}
	}
	return value
}

func mapProperty(prop string, value any) (string, any) {
	return mapPropertyName(prop), mapPropertyValue(prop, value.(string))
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
