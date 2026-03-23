package device

import (
	"context"

	"github.com/nekoffski/juno/internal/bus"
)

type VendorAdapter interface {
	Discover(ctx context.Context) ([]DeviceAddr, error)
	CreateDevice(ctx context.Context, id int, addr DeviceAddr, name string, publisher *bus.Publisher) (Device, error)
	Name() DeviceVendor
}
