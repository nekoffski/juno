package supervisor

import (
	"context"

	"github.com/nekoffski/juno/internal/bus"
)

type Service interface {
	Init(ctx context.Context, bus *bus.MessageBus) error
	Run(ctx context.Context) error
	Name() string
}
