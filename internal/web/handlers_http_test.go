package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func parseTemplates(t *testing.T) *template.Template {
	t.Helper()
	tmpl, err := template.New("").Funcs(template.FuncMap{
		"rgbHex":  func(v interface{}) string { return "#000000" },
		"propInt": func(v interface{}) int { return 0 },
	}).ParseFS(TemplateFS, "templates/*.html")
	require.NoError(t, err)
	return tmpl
}

func fakeRestServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv
}

func TestDashboard(t *testing.T) {
	srv := fakeRestServer(t, func(w http.ResponseWriter, r *http.Request) {})
	h := NewHandlers(srv.URL, parseTemplates(t))
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	require.NoError(t, h.Dashboard(c))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Type"), "text/html")
}

func TestDevicesTab_OK(t *testing.T) {
	devices := []map[string]any{{
		"id": float64(1), "name": "light",
		"vendor": "Yeelight", "status": "online",
		"capabilities": []string{"on"},
		"properties":   map[string]any{},
	}}
	body, _ := json.Marshal(devices)
	srv := fakeRestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(body)
	})
	h := NewHandlers(srv.URL, parseTemplates(t))
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/tabs/devices", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	require.NoError(t, h.DevicesTab(c))
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestMetricsTab(t *testing.T) {
	srv := fakeRestServer(t, func(w http.ResponseWriter, r *http.Request) {})
	h := NewHandlers(srv.URL, parseTemplates(t))
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/tabs/metrics", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	require.NoError(t, h.MetricsTab(c))
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestEventsTab(t *testing.T) {
	srv := fakeRestServer(t, func(w http.ResponseWriter, r *http.Request) {})
	h := NewHandlers(srv.URL, parseTemplates(t))
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/tabs/events", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	require.NoError(t, h.EventsTab(c))
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestPerformAction_OK(t *testing.T) {
	device := map[string]any{
		"id": float64(1), "name": "light",
		"vendor": "Yeelight", "status": "online",
		"capabilities": []string{"on"},
		"properties":   map[string]any{},
	}
	devBody, _ := json.Marshal(device)
	srv := fakeRestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(devBody)
	})
	h := NewHandlers(srv.URL, parseTemplates(t))
	e := echo.New()
	form := url.Values{}
	req := httptest.NewRequest(http.MethodPost, "/device/1/action/toggle", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "action")
	c.SetParamValues("1", "toggle")
	require.NoError(t, h.PerformAction(c))
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestPerformAction_InvalidBrightness(t *testing.T) {
	srv := fakeRestServer(t, func(w http.ResponseWriter, r *http.Request) {})
	h := NewHandlers(srv.URL, parseTemplates(t))
	e := echo.New()
	form := url.Values{"brightness": {"notanumber"}}
	req := httptest.NewRequest(http.MethodPost, "/device/1/action/brightness", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "action")
	c.SetParamValues("1", "brightness")
	err := h.PerformAction(c)
	var he *echo.HTTPError
	require.ErrorAs(t, err, &he)
	assert.Equal(t, http.StatusBadRequest, he.Code)
}

func TestHandlers_DevicesTab_ServerError(t *testing.T) {
	srv := fakeRestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "oops")
	})
	h := NewHandlers(srv.URL, parseTemplates(t))
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/tabs/devices", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.DevicesTab(c)
	assert.Error(t, err)
}
