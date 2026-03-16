package supervisor

import "context"

type Service interface {
	Init(*MessageBus) error
	Run(ctx context.Context) error
	Name() string
}
