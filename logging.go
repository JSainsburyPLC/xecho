package xecho

import (
	"bytes"
	"encoding/json"
	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
)

func RequestLoggerMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return EchoHandler(func(c *Context) error {
			lrw := &statefulResponseWriter{w: c.Response().Writer}
			c.Response().Writer = lrw

			c.Logger()

			c.Logger().Infof("Inbound request on path: '%s'", c.Request().URL.Path)
			err := next(c)
			c.Logger().Infof("Response with code: '%d'", lrw.statusCode)

			return err
		})
	}
}

func DebugLoggerMiddleware(isDebug bool) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return EchoHandler(func(c *Context) error {
			if !isDebug {
				return next(c)
			}

			drw := &debugResponseWriter{w: c.Response().Writer}
			c.Response().Writer = drw

			dumpRequest(c)
			err := next(c)
			dumpResponse(c, drw)

			return err
		})
	}
}

type debugResponseWriter struct {
	isDebug    bool
	w          http.ResponseWriter
	statusCode int
	body       bytes.Buffer
}

func (drw *debugResponseWriter) Header() http.Header {
	return drw.w.Header()
}

func (drw *debugResponseWriter) Write(b []byte) (int, error) {
	_, _ = drw.body.Write(b)
	return drw.w.Write(b)
}

func (drw *debugResponseWriter) WriteHeader(code int) {
	drw.statusCode = code
	drw.w.WriteHeader(code)
}

func dumpRequest(c *Context) {
	reqDump, err := httputil.DumpRequest(c.Request(), true)
	if err == nil {
		c.Logger().Debugf("%s", string(reqDump))
	}
}

func dumpResponse(c *Context, drw *debugResponseWriter) {
	res := &http.Response{
		Body:          ioutil.NopCloser(&drw.body),
		Header:        drw.Header(),
		StatusCode:    drw.statusCode,
		ProtoMajor:    1,
		ProtoMinor:    1,
		ContentLength: int64(drw.body.Len()),
	}

	body := res.Header.Get("Content-Type") != "text/html"
	resDump, err := httputil.DumpResponse(res, body)
	if err == nil {
		c.Logger().Debugf("%s", string(resDump))
	}
}

func appScopeLogger(
	logger *logrus.Logger,
	appName string,
	envName string,
) *Logger {
	entry := logger.WithFields(logrus.Fields{
		"app":   appName,
		"env":   envName,
		"scope": "app",
	})
	return &Logger{entry}
}

func requestScopeLogger(
	logger *logrus.Logger,
	r *http.Request,
	route string,
	ip string,
	correlationID string,
	appName string,
	envName string,
) *Logger {
	ctxLogger := logger.WithFields(logrus.Fields{
		"app":            appName,
		"env":            envName,
		"scope":          "request",
		"correlation_id": correlationID,
		"url":            r.RequestURI,
		"route":          route,
		"remote_addr":    r.RemoteAddr,
		"method":         r.Method,
		"ip":             ip,
		"headers": logrus.Fields{
			"host":              r.Host,
			"user-agent":        r.UserAgent(),
			"referer":           r.Referer(),
			"x-forwarded-for":   r.Header.Get("X-Forwarded-For"),
			"x-forwarded-proto": r.Header.Get("X-Forwarded-Proto"),
		},
	})
	return &Logger{ctxLogger}
}

// Wrap logrus entry and implement additional methods required to
// satisfy echo logger interface
type Logger struct {
	*logrus.Entry
}

func (l *Logger) Output() io.Writer {
	return l.Logger.Out
}

func (l *Logger) SetOutput(w io.Writer) {
	l.Logger.Out = w
}

func (l *Logger) Prefix() string {
	return "" // not implemented - only added for API compatibility with echo logger
}

func (l *Logger) SetPrefix(prefix string) {
	// not implemented - only added for API compatibility with echo logger
}

func (l *Logger) Level() log.Lvl {
	return logrusLeveltoEchoLevel(l.Logger.Level)
}

func (l *Logger) SetLevel(lvl log.Lvl) {
	l.Logger.Level = echoLeveltoLogrusLevel(lvl)
}

func (l *Logger) SetHeader(h string) {
	// not implemented - only added for API compatibility with echo logger
}

func (l *Logger) Printj(j log.JSON) {
	b, err := json.Marshal(j)
	if err != nil {
		panic(err)
	}
	l.Println(string(b))
}

func (l *Logger) Debugj(j log.JSON) {
	b, err := json.Marshal(j)
	if err != nil {
		panic(err)
	}
	l.Debugln(string(b))
}

func (l *Logger) Infoj(j log.JSON) {
	b, err := json.Marshal(j)
	if err != nil {
		panic(err)
	}
	l.Infoln(string(b))
}

func (l *Logger) Warnj(j log.JSON) {
	b, err := json.Marshal(j)
	if err != nil {
		panic(err)
	}
	l.Warnln(string(b))
}

func (l *Logger) Errorj(j log.JSON) {
	b, err := json.Marshal(j)
	if err != nil {
		panic(err)
	}
	l.Errorln(string(b))
}

func (l *Logger) Fatalj(j log.JSON) {
	b, err := json.Marshal(j)
	if err != nil {
		panic(err)
	}
	l.Fatalln(string(b))
}

func (l *Logger) Panicj(j log.JSON) {
	b, err := json.Marshal(j)
	if err != nil {
		panic(err)
	}
	l.Panicln(string(b))
}

func echoLeveltoLogrusLevel(level log.Lvl) logrus.Level {
	switch level {
	case log.DEBUG:
		return logrus.DebugLevel
	case log.INFO:
		return logrus.InfoLevel
	case log.WARN:
		return logrus.WarnLevel
	case log.ERROR:
		return logrus.ErrorLevel
	}

	return logrus.InfoLevel
}

func logrusLeveltoEchoLevel(level logrus.Level) log.Lvl {
	switch level {
	case logrus.DebugLevel:
		return log.DEBUG
	case logrus.InfoLevel:
		return log.INFO
	case logrus.WarnLevel:
		return log.WARN
	case logrus.ErrorLevel:
		return log.ERROR
	}

	return log.OFF
}
