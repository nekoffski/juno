package core

import "errors"

var (
	ErrDeviceNotFound = errors.New("device not found")
	ErrBusNotFound    = errors.New("bus not found")
	ErrQueueFull      = errors.New("queue full")
	ErrEmptyName      = errors.New("bus name must not be empty")
)
