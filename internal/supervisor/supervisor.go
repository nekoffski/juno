package supervisor

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/nekoffski/juno/internal/bus"
	"github.com/rs/zerolog/log"
)

type Supervisor struct {
	services   []Service
	messageBus *bus.MessageBus
}

func NewSupervisor(services ...Service) *Supervisor {
	return &Supervisor{
		services:   services,
		messageBus: bus.New(),
	}
}

func (s *Supervisor) initServices(ctx context.Context) error {
	for _, svc := range s.services {
		log.Info().Str("service_name", svc.Name()).Msg("initializing")
		if err := svc.Init(ctx, s.messageBus); err != nil {
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
			log.Info().Str("service_name", svc.Name()).Msg("starting")
			if err := svc.Run(ctx); err != nil {
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

	if e := s.initServices(ctx); e != nil {
		log.Fatal().Err(e).Msg("could not init services")
		return e
	}
	return s.startServices(ctx)
}
