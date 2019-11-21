package xecho

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo"
	"github.com/newrelic/go-agent"
	"github.com/newrelic/go-agent/_integrations/nrlogrus"
	"github.com/sirupsen/logrus"
)

const headerBuildVersion = "Build-Version"

type Config struct {
	ProjectName       string
	AppName           string
	EnvName           string
	BuildVersion      string
	LogLevel          logrus.Level
	LogFormatter      logrus.Formatter
	IsDebug           bool
	NewRelicLicense   string
	NewRelicEnabled   bool
	ErrorHandler      ErrorHandlerFunc
	UseDefaultHeaders bool
	LoggerProvider    func(config Config) *logrus.Logger
}

func NewConfig() Config {
	return Config{
		ProjectName:       "",
		AppName:           "",
		EnvName:           "",
		BuildVersion:      "",
		LogLevel:          logrus.InfoLevel,
		LogFormatter:      &logrus.JSONFormatter{},
		IsDebug:           false,
		NewRelicLicense:   "",
		NewRelicEnabled:   true,
		ErrorHandler:      DefaultErrorHandler(),
		UseDefaultHeaders: true,
		LoggerProvider:    newLogger,
	}
}

func Echo(conf Config) *echo.Echo {

	logger := conf.LoggerProvider(conf)

	newRelicApp := createNewRelicApp(conf, logger)

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Logger = appScopeLogger(logger, conf.AppName, conf.EnvName, conf.BuildVersion)

	// The order of the beginning middleware is important:
	// - ContextMiddleware should be first: Creating the context - head call (before next())
	// - Error handling return message and New Relic logging - tail call (after next())
	// - Request logger is next to log the request as well as the converted error on a - tail call
	// - Convert error ( pre error handler) to convert the error coming down from the handlers in the call chain - tail call.
	// - Panic handler catch any panics caused by handlers that are added to the framework and convert them into an error to be handled by the error handler - tail call
	e.Use(ContextMiddleware(conf.AppName, conf.EnvName, conf.BuildVersion, logger, conf.IsDebug, newRelicApp))
	e.Use(ErrorHandlerMiddleware(conf.ErrorHandler))
	e.Use(RequestLoggerMiddleware(time.Now))
	e.Use(ErrorConverter())
	e.Use(PanicHandlerMiddleware(conf.ErrorHandler))

	if conf.UseDefaultHeaders {
		e.Use(DefaultHeadersMiddleware())
	}
	e.Use(DebugLoggerMiddleware(conf.IsDebug))
	e.GET("/health", EchoHandler(func(c *Context) error {
		if len(conf.BuildVersion) > 0 {
			c.Response().Header().Add(headerBuildVersion, conf.BuildVersion)
		}
		return c.JSONBlob(http.StatusOK, []byte(`{"status": "ok"}`))
	}))

	return e
}

func createNewRelicApp(conf Config, logger *logrus.Logger) newrelic.Application {
	nrConf := newrelic.NewConfig(fmt.Sprintf("%s-%s-%s", conf.ProjectName, conf.AppName, conf.EnvName), conf.NewRelicLicense)
	nrConf.CrossApplicationTracer.Enabled = false
	nrConf.DistributedTracer.Enabled = true
	nrConf.Logger = nrlogrus.Transform(logger)
	nrConf.Enabled = conf.NewRelicEnabled
	nrConf.Labels = map[string]string{"Env": conf.EnvName, "Project": conf.ProjectName}
	app, err := newrelic.NewApplication(nrConf)
	if err != nil {
		panic(fmt.Sprintf("Failed to register New Relic Agent, error: %s", err.Error()))
	}
	return app
}
