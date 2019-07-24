package xecho

import (
	"fmt"
	"github.com/labstack/echo"
	"net/http"
	"runtime"
	"strings"
)

const stackSize = 4 << 10 // 4kb

type Error struct {
	Status int               `json:"-"`
	Code   string            `json:"code"`
	Detail string            `json:"detail"`
	Params map[string]string `json:"-"`
}

func (err *Error) Error() string {
	errorParts := []string{
		fmt.Sprintf("Code: %s; Status: %d; Detail: %s", err.Code, err.Status, err.Detail),
	}
	for k, v := range err.Params {
		errorParts = append(errorParts, fmt.Sprintf("%s: %s", k, v))
	}
	return strings.Join(errorParts, "; ")
}

type ErrorHandlerFunc func(c *Context, err *Error)

func ErrorHandlerMiddleware(errorHandler ErrorHandlerFunc) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			if err != nil {
				cc := c.(*Context)
				handleError(errorHandler, cc, err)
			}
			return nil
		}
	}
}

func PanicHandlerMiddleware(errorHandler ErrorHandlerFunc) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					err, ok := r.(error)
					if !ok {
						err = fmt.Errorf("%v", r)
					}
					stack := make([]byte, stackSize)
					length := runtime.Stack(stack, true)
					c.Logger().Printf("[PANIC RECOVER] %v %s\n", err, stack[:length])
					handleError(errorHandler, c.(*Context), err)
				}
			}()
			return next(c)
		}
	}
}

func DefaultErrorHandler() ErrorHandlerFunc {
	return func(c *Context, err *Error) {
		_ = c.JSON(err.Status, err)
	}
}

func handleError(errorHandler ErrorHandlerFunc, c *Context, err error) {
	var newErr *Error
	switch err := err.(type) {
	case *Error:
		newErr = err
	case *echo.HTTPError:
		newErr = &Error{
			Status: err.Code,
			Code:   "ECHO_HTTP_ERROR",
			Detail: fmt.Sprintf("%v", err.Message),
			Params: map[string]string{"reason": err.Error()},
		}
	default:
		newErr = &Error{
			Status: ErrInternalServer.Status,
			Code:   ErrInternalServer.Code,
			Detail: ErrInternalServer.Detail,
			Params: map[string]string{"reason": err.Error()},
		}
	}
	recordError(newErr, c)
	errorHandler(c, newErr)
}

func recordError(err *Error, c *Context) {
	c.Logger().Error(err)
	c.AddNewRelicAttribute("errorCode", err.Code)
	c.AddNewRelicAttribute("errorDetail", err.Detail)
	c.AddNewRelicAttribute("errorReason", err.Params["reason"])
}

var ErrInternalServer = &Error{
	Status: http.StatusInternalServerError,
	Code:   "INTERNAL_SERVER_ERROR",
	Detail: "Internal server error",
}

var ErrBadRequest = &Error{
	Status: http.StatusBadRequest,
	Code:   "BAD_REQUEST",
	Detail: "Bad request",
}

var ErrUnauthorised = &Error{
	Status: http.StatusUnauthorized,
	Code:   "UNAUTHORISED",
	Detail: "Unauthorised",
}

var ErrNotFound = &Error{
	Status: http.StatusNotFound,
	Code:   "NOT_FOUND",
	Detail: "Not found",
}

var ErrMethodNotAllowed = &Error{
	Status: http.StatusMethodNotAllowed,
	Code:   "METHOD_NOT_ALLOWED",
	Detail: "Method not allowed",
}
