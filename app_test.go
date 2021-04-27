package xecho_test

import (
	"net/http"
	"testing"

	"github.com/JSainsburyPLC/xecho"
	"github.com/steinfletcher/apitest"
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
