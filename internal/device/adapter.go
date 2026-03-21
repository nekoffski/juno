package device

import (
	"context"
)

type VendorAdapter interface {
	Discover(ctx context.Context) ([]DeviceAddr, error)
	CreateDevice(ctx context.Context, id int, addr DeviceAddr, name string) (Device, error)
	Name() DeviceVendor
}
