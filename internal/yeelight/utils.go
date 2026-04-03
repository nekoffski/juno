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

const (
	ctMin = 1700
	ctMax = 6500
)

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
	case "ct":
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
		return "set_rgb", []any{packRgb(color), "smooth", 500}, nil

	case "brightness":
		brightness, ok := action.Params.(int)
		if !ok {
			return "", nil, core.ErrInvalidArguments
		}
		return "set_bright", []any{brightness, "smooth", 500}, nil

	case "ct":
		ct, ok := action.Params.(int)
		if !ok {
			return "", nil, core.ErrInvalidArguments
		}
		if ct < ctMin {
			ct = ctMin
		} else if ct > ctMax {
			ct = ctMax
		}
		return "set_ct_abx", []any{ct, "smooth", 500}, nil

	default:
		return "", nil, core.ErrDeviceNotCapable
	}
}
