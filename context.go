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
	NewRelicApp   newrelic.Application
	NewRelicTx    newrelic.Transaction
	logger        *Logger
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
	logger *logrus.Logger,
	isDebug bool,
	newRelicApp newrelic.Application,
) echo.MiddlewareFunc {
	return func(h echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			correlationID := getCorrelationID(c.Request())
			ip := c.RealIP()
			newRelicTx := newRelicApp.StartTransaction(
				c.Request().URL.Path,
				c.Response().Writer,
				c.Request(),
			)
			defer func() { _ = newRelicTx.End() }()
			// new relic tx wraps response writer
			c.Response().Writer = newRelicTx

			cc := &Context{
				Context:       c,
				CorrelationID: correlationID,
				NewRelicApp:   newRelicApp,
				NewRelicTx:    newRelicTx,
				logger: requestScopeLogger(
					logger,
					c.Request(),
					c.Path(),
					ip,
					correlationID,
					appName,
					envName,
				),
			}

			// TODO: build version attribute (and in logs)
			cc.AddNewRelicAttribute("route", c.Path())
			cc.AddNewRelicAttribute("correlationID", correlationID)
			cc.AddNewRelicAttribute("ip", ip)

			return h(cc)
		}
	}
}

func getCorrelationID(r *http.Request) string {
	correlationID := r.Header.Get(correlationIDHeaderName)
	if correlationID != "" {
		return correlationID
	}
	return uuid.New().String()
}