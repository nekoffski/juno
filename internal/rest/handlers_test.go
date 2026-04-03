package rest

import (
	"context"
	"testing"
	"time"

	"github.com/nekoffski/juno/internal/bus"
	"github.com/nekoffski/juno/internal/core"
	"github.com/nekoffski/juno/internal/device"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupHandlers(t *testing.T, handler func(bus.Message)) (*DeviceHandlers, *HealthHandlers) {
	t.Helper()
	mb := bus.New()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	require.NoError(t, mb.RegisterReceiver(ctx, "device-service", handler))
	return &DeviceHandlers{sender: mb.NewSender()}, &HealthHandlers{sender: mb.NewSender()}
}

func awaitCtx() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	return ctx
}

func TestGetDevices_OK(t *testing.T) {
	devs, _ := setupHandlers(t, func(msg bus.Message) {
		msg.Reply(bus.Response{Payload: device.GetDevicesResponse{
			Devices: []*device.DeviceModel{
				{Id: 1, Name: "light", Capabilities: []string{"on"}},
			},
		}})
	})
	resp, err := devs.GetDevices(awaitCtx(), GetDevicesRequestObject{})
	require.NoError(t, err)
	res := resp.(GetDevices200JSONResponse)
	require.Len(t, res, 1)
	assert.Equal(t, 1, res[0].Id)
}

func TestGetDeviceById_Found(t *testing.T) {
	devs, _ := setupHandlers(t, func(msg bus.Message) {
		msg.Reply(bus.Response{Payload: device.GetDeviceByIdResponse{
			Device: &device.DeviceModel{Id: 42, Name: "lamp"},
		}})
	})
	resp, err := devs.GetDeviceById(awaitCtx(), GetDeviceByIdRequestObject{Id: 42})
	require.NoError(t, err)
	res := resp.(GetDeviceById200JSONResponse)
	assert.Equal(t, 42, res.Id)
}

func TestGetDeviceById_NotFound(t *testing.T) {
	devs, _ := setupHandlers(t, func(msg bus.Message) {
		msg.Reply(bus.Response{Err: core.ErrDeviceNotFound})
	})
	resp, err := devs.GetDeviceById(awaitCtx(), GetDeviceByIdRequestObject{Id: 99})
	require.NoError(t, err)
	_, ok := resp.(GetDeviceById404Response)
	assert.True(t, ok)
}

func TestDiscoverDevices_OK(t *testing.T) {
	devs, _ := setupHandlers(t, func(msg bus.Message) {})
	resp, err := devs.DiscoverDevices(awaitCtx(), DiscoverDevicesRequestObject{})
	require.NoError(t, err)
	_, ok := resp.(DiscoverDevices202Response)
	assert.True(t, ok)
}

func TestGetDeviceProperties_Found(t *testing.T) {
	fields := []string{"brightness"}
	devs, _ := setupHandlers(t, func(msg bus.Message) {
		msg.Reply(bus.Response{Payload: device.GetDevicePropertiesResponse{
			Properties: map[string]any{"brightness": 80},
		}})
	})
	resp, err := devs.GetDeviceProperties(awaitCtx(), GetDevicePropertiesRequestObject{
		Id:     1,
		Params: GetDevicePropertiesParams{Fields: &fields},
	})
	require.NoError(t, err)
	res := resp.(GetDeviceProperties200JSONResponse)
	assert.Equal(t, 80, res["brightness"])
}

func TestGetDeviceProperties_NotFound(t *testing.T) {
	fields := []string{"brightness"}
	devs, _ := setupHandlers(t, func(msg bus.Message) {
		msg.Reply(bus.Response{Err: core.ErrDeviceNotFound})
	})
	resp, err := devs.GetDeviceProperties(awaitCtx(), GetDevicePropertiesRequestObject{
		Id:     99,
		Params: GetDevicePropertiesParams{Fields: &fields},
	})
	require.NoError(t, err)
	_, ok := resp.(GetDeviceProperties404Response)
	assert.True(t, ok)
}

func TestDeleteDevice_OK(t *testing.T) {
	devs, _ := setupHandlers(t, func(msg bus.Message) {
		msg.Reply(bus.Response{Payload: device.AckResponse{}})
	})
	resp, err := devs.DeleteDevice(awaitCtx(), DeleteDeviceRequestObject{Id: 1})
	require.NoError(t, err)
	_, ok := resp.(DeleteDevice200Response)
	assert.True(t, ok)
}

func TestDeleteDevice_NotFound(t *testing.T) {
	devs, _ := setupHandlers(t, func(msg bus.Message) {
		msg.Reply(bus.Response{Err: core.ErrDeviceNotFound})
	})
	resp, err := devs.DeleteDevice(awaitCtx(), DeleteDeviceRequestObject{Id: 99})
	require.NoError(t, err)
	_, ok := resp.(DeleteDevice404Response)
	assert.True(t, ok)
}

func TestPerformDeviceAction_OK(t *testing.T) {
	devs, _ := setupHandlers(t, func(msg bus.Message) {
		msg.Reply(bus.Response{Payload: device.AckResponse{}})
	})
	resp, err := devs.PerformDeviceAction(awaitCtx(), PerformDeviceActionRequestObject{
		Id:     1,
		Action: "on",
		Body:   &PerformDeviceActionJSONRequestBody{},
	})
	require.NoError(t, err)
	_, ok := resp.(PerformDeviceAction200Response)
	assert.True(t, ok)
}

func TestPerformDeviceAction_NotFound(t *testing.T) {
	devs, _ := setupHandlers(t, func(msg bus.Message) {
		msg.Reply(bus.Response{Err: core.ErrDeviceNotFound})
	})
	resp, err := devs.PerformDeviceAction(awaitCtx(), PerformDeviceActionRequestObject{
		Id:     99,
		Action: "on",
		Body:   &PerformDeviceActionJSONRequestBody{},
	})
	require.NoError(t, err)
	_, ok := resp.(PerformDeviceAction404Response)
	assert.True(t, ok)
}

func TestGetHealth_OK(t *testing.T) {
	_, health := setupHandlers(t, func(msg bus.Message) {})
	resp, err := health.GetHealth(awaitCtx(), GetHealthRequestObject{})
	require.NoError(t, err)
	res := resp.(GetHealth200JSONResponse)
	assert.Equal(t, "ok", res.Status)
}

func TestGetDeviceServiceHealth_OK(t *testing.T) {
	_, health := setupHandlers(t, func(msg bus.Message) {
		msg.Reply(bus.Response{Payload: core.HeartbeatResponse{Healthy: true, Magic: "ping"}})
	})
	resp, err := health.GetDeviceServiceHealth(awaitCtx(), GetDeviceServiceHealthRequestObject{})
	require.NoError(t, err)
	res := resp.(GetDeviceServiceHealth200JSONResponse)
	assert.Equal(t, "ok", res.Status)
}

func TestGetDeviceServiceHealth_Unhealthy(t *testing.T) {
	_, health := setupHandlers(t, func(msg bus.Message) {
		msg.Reply(bus.Response{Payload: core.HeartbeatResponse{Healthy: false, Magic: "ping"}})
	})
	resp, err := health.GetDeviceServiceHealth(awaitCtx(), GetDeviceServiceHealthRequestObject{})
	require.NoError(t, err)
	res := resp.(GetDeviceServiceHealth200JSONResponse)
	assert.Equal(t, "unhealthy", res.Status)
}
