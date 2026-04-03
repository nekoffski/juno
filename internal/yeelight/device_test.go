package yeelight

import (
	"testing"

	"github.com/nekoffski/juno/internal/core"
	"github.com/nekoffski/juno/internal/device"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestDevice(caps []string) *Device {
	return &Device{
		model: &device.DeviceModel{
			Id:           1,
			Name:         "test",
			Capabilities: caps,
			Properties:   map[string]any{},
		},
		actionsQueue: make(chan device.Action, 10),
		done:         make(chan struct{}),
	}
}

func TestDevice_IsCapable(t *testing.T) {
	d := newTestDevice(getYeelightCapabilities())
	for _, cap := range getYeelightCapabilities() {
		assert.True(t, d.IsCapable(cap), "expected capable: %s", cap)
	}
	assert.False(t, d.IsCapable("flash"))
	assert.False(t, d.IsCapable(""))
}

func TestDevice_Model_ReturnsCopy(t *testing.T) {
	d := newTestDevice(nil)
	m := d.Model()
	m.Name = "mutated"
	assert.Equal(t, "test", d.Model().Name)
}

func TestDevice_EnqueueAction_OK(t *testing.T) {
	d := newTestDevice(nil)
	err := d.EnqueueAction(device.Action{Method: "on"})
	require.NoError(t, err)
}

func TestDevice_EnqueueAction_QueueFull(t *testing.T) {
	d := newTestDevice(nil)
	for i := 0; i < 10; i++ {
		require.NoError(t, d.EnqueueAction(device.Action{Method: "on"}))
	}
	err := d.EnqueueAction(device.Action{Method: "on"})
	assert.ErrorIs(t, err, core.ErrDeviceBusy)
}
