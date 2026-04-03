package yeelight

import (
	"testing"

	"github.com/nekoffski/juno/internal/core"
	"github.com/nekoffski/juno/internal/device"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnpackRgb(t *testing.T) {
	tests := []struct {
		packed int
		want   device.ColorRGB
	}{
		{0xFF8040, device.ColorRGB{R: 255, G: 128, B: 64}},
		{0x000000, device.ColorRGB{R: 0, G: 0, B: 0}},
		{0xFFFFFF, device.ColorRGB{R: 255, G: 255, B: 255}},
		{0x010203, device.ColorRGB{R: 1, G: 2, B: 3}},
	}
	for _, tc := range tests {
		assert.Equal(t, tc.want, unpackRgb(tc.packed))
	}
}

func TestPackRgb(t *testing.T) {
	tests := []struct {
		color device.ColorRGB
		want  int
	}{
		{device.ColorRGB{R: 255, G: 128, B: 64}, 0xFF8040},
		{device.ColorRGB{R: 0, G: 0, B: 0}, 0x000000},
		{device.ColorRGB{R: 255, G: 255, B: 255}, 0xFFFFFF},
	}
	for _, tc := range tests {
		assert.Equal(t, tc.want, packRgb(tc.color))
	}
}

func TestPackUnpackRgbRoundtrip(t *testing.T) {
	original := device.ColorRGB{R: 12, G: 34, B: 56}
	assert.Equal(t, original, unpackRgb(packRgb(original)))
}

func TestMapPropertyName_Bright(t *testing.T) {
	assert.Equal(t, "brightness", mapPropertyName("bright"))
}

func TestMapPropertyName_Passthrough(t *testing.T) {
	assert.Equal(t, "power", mapPropertyName("power"))
	assert.Equal(t, "rgb", mapPropertyName("rgb"))
	assert.Equal(t, "ct", mapPropertyName("ct"))
}

func TestMapPropertyValue_RgbFloat64(t *testing.T) {
	v := mapPropertyValue("rgb", float64(0xFF8040))
	assert.Equal(t, device.ColorRGB{R: 255, G: 128, B: 64}, v)
}

func TestMapPropertyValue_RgbString(t *testing.T) {
	v := mapPropertyValue("rgb", "16744512")
	assert.Equal(t, device.ColorRGB{R: 255, G: 128, B: 64}, v)
}

func TestMapPropertyValue_RgbInvalidString(t *testing.T) {
	v := mapPropertyValue("rgb", "notanumber")
	assert.Equal(t, "notanumber", v)
}

func TestMapPropertyValue_BrightFloat64(t *testing.T) {
	v := mapPropertyValue("bright", float64(75))
	assert.Equal(t, 75, v)
}

func TestMapPropertyValue_BrightString(t *testing.T) {
	v := mapPropertyValue("bright", "42")
	assert.Equal(t, 42, v)
}

func TestMapPropertyValue_BrightInvalidString(t *testing.T) {
	v := mapPropertyValue("bright", "notanumber")
	assert.Equal(t, "notanumber", v)
}

func TestMapPropertyValue_UnknownProp(t *testing.T) {
	v := mapPropertyValue("power", "on")
	assert.Equal(t, "on", v)
}

func TestMapProperty_BrightRenamesAndConverts(t *testing.T) {
	name, value := mapProperty("bright", float64(50))
	assert.Equal(t, "brightness", name)
	assert.Equal(t, 50, value)
}

func TestMapProperty_Passthrough(t *testing.T) {
	name, value := mapProperty("power", "on")
	assert.Equal(t, "power", name)
	assert.Equal(t, "on", value)
}

func TestToYeelightAction_On(t *testing.T) {
	method, params, err := toYeelightAction(device.Action{Method: "on"})
	require.NoError(t, err)
	assert.Equal(t, "on", method)
	assert.Empty(t, params)
}

func TestToYeelightAction_Off(t *testing.T) {
	method, params, err := toYeelightAction(device.Action{Method: "off"})
	require.NoError(t, err)
	assert.Equal(t, "off", method)
	assert.Empty(t, params)
}

func TestToYeelightAction_Toggle(t *testing.T) {
	method, params, err := toYeelightAction(device.Action{Method: "toggle"})
	require.NoError(t, err)
	assert.Equal(t, "toggle", method)
	assert.Empty(t, params)
}

func TestToYeelightAction_Rgb(t *testing.T) {
	color := device.ColorRGB{R: 255, G: 128, B: 64}
	method, params, err := toYeelightAction(device.Action{Method: "rgb", Params: color})
	require.NoError(t, err)
	assert.Equal(t, "set_rgb", method)
	require.Len(t, params, 3)
	assert.Equal(t, packRgb(color), params[0])
	assert.Equal(t, "smooth", params[1])
	assert.Equal(t, 500, params[2])
}

func TestToYeelightAction_RgbInvalidParams(t *testing.T) {
	_, _, err := toYeelightAction(device.Action{Method: "rgb", Params: "notacolor"})
	assert.ErrorIs(t, err, core.ErrInvalidArguments)
}

func TestToYeelightAction_Brightness(t *testing.T) {
	method, params, err := toYeelightAction(device.Action{Method: "brightness", Params: 80})
	require.NoError(t, err)
	assert.Equal(t, "set_bright", method)
	require.Len(t, params, 3)
	assert.Equal(t, 80, params[0])
	assert.Equal(t, "smooth", params[1])
	assert.Equal(t, 500, params[2])
}

func TestToYeelightAction_BrightnessInvalidParams(t *testing.T) {
	_, _, err := toYeelightAction(device.Action{Method: "brightness", Params: "notanint"})
	assert.ErrorIs(t, err, core.ErrInvalidArguments)
}

func TestToYeelightAction_UnknownMethod(t *testing.T) {
	_, _, err := toYeelightAction(device.Action{Method: "flash"})
	assert.ErrorIs(t, err, core.ErrDeviceNotCapable)
}
