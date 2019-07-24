package xecho

import "github.com/labstack/echo"

func DefaultHeadersMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return EchoHandler(func(c *Context) error {
			c.Response().Header().Set(correlationIDHeaderName, c.CorrelationID)
			c.Response().Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
			c.Response().Header().Set("Expires", "Thu, 01 Jan 1970 00:00:00 GMT")
			c.Response().Header().Set("Pragma", "no-cache")
			c.Response().Header().Set("Strict-Transport-Security", "max-age=15724800; includeSubDomains")
			c.Response().Header().Set("X-Content-Type-Options", "nosniff")
			c.Response().Header().Set("X-Frame-Options", "SAMEORIGIN")
			return next(c)
		})
	}
}
