package lan

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"sync"

	"github.com/rs/zerolog/log"
)

type discoverRequest struct {
	Addr       string `json:"addr"`
	Message    string `json:"message"`
	TimeoutSec int    `json:"timeout_sec"`
}

type discoverResponse struct {
	Devices []DiscoveryResult `json:"devices"`
}

func handleDiscover(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req discoverRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Addr == "" || req.Message == "" {
		http.Error(w, "addr and message are required", http.StatusBadRequest)
		return
	}

	timeout := req.TimeoutSec
	if timeout <= 0 {
		timeout = 5
	}

	devices, err := discoverDevices(req.Addr, req.Message, timeout)
	if err != nil {
		log.Error().Err(err).Msg("discovery error")
		http.Error(w, "discovery failed", http.StatusInternalServerError)
		return
	}

	if devices == nil {
		devices = []DiscoveryResult{}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(discoverResponse{Devices: devices})
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func handleConnect(w http.ResponseWriter, r *http.Request) {
	target := r.Host
	if target == "" {
		http.Error(w, "missing target host", http.StatusBadRequest)
		return
	}

	deviceConn, err := net.DialTimeout("tcp", target, 10e9)
	if err != nil {
		log.Error().Err(err).Str("target", target).Msg("CONNECT: failed to dial")
		http.Error(w, "bad gateway", http.StatusBadGateway)
		return
	}

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		deviceConn.Close()
		http.Error(w, "hijacking not supported", http.StatusInternalServerError)
		return
	}

	clientConn, bufrw, err := hijacker.Hijack()
	if err != nil {
		deviceConn.Close()
		http.Error(w, "hijack failed", http.StatusInternalServerError)
		return
	}

	if _, err := bufrw.WriteString("HTTP/1.1 200 Connection established\r\n\r\n"); err != nil {
		clientConn.Close()
		deviceConn.Close()
		return
	}
	if err := bufrw.Flush(); err != nil {
		clientConn.Close()
		deviceConn.Close()
		return
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		defer deviceConn.Close()
		_, _ = io.Copy(deviceConn, bufrw)
	}()

	go func() {
		defer wg.Done()
		defer clientConn.Close()
		_, _ = io.Copy(bufrw, deviceConn)
	}()

	wg.Wait()
}

func newMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/discover", handleDiscover)
	mux.HandleFunc("/health", handleHealth)
	return mux
}

type topHandler struct {
	mux *http.ServeMux
}

func (h *topHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		handleConnect(w, r)
		return
	}
	h.mux.ServeHTTP(w, r)
}
