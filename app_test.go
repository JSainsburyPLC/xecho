package xecho_test

import (
	"github.com/JSainsburyPLC/xecho"
	"github.com/steinfletcher/apitest"
	"net/http"
	"testing"
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
		Handler(xecho.Echo(config("test"))).
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
