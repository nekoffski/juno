package supervisor

type Service interface {
	Init(*MessageBusManager) error
	Stop()
	Run() error
	Name() string
}
