package xecho

import (
	"github.com/labstack/echo"
	"github.com/newrelic/go-agent"
	"github.com/newrelic/go-agent/_integrations/nrlogrus"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEchoHandler(t *testing.T) {
	h := EchoHandler(func(c *Context) error {
		return assert.AnError
	})
	err := h(&Context{})
	assert.Equal(t, assert.AnError, err)
}

func TestContextMiddleware(t *testing.T) {
	ctx, _, _ := getEchoTestCtx()
	mw := ContextMiddleware("testApp", "testEnv", "build-1.2.3", NullLogger(), true, stubNewRelicApp())
	hCalled := false
	h := EchoHandler(func(c *Context) error {
		hCalled = true
		assert.Equal(t, "/test", c.Request().URL.Path)
		assert.NotEmpty(t, c.CorrelationID)
		return nil
	})

	err := mw(h)(ctx)

	assert.True(t, hCalled)
	assert.Nil(t, err)
}

func TestGetCorrelationID(t *testing.T) {
	// No id in request - generate a new one
	r := httptest.NewRequest("GET", "/", nil)
	id := getCorrelationID(r)
	assert.NotEmpty(t, id)

	// Id in request - use passed one
	r = httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Correlation-Id", "testing-id")
	id = getCorrelationID(r)
	assert.Equal(t, "testing-id", id)
}

func NullLogger() *logrus.Logger {
	log := logrus.New()
	log.SetOutput(ioutil.Discard)
	return log
}

func getEchoTestCtx() (echo.Context, *http.Request, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), req, rec
}

func stubNewRelicApp() newrelic.Application {
	config := newrelic.NewConfig("ApplicationName", "1111111111111111111111111111111111111111")
	config.Logger = nrlogrus.StandardLogger()
	config.Enabled = false
	app, _ := newrelic.NewApplication(config)
	return app
}

func stubNewRelicTX() newrelic.Transaction {
	return stubNewRelicApp().StartTransaction("test", nil, nil)
}
