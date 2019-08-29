package xecho

import (
	"fmt"
	"net/http"

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
	IsDebug           bool
	NewRelicLicense   string
	NewRelicEnabled   bool
	ErrorHandler      ErrorHandlerFunc
	UseDefaultHeaders bool
}

func NewConfig() Config {
	return Config{
		ProjectName:       "",
		AppName:           "",
		EnvName:           "",
		BuildVersion:      "",
		LogLevel:          logrus.InfoLevel,
		IsDebug:           false,
		NewRelicLicense:   "",
		NewRelicEnabled:   true,
		ErrorHandler:      DefaultErrorHandler(),
		UseDefaultHeaders: true,
	}
}

func Echo(conf Config) *echo.Echo {
	logger := logrus.New()
	logger.SetLevel(conf.LogLevel)
	logger.SetFormatter(&logrus.JSONFormatter{})

	newRelicApp := createNewRelicApp(conf, logger)

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Logger = appScopeLogger(logger, conf.AppName, conf.EnvName, conf.BuildVersion)

	// the order of these middleware is important - context should be first, error should be after logging ones
	e.Use(ContextMiddleware(conf.AppName, conf.EnvName, conf.BuildVersion, logger, conf.IsDebug, newRelicApp))
	e.Use(PanicHandlerMiddleware(conf.ErrorHandler))
	if conf.UseDefaultHeaders {
		e.Use(DefaultHeadersMiddleware())
	}
	e.Use(RequestLoggerMiddleware())
	e.Use(DebugLoggerMiddleware(conf.IsDebug))
	e.Use(ErrorHandlerMiddleware(conf.ErrorHandler))

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
