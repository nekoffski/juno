package bus

import "errors"

var (
	ErrBusNotFound = errors.New("bus not found")
	ErrQueueFull   = errors.New("queue full")
	ErrEmptyName   = errors.New("bus name must not be empty")
)
