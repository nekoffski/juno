package main

import (
	"log"
	"time"

	"github.com/nekoffski/juno/internal/supervisor"
)

type dummyService struct {
	running bool
}

func (s *dummyService) Init() error {
	s.running = true
	return nil
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
