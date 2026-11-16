package xecho_test

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JSainsburyPLC/xecho"
	"github.com/labstack/echo/v4"
	"github.com/steinfletcher/apitest"
	"github.com/stretchr/testify/assert"
)

func TestHealthCheck(t *testing.T) {
	apitest.New().
		Handler(xecho.Echo(config(""))).
		Get("/health").
		Expect(t).
		Status(http.StatusOK).
		Body(`{"status": "ok"}`).
		End()
	apitest.New().
		Handler(xecho.New(config("test")).Echo).
		Get("/test/health").
		Expect(t).
		Status(http.StatusOK).
		Body(`{"status": "ok"}`).
		End()
}

func TestIPExtractor(t *testing.T) {
	t.Parallel()

	_, proxyNet, _ := net.ParseCIDR("192.0.2.0/24") // httptest's RemoteAddr is 192.0.2.1
	tests := map[string]struct {
		extractor echo.IPExtractor
		want      string
	}{
		"default ignores X-Forwarded-For":         {nil, "192.0.2.1"},
		"trusted proxy uses X-Forwarded-For":      {echo.ExtractIPFromXFFHeader(echo.TrustIPRange(proxyNet)), "203.0.113.9"},
		"untrusted proxy ignores X-Forwarded-For": {echo.ExtractIPFromXFFHeader(), "192.0.2.1"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			conf := config("")
			conf.IPExtractor = tt.extractor
			e := xecho.Echo(conf)
			e.GET("/ip", func(c echo.Context) error { return c.String(http.StatusOK, c.RealIP()) })

			req := httptest.NewRequest(http.MethodGet, "/ip", nil)
			req.Header.Set(echo.HeaderXForwardedFor, "203.0.113.9")
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.want, rec.Body.String())
		})
	}
}

func config(routePrefix string) xecho.Config {
	config := xecho.NewConfig()
	config.RoutePrefix = routePrefix
	config.ProjectName = "acme"
	config.AppName = "login"
	config.EnvName = "dev"
	config.NewRelicLicense = "1111111111111111111111111111111111111111"
	config.NewRelicEnabled = false
	return config
}
