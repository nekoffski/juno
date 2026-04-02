package device

import (
"testing"

"github.com/nekoffski/juno/internal/core"
"github.com/stretchr/testify/assert"
"github.com/stretchr/testify/require"
)

func TestParseActionParams_On(t *testing.T) {
	params, err := parseActionParams("on", nil)
	require.NoError(t, err)
	assert.Nil(t, params)
}

func TestParseActionParams_Off(t *testing.T) {
	params, err := parseActionParams("off", nil)
	require.NoError(t, err)
	assert.Nil(t, params)
}

func TestParseActionParams_Toggle(t *testing.T) {
	params, err := parseActionParams("toggle", nil)
	require.NoError(t, err)
	assert.Nil(t, params)
}

func TestParseActionParams_Rgb_OK(t *testing.T) {
	payload := map[string]any{
		"color": map[string]any{
			"r": float64(255),
			"g": float64(128),
			"b": float64(64),
		},
	}
	params, err := parseActionParams("rgb", payload)
	require.NoError(t, err)
	assert.Equal(t, ColorRGB{R: 255, G: 128, B: 64}, params)
}

func TestParseActionParams_Rgb_MissingColorKey(t *testing.T) {
	payload := map[string]any{}
	_, err := parseActionParams("rgb", payload)
	assert.ErrorIs(t, err, core.ErrInvalidArguments)
}

func TestParseActionParams_Rgb_MissingChannel(t *testing.T) {
	payload := map[string]any{
		"color": map[string]any{
			"r": float64(255),
		},
	}
	_, err := parseActionParams("rgb", payload)
	assert.ErrorIs(t, err, core.ErrInvalidArguments)
}

func TestParseActionParams_Brightness_OK(t *testing.T) {
	payload := map[string]any{
		"brightness": float64(80),
	}
	params, err := parseActionParams("brightness", payload)
	require.NoError(t, err)
	assert.Equal(t, 80, params)
}

func TestParseActionParams_Brightness_Missing(t *testing.T) {
	payload := map[string]any{}
	_, err := parseActionParams("brightness", payload)
	assert.ErrorIs(t, err, core.ErrInvalidArguments)
}

func TestParseActionParams_Brightness_WrongType(t *testing.T) {
	payload := map[string]any{
		"brightness": "notanumber",
	}
	_, err := parseActionParams("brightness", payload)
	assert.ErrorIs(t, err, core.ErrInvalidArguments)
}

func TestParseActionParams_UnknownAction(t *testing.T) {
	_, err := parseActionParams("flash", nil)
	assert.ErrorIs(t, err, core.ErrInvalidArguments)
}

func TestParseActionParams_NilPayloadForRgb(t *testing.T) {
	_, err := parseActionParams("rgb", nil)
	assert.ErrorIs(t, err, core.ErrInvalidArguments)
}
