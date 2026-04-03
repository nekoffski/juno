package web

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newEchoCtx(method string, form url.Values) echo.Context {
	e := echo.New()
	var body string
	if form != nil {
		body = form.Encode()
	}
	req := httptest.NewRequest(method, "/", strings.NewReader(body))
	if form != nil {
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	}
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec)
}

func TestHexToRGB(t *testing.T) {
	tests := []struct {
		input   string
		r, g, b int
	}{
		{"#FF8040", 255, 128, 64},
		{"FF8040", 255, 128, 64},
		{"#000000", 0, 0, 0},
		{"#ffffff", 255, 255, 255},
		{"", 0, 0, 0},
		{"#12345", 0, 0, 0},
		{"#ZZZZZZ", 0, 0, 0},
	}
	for _, tc := range tests {
		r, g, b := hexToRGB(tc.input)
		assert.Equal(t, tc.r, r, "input %q R", tc.input)
		assert.Equal(t, tc.g, g, "input %q G", tc.input)
		assert.Equal(t, tc.b, b, "input %q B", tc.input)
	}
}

func TestBuildActionBody_Toggle(t *testing.T) {
	c := newEchoCtx(http.MethodPost, nil)
	body, err := buildActionBody("toggle", c)
	require.NoError(t, err)
	assert.JSONEq(t, `{}`, string(body))
}

func TestBuildActionBody_Brightness_OK(t *testing.T) {
	c := newEchoCtx(http.MethodPost, url.Values{"brightness": {"75"}})
	body, err := buildActionBody("brightness", c)
	require.NoError(t, err)
	assert.JSONEq(t, `{"params":{"brightness":75}}`, string(body))
}

func TestBuildActionBody_Brightness_Invalid(t *testing.T) {
	c := newEchoCtx(http.MethodPost, url.Values{"brightness": {"notanumber"}})
	_, err := buildActionBody("brightness", c)
	assert.Error(t, err)
}

func TestBuildActionBody_RGB(t *testing.T) {
	c := newEchoCtx(http.MethodPost, url.Values{"color": {"#FF8040"}})
	body, err := buildActionBody("rgb", c)
	require.NoError(t, err)
	assert.JSONEq(t, `{"params":{"color":{"r":255,"g":128,"b":64}}}`, string(body))
}

func TestBuildActionBody_Default(t *testing.T) {
	c := newEchoCtx(http.MethodPost, nil)
	body, err := buildActionBody("on", c)
	require.NoError(t, err)
	assert.JSONEq(t, `{}`, string(body))
}
