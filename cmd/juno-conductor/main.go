package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/nekoffski/juno/internal/logger"
	"github.com/rs/zerolog/log"
)

type ProcessDef struct {
	Name   string
	Binary string   `json:"binary"`
	Args   []string `json:"args"`
}

type Config struct {
	Processes []ProcessDef
}

func loadConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	defer f.Close()

	var raw struct {
		Processes map[string]struct {
			Binary string   `json:"binary"`
			Args   []string `json:"args"`
		} `json:"processes"`
	}
	if err := json.NewDecoder(f).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	if len(raw.Processes) == 0 {
		return nil, fmt.Errorf("no processes defined in config")
	}
	var cfg Config
	for name, p := range raw.Processes {
		cfg.Processes = append(cfg.Processes, ProcessDef{
			Name:   name,
			Binary: p.Binary,
			Args:   p.Args,
		})
	}
	return &cfg, nil
}

func runProcess(ctx context.Context, pd ProcessDef) {
	const (
		backoffInitial = 500 * time.Millisecond
		backoffMax     = 30 * time.Second
		backoffFactor  = 2
	)

	backoff := backoffInitial

	for {
		if ctx.Err() != nil {
			return
		}

		cmd := exec.Command(pd.Binary, pd.Args...)
		cmd.Env = os.Environ()
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setpgid:   true,
			Pdeathsig: syscall.SIGTERM,
		}

		log.Info().Str("process", pd.Name).Str("binary", pd.Binary).Strs("args", pd.Args).Msg("starting process")
		if err := cmd.Start(); err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Error().Err(err).Str("process", pd.Name).Str("backoff", backoff.String()).Msg("failed to start process")
			select {
			case <-ctx.Done():
				return
			case <-time.After(backoff):
			}
			backoff = min(backoff*backoffFactor, backoffMax)
			continue
		}

		done := make(chan error, 1)
		go func() { done <- cmd.Wait() }()

		select {
		case <-ctx.Done():
			_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
			<-done
			return
		case err := <-done:
			if ctx.Err() != nil {
				return
			}
			if err != nil {
				log.Error().Err(err).Str("process", pd.Name).Str("backoff", backoff.String()).Msg("process exited with error, restarting")
			} else {
				log.Warn().Str("process", pd.Name).Str("backoff", backoff.String()).Msg("process exited cleanly, restarting")
			}
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
		}
		backoff = min(backoff*backoffFactor, backoffMax)
	}
}

func main() {
	logger.Init("conductor")

	configPath := flag.String("config", "conf/conductor.json", "path to conductor config file")
	flag.Parse()

	cfg, err := loadConfig(*configPath)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	var wg sync.WaitGroup
	for _, pd := range cfg.Processes {
		wg.Add(1)
		go func(pd ProcessDef) {
			defer wg.Done()
			runProcess(ctx, pd)
		}(pd)
	}

	wg.Wait()
	log.Info().Msg("all processes stopped")
}
