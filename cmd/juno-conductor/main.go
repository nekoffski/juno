package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"encoding/json"
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

type prefixWriter struct {
	prefix string
	dst    io.Writer
	buf    []byte
}

func newPrefixWriter(prefix string, dst io.Writer) *prefixWriter {
	return &prefixWriter{prefix: prefix, dst: dst}
}

func (w *prefixWriter) Write(p []byte) (int, error) {
	w.buf = append(w.buf, p...)
	for {
		idx := -1
		for i, b := range w.buf {
			if b == '\n' {
				idx = i
				break
			}
		}
		if idx < 0 {
			break
		}
		line := w.buf[:idx+1]
		if _, err := fmt.Fprintf(w.dst, "[%s] %s", w.prefix, line); err != nil {
			return 0, err
		}
		w.buf = append([]byte{}, w.buf[idx+1:]...)
	}
	return len(p), nil
}

func runProcess(ctx context.Context, pd ProcessDef) {
	const (
		backoffInitial = 500 * time.Millisecond
		backoffMax     = 30 * time.Second
		backoffFactor  = 2
	)

	stdout := newPrefixWriter(pd.Name, os.Stdout)
	stderr := newPrefixWriter(pd.Name, os.Stderr)
	backoff := backoffInitial

	for {
		if ctx.Err() != nil {
			return
		}

		cmd := exec.Command(pd.Binary, pd.Args...)
		cmd.Env = os.Environ()
		cmd.Stdout = stdout
		cmd.Stderr = stderr

		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setpgid:   true,
			Pdeathsig: syscall.SIGTERM,
		}

		log.Printf("[conductor] starting %s (%s %v)", pd.Name, pd.Binary, pd.Args)
		if err := cmd.Start(); err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("[conductor] failed to start %s: %v; retrying in %s", pd.Name, err, backoff)
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
				log.Printf("[conductor] %s exited with error: %v; restarting in %s", pd.Name, err, backoff)
			} else {
				log.Printf("[conductor] %s exited cleanly; restarting in %s", pd.Name, backoff)
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
	configPath := flag.String("config", "conf/conductor.json", "path to conductor config file")
	flag.Parse()

	cfg, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("[conductor] failed to load config: %v", err)
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
	log.Println("[conductor] all processes stopped")
}
