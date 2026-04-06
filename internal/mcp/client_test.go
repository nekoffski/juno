package mcp

import (
"context"
"encoding/json"
"net/http"
"net/http/httptest"
"testing"

"github.com/stretchr/testify/assert"
"github.com/stretchr/testify/require"
)

func fakeRestServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv
}

func TestGetDevices_OK(t *testing.T) {
	devices := []Device{{Id: 1, Name: "light", Capabilities: []string{"on", "off"}}}
	body, _ := json.Marshal(devices)
	srv := fakeRestServer(t, func(w http.ResponseWriter, r *http.Request) {
assert.Equal(t, http.MethodGet, r.Method)
assert.Equal(t, "/device", r.URL.Path)
w.Header().Set("Content-Type", "application/json")
		w.Write(body)
	})
	client := NewHTTPClient(srv.URL)
	got, err := client.GetDevices(context.Background())
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, 1, got[0].Id)
	assert.Equal(t, "light", got[0].Name)
}

func TestGetDevices_Empty(t *testing.T) {
	srv := fakeRestServer(t, func(w http.ResponseWriter, r *http.Request) {
w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
	})
	client := NewHTTPClient(srv.URL)
	got, err := client.GetDevices(context.Background())
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestGetDevices_ServerError(t *testing.T) {
	srv := fakeRestServer(t, func(w http.ResponseWriter, r *http.Request) {
w.WriteHeader(http.StatusInternalServerError)
})
	client := NewHTTPClient(srv.URL)
	_, err := client.GetDevices(context.Background())
	require.Error(t, err)
}

func TestGetDevice_Found(t *testing.T) {
	device := Device{Id: 42, Name: "bulb", Capabilities: []string{"toggle"}}
	body, _ := json.Marshal(device)
	srv := fakeRestServer(t, func(w http.ResponseWriter, r *http.Request) {
assert.Equal(t, http.MethodGet, r.Method)
assert.Equal(t, "/device/id/42", r.URL.Path)
w.Header().Set("Content-Type", "application/json")
		w.Write(body)
	})
	client := NewHTTPClient(srv.URL)
	got, err := client.GetDevice(context.Background(), 42)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, 42, got.Id)
	assert.Equal(t, "bulb", got.Name)
}

func TestGetDevice_NotFound(t *testing.T) {
	srv := fakeRestServer(t, func(w http.ResponseWriter, r *http.Request) {
w.WriteHeader(http.StatusNotFound)
})
	client := NewHTTPClient(srv.URL)
	_, err := client.GetDevice(context.Background(), 99)
	require.ErrorIs(t, err, ErrDeviceNotFound)
}

func TestGetDevice_ServerError(t *testing.T) {
	srv := fakeRestServer(t, func(w http.ResponseWriter, r *http.Request) {
w.WriteHeader(http.StatusInternalServerError)
})
	client := NewHTTPClient(srv.URL)
	_, err := client.GetDevice(context.Background(), 1)
	require.Error(t, err)
	assert.NotErrorIs(t, err, ErrDeviceNotFound)
}

func TestPerformAction_OK(t *testing.T) {
	srv := fakeRestServer(t, func(w http.ResponseWriter, r *http.Request) {
assert.Equal(t, http.MethodPost, r.Method)
assert.Equal(t, "/device/id/1/action/toggle", r.URL.Path)
w.WriteHeader(http.StatusOK)
})
	client := NewHTTPClient(srv.URL)
	err := client.PerformAction(context.Background(), 1, "toggle", nil)
	require.NoError(t, err)
}

func TestPerformAction_WithParams(t *testing.T) {
	srv := fakeRestServer(t, func(w http.ResponseWriter, r *http.Request) {
assert.Equal(t, "/device/id/5/action/brightness", r.URL.Path)
var body actionRequest
require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, float64(80), body.Params["brightness"])
		w.WriteHeader(http.StatusOK)
	})
	client := NewHTTPClient(srv.URL)
	err := client.PerformAction(context.Background(), 5, "brightness", map[string]any{"brightness": float64(80)})
	require.NoError(t, err)
}

func TestPerformAction_NotFound(t *testing.T) {
	srv := fakeRestServer(t, func(w http.ResponseWriter, r *http.Request) {
w.WriteHeader(http.StatusNotFound)
})
	client := NewHTTPClient(srv.URL)
	err := client.PerformAction(context.Background(), 99, "toggle", nil)
	require.ErrorIs(t, err, ErrDeviceNotFound)
}

func TestPerformAction_BadRequest(t *testing.T) {
	srv := fakeRestServer(t, func(w http.ResponseWriter, r *http.Request) {
w.WriteHeader(http.StatusBadRequest)
})
	client := NewHTTPClient(srv.URL)
	err := client.PerformAction(context.Background(), 1, "unknown", nil)
	require.Error(t, err)
	assert.NotErrorIs(t, err, ErrDeviceNotFound)
}
