package core

import "errors"

var (
	ErrDeviceNotFound   = errors.New("device not found")
	ErrInvalidArguments = errors.New("invalid arguments")
	ErrDeviceNotCapable = errors.New("device does not support this action")
	ErrDeviceBusy       = errors.New("device is busy")
	ErrBusNotFound      = errors.New("bus not found")
	ErrQueueFull        = errors.New("queue full")
	ErrEmptyName        = errors.New("bus name must not be empty")
)
