package main

import (
	"context"
	"log"
	"time"

	"github.com/nekoffski/juno/internal/supervisor"
)

type dummyService struct {
	sender *supervisor.Sender
}

func (s *dummyService) Init(messageBus *supervisor.MessageBus) error {
	s.sender = messageBus.NewSender()
	return nil
}

func (s *dummyService) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(time.Second):
			log.Printf("Running! %v", s.Name())
		}
	}
}

func (s *dummyService) Name() string { return "dummyService" }

func main() {
	supervisor := supervisor.NewSupervisor(&dummyService{})
	if err := supervisor.Run(); err != nil {
		log.Fatalf("Failed to start supervisor: %v", err)
	}
}
