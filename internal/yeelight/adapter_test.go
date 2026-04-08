package yeelight

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nekoffski/juno/internal/device"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseResponse_NotYeelight(t *testing.T) {
	_, ok := parseResponse("192.168.1.1", "HTTP/1.1 200 OK\r\nServer: generic\r\n\r\n")
	assert.False(t, ok)
}

func TestParseResponse_YeelightDefaultPort(t *testing.T) {
	raw := "HTTP/1.1 200 OK\r\nServer: YEELIGHT 1\r\n\r\n"
	addr, ok := parseResponse("192.168.1.5", raw)
	require.True(t, ok)
	assert.Equal(t, "192.168.1.5", addr.Ip)
	assert.Equal(t, defaultPort, addr.Port)
}

func TestParseResponse_YeelightCustomPort(t *testing.T) {
	raw := "HTTP/1.1 200 OK\r\nLocation: yeelight://192.168.1.5:12345\r\nServer: YEELIGHT 1\r\n\r\n"
	addr, ok := parseResponse("192.168.1.5", raw)
	require.True(t, ok)
	assert.Equal(t, 12345, addr.Port)
}

func TestParseResponse_LocationMalformed_FallsBackToDefault(t *testing.T) {
	raw := "HTTP/1.1 200 OK\r\nLocation: yeelight\r\nServer: YEELIGHT 1\r\n\r\n"
	addr, ok := parseResponse("10.0.0.1", raw)
	require.True(t, ok)
	assert.Equal(t, defaultPort, addr.Port)
}

func TestParseResponse_CaseInsensitive(t *testing.T) {
	raw := "HTTP/1.1 200 OK\r\nlocation: YEELIGHT://1.2.3.4:60000\r\n\r\n"
	addr, ok := parseResponse("1.2.3.4", raw)
	require.True(t, ok)
	assert.Equal(t, 60000, addr.Port)
}

func makeLanAgentServer(t *testing.T, devices []lanDiscoveryResult) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/discover", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(lanDiscoverResponse{Devices: devices})
	}))
}

func TestAdapter_Discover_ViaLanAgent_Empty(t *testing.T) {
	srv := makeLanAgentServer(t, nil)
	defer srv.Close()

	a := NewAdapter("239.255.255.250:1982", srv.URL)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	addrs, err := a.Discover(ctx)
	require.NoError(t, err)
	assert.Empty(t, addrs)
}

func TestAdapter_Discover_ViaLanAgent_FiltersNonYeelight(t *testing.T) {
	srv := makeLanAgentServer(t, []lanDiscoveryResult{
		{IP: "10.0.0.1", RawResponse: "HTTP/1.1 200 OK\r\n\r\n"},
	})
	defer srv.Close()

	a := NewAdapter("239.255.255.250:1982", srv.URL)
	addrs, err := a.Discover(context.Background())
	require.NoError(t, err)
	assert.Empty(t, addrs)
}

func TestAdapter_Discover_ViaLanAgent_ParsesDevice(t *testing.T) {
	raw := "HTTP/1.1 200 OK\r\nLocation: yeelight://10.0.0.5:55555\r\nServer: YEELIGHT 1\r\n\r\n"
	srv := makeLanAgentServer(t, []lanDiscoveryResult{
		{IP: "10.0.0.5", RawResponse: raw},
	})
	defer srv.Close()

	a := NewAdapter("239.255.255.250:1982", srv.URL)
	addrs, err := a.Discover(context.Background())
	require.NoError(t, err)
	require.Len(t, addrs, 1)
	assert.Equal(t, device.DeviceAddr{Ip: "10.0.0.5", Port: 55555}, addrs[0])
}

func TestAdapter_Discover_ViaLanAgent_MultipleDevices(t *testing.T) {
	raw := func(port string) string {
		return "HTTP/1.1 200 OK\r\nLocation: yeelight://10.0.0.1:" + port + "\r\nServer: YEELIGHT 1\r\n\r\n"
	}
	srv := makeLanAgentServer(t, []lanDiscoveryResult{
		{IP: "10.0.0.1", RawResponse: raw("1001")},
		{IP: "10.0.0.2", RawResponse: raw("1002")},
		{IP: "10.0.0.3", RawResponse: "not-a-light"},
	})
	defer srv.Close()

	a := NewAdapter("239.255.255.250:1982", srv.URL)
	addrs, err := a.Discover(context.Background())
	require.NoError(t, err)
	assert.Len(t, addrs, 2)
}

func TestAdapter_Discover_ViaLanAgent_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer srv.Close()

	a := NewAdapter("239.255.255.250:1982", srv.URL)
	_, err := a.Discover(context.Background())
	assert.Error(t, err)
}

func TestAdapter_Name(t *testing.T) {
	a := NewAdapter("", "")
	assert.Equal(t, device.DeviceVendorYeelight, a.Name())
}

func TestAdapter_Discover_Direct_ContextCancelled(t *testing.T) {
	ln, err := net.ListenPacket("udp4", "127.0.0.1:0")
	require.NoError(t, err)
	defer ln.Close()

	a := NewAdapter(ln.LocalAddr().String(), "")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	addrs, err := a.Discover(ctx)
	assert.Nil(t, err)
	assert.Empty(t, addrs)
}
