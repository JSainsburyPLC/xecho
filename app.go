package xecho

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo"
	"github.com/newrelic/go-agent"
	"github.com/newrelic/go-agent/_integrations/nrlogrus"
	"github.com/sirupsen/logrus"
)

const headerBuildVersion = "Build-Version"

type Xecho struct {
	Echo        *echo.Echo
	NewRelicApp newrelic.Application
}

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
	RoutePrefix       string
}

func NewConfig() Config {
	return Config{
		ProjectName:       "",
		AppName:           "",
		EnvName:           "",
		BuildVersion:      "",
		RoutePrefix:       "",
		LogLevel:          logrus.InfoLevel,
		LogFormatter:      &logrus.JSONFormatter{},
		IsDebug:           false,
		NewRelicLicense:   "",
		NewRelicEnabled:   true,
		ErrorHandler:      DefaultErrorHandler(),
		UseDefaultHeaders: true,
	}
}

func New(conf Config) *Xecho {
	e, nrApp := newEcho(conf)
	return &Xecho{NewRelicApp: nrApp, Echo: e}
}

func Echo(conf Config) *echo.Echo {
	e, _ := newEcho(conf)
	return e
}

func newEcho(conf Config) (*echo.Echo, newrelic.Application) {
	logger := logger(conf)

	newRelicApp := createNewRelicApp(conf, logger)

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Logger = appScopeLogger(logger, conf.AppName, conf.EnvName)

	// the order of these middleware is important - context should be first, error should be after logging ones
	e.Use(ContextMiddleware(conf.AppName, conf.EnvName, conf.BuildVersion, logger, conf.IsDebug, newRelicApp))
	e.Use(PanicHandlerMiddleware(conf.ErrorHandler))
	if conf.UseDefaultHeaders {
		e.Use(DefaultHeadersMiddleware())
	}
	e.Use(RequestLoggerMiddleware(time.Now))
	e.Use(DebugLoggerMiddleware(conf.IsDebug))
	e.Use(ErrorHandlerMiddleware(conf.ErrorHandler))

	addHealthCheck(conf, e)

	return e, newRelicApp
}

func addHealthCheck(conf Config, e *echo.Echo) {
	healthRoute := "/health"
	if conf.RoutePrefix != "" {
		healthRoute = fmt.Sprintf("%s/health", conf.RoutePrefix)
	}

	e.GET(healthRoute, EchoHandler(func(c *Context) error {
		if len(conf.BuildVersion) > 0 {
			c.Response().Header().Add(headerBuildVersion, conf.BuildVersion)
		}
		return c.JSONBlob(http.StatusOK, []byte(`{"status": "ok"}`))
	}))
}

func getHostName() string {
	name, err := os.Hostname()
	if err != nil {
		name = "ERROR"
	}
	return name
}

func logger(conf Config) *logrus.Logger {
	logger := logrus.New()
	logger.SetLevel(conf.LogLevel)
	logger.SetFormatter(conf.LogFormatter)
	logger.WithFields(
		logrus.Fields{
			"project":       conf.ProjectName,
			"application":   conf.AppName,
			"environment":   conf.EnvName,
			"build_version": conf.BuildVersion,
			"hostname":      getHostName(),
		}).Infof("XEcho app created %s(%s)", conf.AppName, conf.BuildVersion)
	return logger
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
