package xecho

import (
	"net/http"
	"time"

	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
)

type TimeProvider func() time.Time

func RequestLoggerMiddleware(timeFn TimeProvider) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return EchoHandler(func(c *Context) error {
			return RequestLogger(c, next, timeFn)
		})
	}
}

func RequestLogger(c *Context, next echo.HandlerFunc, time TimeProvider) error {
	before := time()
	err := next(c)
	after := time()

	statusCode := getStatusFromError(c, err)
	logger, ok := c.Logger().(*Logger)
	if !ok {
		c.Logger().Infof("[%s] %s %d", c.Request().Method, c.Path(), statusCode)
		return err
	}
	logger.
		WithFields(createMap(c, after.Sub(before), statusCode, err)).
		Infof("[%s] %s %d", c.Request().Method, c.Path(), statusCode)
	return err
}

func getStatusFromError(c *Context, err error) (statusCode int) {
	code := c.Response().Status
	if err != nil {
		aError, ok := err.(*Error)
		if aError != nil && ok {
			code = aError.Status
		}
		code = 500
	}
	return code
}

func createMap(c echo.Context, timeTaken time.Duration, statusCode int, anError error) logrus.Fields {
	r := c.Request()
	fields := logrus.Fields{
		"duration_ms": milliseconds(timeTaken),
		"request":     requestMap(r, c),
		"response":    responseMap(c.Response(), statusCode),
	}
	if anError == nil {
		return fields

	} else {
		fields["error"] = errorFields(anError)
		return fields
	}

}

func errorFields(err error) logrus.Fields {
	fields := logrus.Fields{
		"message": err.Error(),
	}

	panicError, isPanic := err.(*PanicError)
	if panicError != nil && isPanic {
		err = panicError.error
	}

	exErr, isError := err.(*Error)
	if exErr != nil && isError {
		fields["status"] = exErr.Status
		fields["code"] = exErr.Code
		fields["detail"] = exErr.Detail
		fields["params"] = exErr.Params
	}
	return fields
}

func milliseconds(timeTaken time.Duration) int64 {
	return int64(timeTaken) / 1e6
}

func responseMap(r *echo.Response, statusCode int) logrus.Fields {
	fields := logrus.Fields{
		"status_code":    statusCode,
		"content_length": r.Size,
	}
	return fields
}

func requestMap(r *http.Request, c echo.Context) logrus.Fields {
	fields := logrus.Fields{
		"method":       r.Method,
		"host_name":    r.Host,
		"query_params": c.QueryParams(),
		"headers": logrus.Fields{
			"user-agent": r.UserAgent(),
			"referer":    r.Referer(),
		},
	}
	if r.ContentLength > 0 {
		fields["Content-length"] = r.ContentLength
	}
	return fields
}
