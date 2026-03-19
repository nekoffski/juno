package yeelight

import "github.com/nekoffski/juno/internal/device"

type Device struct {
	model *device.DeviceModel
}

func (d *Device) Model() *device.DeviceModel {
	return d.model
}

func connectDevice(id int, addr device.DeviceAddr, name string) (device.Device, error) {
	model := &device.DeviceModel{
		Id:     id,
		Name:   name,
		Vendor: device.DeviceVendorYeelight,
		Status: device.DeviceStatusOffline,
	}
	return &Device{model: model}, nil
}
