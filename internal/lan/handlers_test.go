package lan

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleHealth(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	handleHealth(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleDiscover_MethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/discover", nil)
	w := httptest.NewRecorder()
	handleDiscover(w, req)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestHandleDiscover_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/discover", strings.NewReader("not-json"))
	w := httptest.NewRecorder()
	handleDiscover(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleDiscover_MissingAddr(t *testing.T) {
	body, _ := json.Marshal(discoverRequest{Message: "hello", TimeoutSec: 1})
	req := httptest.NewRequest(http.MethodPost, "/discover", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handleDiscover(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleDiscover_MissingMessage(t *testing.T) {
	body, _ := json.Marshal(discoverRequest{Addr: "127.0.0.1:9999", TimeoutSec: 1})
	req := httptest.NewRequest(http.MethodPost, "/discover", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handleDiscover(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleDiscover_NoResponders(t *testing.T) {
	body, _ := json.Marshal(discoverRequest{
		Addr:       "127.0.0.1:19099",
		Message:    "ping",
		TimeoutSec: 1,
	})
	req := httptest.NewRequest(http.MethodPost, "/discover", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handleDiscover(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var resp discoverResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, 0, len(resp.Devices))
}

func TestHandleDiscover_WithResponder(t *testing.T) {
	pc, err := net.ListenPacket("udp4", "127.0.0.1:0")
	require.NoError(t, err)
	defer pc.Close()

	addr := pc.LocalAddr().String()
	const reply = "yeelight:ok"

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		buf := make([]byte, 1024)
		pc.SetReadDeadline(time.Now().Add(3 * time.Second))
		n, src, err := pc.ReadFrom(buf)
		if err != nil || n == 0 {
			return
		}
		pc.WriteTo([]byte(reply), src)
	}()

	body, _ := json.Marshal(discoverRequest{
		Addr:       addr,
		Message:    "discover",
		TimeoutSec: 2,
	})
	req := httptest.NewRequest(http.MethodPost, "/discover", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handleDiscover(w, req)
	wg.Wait()

	require.Equal(t, http.StatusOK, w.Code)
	var resp discoverResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	require.Len(t, resp.Devices, 1)
	assert.Equal(t, "127.0.0.1", resp.Devices[0].IP)
	assert.Equal(t, reply, resp.Devices[0].RawResponse)
}

func TestHandleConnect_Success(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer ln.Close()

	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		io.Copy(conn, conn)
	}()

	targetAddr := ln.Addr().String()

	srv := httptest.NewServer(&topHandler{mux: newMux()})
	defer srv.Close()

	proxyAddr := strings.TrimPrefix(srv.URL, "http://")
	conn, err := net.Dial("tcp", proxyAddr)
	require.NoError(t, err)
	defer conn.Close()

	connectLine := fmt.Sprintf("CONNECT %s HTTP/1.1\r\nHost: %s\r\n\r\n", targetAddr, targetAddr)
	_, err = conn.Write([]byte(connectLine))
	require.NoError(t, err)

	buf := make([]byte, 512)
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, err := conn.Read(buf)
	require.NoError(t, err)
	assert.Contains(t, string(buf[:n]), "200")

	conn.SetDeadline(time.Now().Add(2 * time.Second))
	_, err = conn.Write([]byte("hello"))
	require.NoError(t, err)

	n, err = conn.Read(buf)
	require.NoError(t, err)
	assert.Equal(t, "hello", string(buf[:n]))
}

func TestHandleConnect_BadTarget(t *testing.T) {
	srv := httptest.NewServer(&topHandler{mux: newMux()})
	defer srv.Close()

	proxyAddr := strings.TrimPrefix(srv.URL, "http://")
	conn, err := net.Dial("tcp", proxyAddr)
	require.NoError(t, err)
	defer conn.Close()

	_, err = conn.Write([]byte("CONNECT 127.0.0.1:1 HTTP/1.1\r\nHost: 127.0.0.1:1\r\n\r\n"))
	require.NoError(t, err)

	buf := make([]byte, 512)
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, _ := conn.Read(buf)
	assert.Contains(t, string(buf[:n]), "502")
}
