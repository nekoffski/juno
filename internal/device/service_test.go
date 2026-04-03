package device

import (
	"context"
	"testing"
	"time"

	"github.com/nekoffski/juno/internal/bus"
	"github.com/nekoffski/juno/internal/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRepo struct {
	insertErr error
	deleteErr error
	inserted  []DeviceAddr
	deleted   []int
}

func (r *mockRepo) InsertDevice(_ context.Context, addr DeviceAddr, _ DeviceVendor) (int, string, error) {
	if r.insertErr != nil {
		return 0, "", r.insertErr
	}
	r.inserted = append(r.inserted, addr)
	return 1, "Yeelight_1", nil
}

func (r *mockRepo) DeleteDevice(_ context.Context, id int) error {
	r.deleted = append(r.deleted, id)
	return r.deleteErr
}

func (r *mockRepo) FetchDevices(_ context.Context, _ func(int, DeviceAddr, DeviceVendor, string)) error {
	return nil
}

type mockDevice struct {
	model      DeviceModel
	capable    map[string]bool
	enqueueErr error
	enqueued   []Action
	closed     bool
}

func (d *mockDevice) Model() DeviceModel { return d.model }

func (d *mockDevice) IsCapable(action string) bool { return d.capable[action] }

func (d *mockDevice) EnqueueAction(a Action) error {
	if d.enqueueErr != nil {
		return d.enqueueErr
	}
	d.enqueued = append(d.enqueued, a)
	return nil
}

func (d *mockDevice) Close() error {
	d.closed = true
	return nil
}

const testTimeout = time.Second

func initService(t *testing.T, devs ...Device) (*DeviceService, *bus.Sender) {
	t.Helper()
	mb := bus.New()
	s := NewDeviceService(&mockRepo{}, nil)
	for _, d := range devs {
		s.devices[d.Model().Addr] = d
	}
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	require.NoError(t, s.Init(ctx, mb))
	return s, mb.NewSender()
}

func mustAwait(t *testing.T, sender *bus.Sender, payload any) (any, error) {
	t.Helper()
	f, err := sender.Request("device-service", payload)
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()
	return f.Await(ctx)
}

func TestOnGetDevicesRequest_Empty(t *testing.T) {
	_, sender := initService(t)
	raw, err := mustAwait(t, sender, GetDevicesRequest{})
	require.NoError(t, err)
	res := raw.(GetDevicesResponse)
	assert.Empty(t, res.Devices)
}

func TestOnGetDevicesRequest_WithDevices(t *testing.T) {
	dev := &mockDevice{model: DeviceModel{Id: 1, Name: "light", Addr: DeviceAddr{Ip: "1.2.3.4", Port: 55443}}}
	_, sender := initService(t, dev)
	raw, err := mustAwait(t, sender, GetDevicesRequest{})
	require.NoError(t, err)
	res := raw.(GetDevicesResponse)
	require.Len(t, res.Devices, 1)
	assert.Equal(t, 1, res.Devices[0].Id)
}

func TestOnGetDeviceByIdRequest_Found(t *testing.T) {
	dev := &mockDevice{model: DeviceModel{Id: 42, Addr: DeviceAddr{Ip: "1.2.3.4", Port: 55443}}}
	_, sender := initService(t, dev)
	raw, err := mustAwait(t, sender, GetDeviceByIdRequest{Id: 42})
	require.NoError(t, err)
	res := raw.(GetDeviceByIdResponse)
	assert.Equal(t, 42, res.Device.Id)
}

func TestOnGetDeviceByIdRequest_NotFound(t *testing.T) {
	_, sender := initService(t)
	_, err := mustAwait(t, sender, GetDeviceByIdRequest{Id: 99})
	assert.ErrorIs(t, err, core.ErrDeviceNotFound)
}

func TestOnGetDevicePropertiesRequest_Found(t *testing.T) {
	dev := &mockDevice{model: DeviceModel{
		Id:         1,
		Addr:       DeviceAddr{Ip: "1.2.3.4", Port: 55443},
		Properties: map[string]any{"brightness": 80, "rgb": "FF0000"},
	}}
	_, sender := initService(t, dev)
	raw, err := mustAwait(t, sender, GetDevicePropertiesRequest{Id: 1, Properties: []string{"brightness", "missing"}})
	require.NoError(t, err)
	res := raw.(GetDevicePropertiesResponse)
	assert.Equal(t, 80, res.Properties["brightness"])
	assert.Nil(t, res.Properties["missing"])
}

