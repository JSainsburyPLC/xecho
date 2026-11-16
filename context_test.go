package xecho

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/newrelic/go-agent"
	"github.com/newrelic/go-agent/_integrations/nrlogrus"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
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
	mw := ContextMiddleware("build-1.2.3", NullLogger().WithFields(logrus.Fields{
		"application": "testApp",
		"environment": "testEnv",
	}), true, stubNewRelicApp())
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

type endErrorTx struct{ newrelic.Transaction }

func (endErrorTx) End() error { return assert.AnError }

type endErrorApp struct{ newrelic.Application }

func (a endErrorApp) StartTransaction(name string, w http.ResponseWriter, r *http.Request) newrelic.Transaction {
	return endErrorTx{a.Application.StartTransaction(name, w, r)}
}

func TestContextMiddleware_LogsNewRelicEndError(t *testing.T) {
	t.Parallel()

	ctx, _, _ := getEchoTestCtx()
	var buf bytes.Buffer
	logger := logrus.New()
	logger.SetOutput(&buf)
	mw := ContextMiddleware("build-1.2.3", logrus.NewEntry(logger), false, endErrorApp{stubNewRelicApp()})

	err := mw(EchoHandler(func(c *Context) error { return nil }))(ctx)

	assert.Nil(t, err)
	assert.Contains(t, buf.String(), "failed to end New Relic transaction")
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
	log.SetOutput(io.Discard)
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
