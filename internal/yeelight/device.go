package yeelight

import (
	"context"
	"log"
	"time"

	"github.com/nekoffski/juno/internal/device"
)

func getYeelightCapabilities() []string {
	return []string{"on", "off", "toggle", "setRgb", "getRgb"}
}

type Device struct {
	model  *device.DeviceModel
	client *client
}

func (d *Device) Model() *device.DeviceModel {
	return d.model
}

func (d *Device) Close() error {
	return d.client.close()
}

func createDevice(ctx context.Context, id int, addr device.DeviceAddr, name string) (device.Device, error) {
	c, err := newClient(ctx, addr, nil)
	if err != nil {
		return nil, err
	}

	props := []string{"power", "bright", "color_mode", "ct", "rgb", "hue", "sat"}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	response, err := c.readProperties(ctx, props)
	if err != nil {
		c.close()
		return nil, err
	}

	for _, prop := range props {
		log.Printf("Device %d: %s = %s", id, prop, response[prop])
	}

	status := device.DeviceStatusOffline
	if response["power"] == "on" {
		status = device.DeviceStatusOnline
	}

	model := &device.DeviceModel{
		Id:           id,
		Name:         name,
		Vendor:       device.DeviceVendorYeelight,
		Status:       status,
		Addr:         addr,
		Capabilities: getYeelightCapabilities(),
	}
	return &Device{model: model, client: c}, nil
}
