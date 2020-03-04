package xecho

import (
	"math"
	"net/http"
	"time"

	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
)

type TimeProvider func() time.Time

func RequestLoggerMiddleware(timeFn TimeProvider) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return EchoHandler(func(c *Context) error { return RequestLogger(c, next, timeFn) })
	}
}

func RequestLogger(c *Context, next echo.HandlerFunc, time TimeProvider) error {
	before := time()
	lrw := &statefulResponseWriter{ResponseWriter: c.Response().Writer}
	c.Response().Writer = lrw
	err := next(c)
	after := time()
	logger, ok := c.Logger().(*Logger)
	if !ok {
		c.Logger().Infof("[%s] %s %d", c.Request().Method, c.Path(), lrw.statusCode)
		return err
	}
	logger.
		WithFields(createMap(c, after.Sub(before), lrw, err)).
		Infof("[%s] %s %d", c.Request().Method, c.Path(), lrw.statusCode)
	return err
}

func createMap(c echo.Context, timeTaken time.Duration, lrw *statefulResponseWriter, err error) logrus.Fields {
	r := c.Request()
	fields := logrus.Fields{
		"duration_ms": milliseconds(timeTaken),
		"request":     requestMap(r, c),
		"response":    responseMap(c.Response(), lrw.statusCode),
	}

	if err != nil {
		fields["error"] = err.Error()
	}

	return fields
}

func milliseconds(timeTaken time.Duration) int64 {
	return int64(timeTaken) / 1e6
}

func responseMap(r *echo.Response, statusCode int) logrus.Fields {
	return logrus.Fields{
		"status_code":    statusCode,
		"content_length": math.Max(float64(r.Size), 0),
	}
}

func requestMap(r *http.Request, c echo.Context) logrus.Fields {
	return logrus.Fields{
		"method":         r.Method,
		"host_name":      r.Host,
		"query_params":   c.QueryParams(),
		"Content-length": math.Max(float64(r.ContentLength), 0),
		"headers": logrus.Fields{
			"user-agent":        r.UserAgent(),
			"referer":           r.Referer(),
			"x-forwarded-for":   r.Header.Get("X-Forwarded-For"),
			"x-forwarded-proto": r.Header.Get("X-Forwarded-Proto"),
		},
	}
}
