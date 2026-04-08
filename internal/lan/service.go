package lan

import (
	"context"
	"fmt"
	"net/http"
)

type Config struct {
	Addr string
	Port int
}

type Service struct {
	server *http.Server
}

func NewService(cfg Config) *Service {
	handler := &topHandler{mux: newMux()}
	return &Service{
		server: &http.Server{
			Addr:    fmt.Sprintf("%s:%d", cfg.Addr, cfg.Port),
			Handler: handler,
		},
	}
}

func (s *Service) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return s.server.Shutdown(context.Background())
	}
}
