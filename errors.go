package xecho

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"

	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
)

const stackSize = 4 << 10 // 4kb

type Error struct {
	Status int               `json:"-" example:"400"`
	Code   string            `json:"code" example:"BAD_REQUEST"`
	Detail string            `json:"detail" example:"Bad request"`
	Params map[string]string `json:"-"`
}

// this causes problems Error struct and Error method ( hence the error member in PanicError) ... changing this would mean a breaking change hmmm
func (err *Error) Error() string {
	errorParts := []string{
		fmt.Sprintf("Code: %s; Status: %d; Detail: %s", err.Code, err.Status, err.Detail),
	}
	for k, v := range err.Params {
		errorParts = append(errorParts, fmt.Sprintf("%s: %s", k, v))
	}
	return strings.Join(errorParts, "; ")
}

var _ error = &PanicError{}

type PanicError struct {
	error *Error
	stack []byte
}

func (p PanicError) Error() string {
	reason, found := p.error.Params["reason"]
	if found {
		return "PANIC:" + reason
	}
	return p.error.Error()
}

type ErrorHandlerFunc func(c *Context, err *Error)

func ErrorConverter() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {

		return func(c echo.Context) error {
			err := next(c)
			if err != nil {
				return convertToXEchoError(err)
			}
			return err
		}
	}
}

func PanicHandlerMiddleware(errorHandler ErrorHandlerFunc) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (returnErr error) {
			defer func() {
				if r := recover(); r != nil {
					err, ok := r.(error)
					if !ok {
						err = fmt.Errorf("PANIC: %v", r)
					}
					stack := make([]byte, stackSize)
					length := runtime.Stack(stack, true)
					panicError := &PanicError{error: convertToXEchoError(err).(*Error), stack: stack[:length]}
					returnErr = panicError
				}
			}()
			return next(c)
		}
	}
}

func DefaultErrorHandler() ErrorHandlerFunc {
	return func(c *Context, err *Error) { _ = c.JSON(err.Status, err) }
}

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

func handleError(errorHandler ErrorHandlerFunc, c *Context, theError error) {
	var errorType *Error
	switch theError.(type) {
	case *Error:
		errorType = theError.(*Error)
		errorLogger(c, errorType).Warn()
	case *PanicError:
		panicError := theError.(*PanicError)
		errorType = panicError.error
		errorLogger(c, errorType).
			WithField("stack_trace", string(panicError.stack)).
			Errorf("PANIC ERROR: %s", errorType.Detail)
	default:
		handleError(errorHandler, c, convertToXEchoError(theError))
		return
	}
	logNewRelic(c, errorType)
	errorHandler(c, errorType)
}

func convertToXEchoError(err error) error {
	if err == nil {
		return nil
	}
	var newErr error
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
	case *PanicError:
		newErr = err
	default:
		newErr = &Error{
			Status: ErrInternalServer.Status,
			Code:   ErrInternalServer.Code,
			Detail: ErrInternalServer.Detail,
			Params: map[string]string{
				"type":   fmt.Sprintf("%T", err),
				"reason": err.Error()},
		}
	}
	return newErr
}

func logNewRelic(c *Context, err *Error) {
	c.AddNewRelicAttribute("errorCode", err.Code)
	c.AddNewRelicAttribute("errorDetail", err.Detail)
	c.AddNewRelicAttribute("errorReason", err.Params["reason"])
	c.AddNewRelicAttribute("errorType", err.Params["type"])
}

func errorLogger(c *Context, err *Error) *logrus.Entry {
	logger := c.Logger().(*Logger).WithFields(logrus.Fields{
		"detail": err.Detail,
		"code":   err.Code,
		"status": err.Status,
		"params": err.Params,
	})
	return logger
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
