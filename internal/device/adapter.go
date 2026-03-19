package device

import (
	"context"
)

type VendorAdapter interface {
	Discover(ctx context.Context) ([]DeviceAddr, error)
	CreateDevice(id int, addr DeviceAddr, name string) (Device, error)
	Name() DeviceVendor
}
