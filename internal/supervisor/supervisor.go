package supervisor

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type Supervisor struct {
	services          []Service
	messageBusManager *MessageBusManager
}

func NewSupervisor(services ...Service) *Supervisor {
	return &Supervisor{
		services:          services,
		messageBusManager: NewMessageBusManager(),
	}
}

func (s *Supervisor) initServices() error {
	for _, svc := range s.services {
		log.Printf("initializing %s", svc.Name())
		if err := svc.Init(s.messageBusManager); err != nil {
			return fmt.Errorf("failed to init %s: %w", svc.Name(), err)
		}
	}
	return nil
}

func (s *Supervisor) startServices(ctx context.Context) error {
	var wg sync.WaitGroup
	errCh := make(chan error, len(s.services))

	for _, svc := range s.services {
		wg.Add(1)
		go func(svc Service) {
			defer wg.Done()

			go func() {
				<-ctx.Done()
				log.Printf("Context is done, stopping service")
				svc.Stop()
			}()

			log.Printf("starting %s", svc.Name())
			if err := svc.Run(); err != nil {
				errCh <- fmt.Errorf("%s: %w", svc.Name(), err)
			}
		}(svc)
	}

	wg.Wait()
	close(errCh)

	var errs []error
	for err := range errCh {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

func (s *Supervisor) Run() error {
	ctx, cancel := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM, syscall.SIGINT,
	)
	defer cancel()

	if e := s.initServices(); e != nil {
		log.Fatalf("Could not init services: %v", e)
		return e
	}
	return s.startServices(ctx)
}
