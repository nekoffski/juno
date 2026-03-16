package main

import (
	"log"
	"time"

	"github.com/nekoffski/juno/internal/supervisor"
)

type dummyService struct {
	running bool
	bus     *supervisor.MessageBus
}

func (s *dummyService) Init(mbm *supervisor.MessageBusManager) error {
	s.running = true
	var err error
	s.bus, err = mbm.RegisterBus(s.Name())
	return err
}

func (s *dummyService) Run() error {
	for {
		if !s.running {
			break
		}
		log.Printf("Running!")
		time.Sleep(time.Second)
	}
	return nil
}

func (s *dummyService) Stop()        { s.running = false }
func (s *dummyService) Name() string { return "dummyService" }

func main() {
	supervisor := supervisor.NewSupervisor(&dummyService{})
	if err := supervisor.Run(); err != nil {
		log.Fatalf("Failed to start supervisor: %v", err)
	}
}
