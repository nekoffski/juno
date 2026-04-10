package yeelight

import (
	"context"
	"sync"
	"time"

	"github.com/nekoffski/juno/internal/bus"
	"github.com/nekoffski/juno/internal/core"
	"github.com/nekoffski/juno/internal/device"
	"github.com/rs/zerolog/log"
)

func getYeelightCapabilities() []string {
	return []string{"on", "off", "toggle", "brightness", "ct", "rgb"}
}

type Device struct {
	model        *device.DeviceModel
	client       *client
	publisher    *bus.Publisher
	modelMtx     sync.RWMutex
	actionsQueue chan device.Action
	done         chan struct{}
}

func (d *Device) Model() device.DeviceModel {
	d.modelMtx.RLock()
	defer d.modelMtx.RUnlock()
	return *d.model
}

func (d *Device) EnqueueAction(action device.Action) error {
	select {
	case d.actionsQueue <- action:
		log.Info().Int("id", d.model.Id).Str("action", action.Method).Interface("params", action.Params).Msg("enqueued action")
		return nil
	default:
		log.Warn().Int("id", d.model.Id).Str("action", action.Method).Msg("device busy, cannot enqueue action")
		return core.ErrDeviceBusy
	}
}

func (d *Device) IsCapable(action string) bool {
	for _, cap := range d.model.Capabilities {
		if cap == action {
			return true
		}
	}
	return false
}

func (d *Device) Close() error {
	close(d.done)
	return d.client.close()
}

func (d *Device) writerLoop(ctx context.Context) {
	for {
		select {
		case action := <-d.actionsQueue:
			log.Info().Int("id", d.model.Id).Str("action", action.Method).Interface("params", action.Params).Msg("processing action")

			method, params, err := toYeelightAction(action)
			if err != nil {
				log.Error().Err(err).Int("id", d.model.Id).Msg("failed to convert action")
				continue
			}
			pr, err := d.client.sendRequest(ctx, method, params)
			if err != nil {
				log.Error().Err(err).Int("id", d.model.Id).Msg("failed to send action")
			}
			_, err = waitForResponse(ctx, pr)
			if err != nil {
				log.Error().Err(err).Int("id", d.model.Id).Msg("failed to get response for action")
				continue
			}
			_ = d.publisher.Publish("device.events", device.DeviceUpdatedEvent{Device: d.Model()})
		case <-ctx.Done():
			log.Info().Int("id", d.model.Id).Msg("writer loop exiting")
			return
		case <-d.done:
			log.Info().Int("id", d.model.Id).Msg("writer loop received done signal")
			return
		}
	}
}

func (d *Device) onNotification(n notification) {
	log.Debug().Int("id", d.model.Id).Interface("params", n.Params).Msg("received notification")

	for k, v := range n.Params {
		nk, nv := mapProperty(k, v)

		d.modelMtx.Lock()
		d.model.Properties[nk] = nv

		if nk == "power" {
			d.model.Status = device.DeviceStatusOnline
			if v == "off" {
				d.model.Status = device.DeviceStatusOffline
			}
		}
		d.modelMtx.Unlock()
	}

	_ = d.publisher.Publish("device.events", device.DeviceUpdatedEvent{Device: d.Model()})
}

func createDevice(ctx context.Context, id int, addr device.DeviceAddr, name string, publisher *bus.Publisher, lanAgentURL string) (device.Device, error) {
	c, err := newClient(ctx, addr, lanAgentURL)
	if err != nil {
		return nil, err
	}

	propsQuery := []string{"power", "bright", "rgb", "ct"}
	readCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	response, err := c.readProperties(readCtx, propsQuery)
	if err != nil {
		c.close()
		return nil, err
	}

	status := device.DeviceStatusOffline
	if response["power"] == "on" {
		status = device.DeviceStatusOnline
	}
	delete(response, "power")

	props := make(map[string]any)
	for k, v := range response {
		nk, nv := mapProperty(k, v)
		props[nk] = nv
	}

	model := &device.DeviceModel{
		Id:           id,
		Name:         name,
		Vendor:       device.DeviceVendorYeelight,
		Status:       status,
		Addr:         addr,
		Capabilities: getYeelightCapabilities(),
		Properties:   props,
	}
	d := &Device{model: model, client: c, publisher: publisher, actionsQueue: make(chan device.Action, 10), done: make(chan struct{})}

	c.setNotificationCallback(d.onNotification)
	go d.writerLoop(ctx)
	return d, nil
}