func TestOnGetDevicePropertiesRequest_NotFound(t *testing.T) {
	_, sender := initService(t)
	_, err := mustAwait(t, sender, GetDevicePropertiesRequest{Id: 99})
	assert.ErrorIs(t, err, core.ErrDeviceNotFound)
}

func TestOnPerformDeviceActionRequest_NotFound(t *testing.T) {
	_, sender := initService(t)
	_, err := mustAwait(t, sender, PerformDeviceActionRequest{Id: 99, Action: "on"})
	assert.ErrorIs(t, err, core.ErrDeviceNotFound)
}

func TestOnPerformDeviceActionRequest_NotCapable(t *testing.T) {
	dev := &mockDevice{
		model:   DeviceModel{Id: 1, Addr: DeviceAddr{Ip: "1.2.3.4", Port: 55443}},
		capable: map[string]bool{},
	}
	_, sender := initService(t, dev)
	_, err := mustAwait(t, sender, PerformDeviceActionRequest{Id: 1, Action: "rgb"})
	assert.ErrorIs(t, err, core.ErrDeviceNotCapable)
}

func TestOnPerformDeviceActionRequest_InvalidParams(t *testing.T) {
	dev := &mockDevice{
		model:   DeviceModel{Id: 1, Addr: DeviceAddr{Ip: "1.2.3.4", Port: 55443}},
		capable: map[string]bool{"rgb": true},
	}
	_, sender := initService(t, dev)
	_, err := mustAwait(t, sender, PerformDeviceActionRequest{Id: 1, Action: "rgb", Params: nil})
	assert.ErrorIs(t, err, core.ErrInvalidArguments)
}

func TestOnPerformDeviceActionRequest_OK(t *testing.T) {
	dev := &mockDevice{
		model:   DeviceModel{Id: 1, Addr: DeviceAddr{Ip: "1.2.3.4", Port: 55443}},
		capable: map[string]bool{"on": true},
	}
	_, sender := initService(t, dev)
	raw, err := mustAwait(t, sender, PerformDeviceActionRequest{Id: 1, Action: "on"})
	require.NoError(t, err)
	_, ok := raw.(AckResponse)
	assert.True(t, ok)
	require.Len(t, dev.enqueued, 1)
	assert.Equal(t, "on", dev.enqueued[0].Method)
}

func TestOnPerformDeviceActionRequest_EnqueueError(t *testing.T) {
	dev := &mockDevice{
		model:      DeviceModel{Id: 1, Addr: DeviceAddr{Ip: "1.2.3.4", Port: 55443}},
		capable:    map[string]bool{"on": true},
		enqueueErr: core.ErrDeviceBusy,
	}
	_, sender := initService(t, dev)
	_, err := mustAwait(t, sender, PerformDeviceActionRequest{Id: 1, Action: "on"})
	assert.ErrorIs(t, err, core.ErrDeviceBusy)
}

func TestOnHeartbeatRequest(t *testing.T) {
	_, sender := initService(t)
	raw, err := mustAwait(t, sender, core.HeartbeatRequest{Magic: "ping"})
	require.NoError(t, err)
	res := raw.(core.HeartbeatResponse)
	assert.True(t, res.Healthy)
	assert.Equal(t, "ping", res.Magic)
}

func TestOnDeleteDeviceRequest_NotFound(t *testing.T) {
	_, sender := initService(t)
	_, err := mustAwait(t, sender, DeleteDeviceRequest{Id: 99})
	assert.ErrorIs(t, err, core.ErrDeviceNotFound)
}

func TestOnDeleteDeviceRequest_Found(t *testing.T) {
	repo := &mockRepo{}
	dev := &mockDevice{model: DeviceModel{Id: 5, Addr: DeviceAddr{Ip: "1.2.3.4", Port: 55443}}}
	mb := bus.New()
	s := NewDeviceService(repo, nil)
	s.devices[dev.model.Addr] = dev
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	require.NoError(t, s.Init(ctx, mb))
	sender := mb.NewSender()

	raw, err := mustAwait(t, sender, DeleteDeviceRequest{Id: 5})
	require.NoError(t, err)
	_, ok := raw.(AckResponse)
	assert.True(t, ok)
	assert.True(t, dev.closed)
	assert.Contains(t, repo.deleted, 5)
	s.devicesMtx.RLock()
	_, stillPresent := s.devices[dev.model.Addr]
	s.devicesMtx.RUnlock()
	assert.False(t, stillPresent)
}
