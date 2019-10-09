package xecho

import (
	"errors"
	"github.com/google/uuid"
	"github.com/labstack/echo"
	"github.com/newrelic/go-agent"
	"github.com/sirupsen/logrus"
	"net/http"
)

const correlationIDHeaderName = "Correlation-Id"

type Context struct {
	echo.Context
	CorrelationID string
	HttpClient    *http.Client // deprecated
	NewRelicApp   newrelic.Application
	NewRelicTx    newrelic.Transaction
	logger        *Logger
	IsDebug       bool
}

type Handler func(c *Context) error

func EchoHandler(handler Handler) echo.HandlerFunc {
	return func(c echo.Context) error {
		cc, ok := c.(*Context)
		if !ok {
			return errors.New("unable to get custom context from echo context")
		}
		return handler(cc)
	}
}

func (c *Context) Logger() echo.Logger {
	return c.logger
}

func (c *Context) AddNewRelicAttribute(key string, val interface{}) {
	if c.NewRelicTx == nil {
		return
	}
	if err := c.NewRelicTx.AddAttribute(key, val); err != nil {
		c.Logger().Errorf("failed to add attr '%s' to new relic tx: %+v", key, err)
	}
}

func ContextMiddleware(
	appName string,
	envName string,
	buildVersion string,
	logger *logrus.Logger,
	isDebug bool,
	newRelicApp newrelic.Application,
) echo.MiddlewareFunc {
	return func(h echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			correlationID := getCorrelationID(c.Request())
			ip := c.RealIP()
			logger := requestScopeLogger(
				logger,
				c.Request(),
				c.Path(),
				ip,
				correlationID,
				appName,
				envName,
				buildVersion,
			)

			cc := NewContext(c, newRelicApp, logger, correlationID, isDebug, buildVersion)

			return h(cc)
		}
	}
}

func NewContext(
	echoCtx echo.Context,
	newRelicApp newrelic.Application,
	logger *Logger,
	correlationID string,
	isDebug bool,
	buildVersion string,
) *Context {
	newRelicTx := newRelicApp.StartTransaction(
		echoCtx.Request().URL.Path,
		echoCtx.Response().Writer,
		echoCtx.Request(),
	)
	defer func() { _ = newRelicTx.End() }()
	// new relic tx wraps response writer
	echoCtx.Response().Writer = newRelicTx

	customCtx := &Context{
		Context:       echoCtx,
		CorrelationID: correlationID,
		NewRelicApp:   newRelicApp,
		NewRelicTx:    newRelicTx,
		logger:        logger,
		IsDebug:       isDebug,
	}

	// deprecated in favour of context.AppendNewRelicToClient()
	customCtx.HttpClient = NewHttpClient(customCtx, &http.Client{Transport: &http.Transport{}})

	// TODO: build version attribute (and in logs)
	customCtx.AddNewRelicAttribute("route", echoCtx.Path())
	customCtx.AddNewRelicAttribute("correlationID", correlationID)
	customCtx.AddNewRelicAttribute("ip", echoCtx.RealIP())
	customCtx.AddNewRelicAttribute("buildVersion", buildVersion)

	return customCtx
}

func getCorrelationID(r *http.Request) string {
	correlationID := r.Header.Get(correlationIDHeaderName)
	if correlationID != "" {
		return correlationID
	}
	return uuid.New().String()
}
