package supervisor

type Service interface {
	Init() error
	Stop()
	Run() error
	Name() string
}
