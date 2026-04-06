package mcp

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/caarlos0/env/v11"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

type Config struct {
	Port        int    `env:"JUNO_MCP_PORT"      envDefault:"6003"`
	RestBaseURL string `env:"JUNO_REST_BASE_URL" envDefault:"http://localhost:6000"`
}

func LoadConfig() (Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return Config{}, fmt.Errorf("failed to load config: %w", err)
	}
	return cfg, nil
}

func Start(ctx context.Context, cfg Config) error {
	client := NewHTTPClient(cfg.RestBaseURL)

	srv := sdkmcp.NewServer(&sdkmcp.Implementation{
		Name:    "juno-mcp",
		Version: "1.0.0",
	}, nil)

	RegisterTools(srv, client)

	handler := sdkmcp.NewStreamableHTTPHandler(func(_ *http.Request) *sdkmcp.Server {
		return srv
	}, nil)

	addr := fmt.Sprintf(":%d", cfg.Port)
	httpSrv := &http.Server{Addr: addr, Handler: handler}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("MCP server listening on %s", addr)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		return httpSrv.Shutdown(context.Background())
	case err := <-errCh:
		return err
	}
}
