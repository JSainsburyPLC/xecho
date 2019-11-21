package xecho_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
	"github.com/steinfletcher/apitest"
	"github.com/stretchr/testify/assert"

	"github.com/JSainsburyPLC/xecho"
)

func TestHealthCheck(t *testing.T) {
	apitest.New().
		Handler(xecho.Echo(config())).
		Get("/health").
		Expect(t).
		Status(http.StatusOK).
		Body(`{"status": "ok"}`).
		End()
}

func TestRequestLog(t *testing.T) {
	buffer, e := createTestLoggingApp()

	apitest.New().
		Handler(e).
		Get("/loggingTest").
		QueryParams(map[string]string{"one": "onevalue", "two": "twoValue"}).
		Expect(t).
		Status(http.StatusOK).
		Body(`{"some": "value"}`).
		End()

	parsedList := getParsedLogs(buffer, t)

	assert.Equal(t, len(parsedList), 2)
	assert.Equal(t, "/loggingTest?one=onevalue&two=twoValue", parsedList[1]["url"])
	response := parsedList[1]["response"].(map[string]interface{})
	assert.Equal(t, float64(200), response["status_code"])
}

func TestRequest_withError_Log(t *testing.T) {
	buffer, e := createTestLoggingApp()

	apitest.New().
		Handler(e).
		Get("/loggingTest/error").
		Expect(t).
		Status(http.StatusInternalServerError).
		Body(`{"code":"INTERNAL_SERVER_ERROR", "detail":"Internal server error"}`).
		End()

	print(buffer.String())
	parsedList := getParsedLogs(buffer, t)

	assert.Equal(t, len(parsedList), 3)
	error := parsedList[1]["error"].(map[string]interface{})
	assert.Equal(t, error["code"], "INTERNAL_SERVER_ERROR")
	assert.Equal(t, error["params"].(map[string]interface{})["reason"], "SomeErrorHappened")
	response := parsedList[1]["response"].(map[string]interface{})
	assert.Equal(t, float64(500), response["status_code"])
}

func TestRequest_withPanic_Log(t *testing.T) {
	buffer, e := createTestLoggingApp()

	apitest.New().
		Handler(e).
		Get("/loggingTest/panic").
		Expect(t).
		Status(http.StatusInternalServerError).
		Body(`{"code":"INTERNAL_SERVER_ERROR", "detail":"Internal server error"}`).
		End()

	print(buffer.String())
	parsedList := getParsedLogs(buffer, t)

	assert.Equal(t, len(parsedList), 3)
	error := parsedList[1]["error"].(map[string]interface{})
	assert.Equal(t, "INTERNAL_SERVER_ERROR", error["code"])
	assert.Equal(t, "PANIC: This is a panic message", error["params"].(map[string]interface{})["reason"])
	response := parsedList[1]["response"].(map[string]interface{})
	assert.Equal(t, float64(500), response["status_code"])

	errorLog := parsedList[2]["stack_trace"].(string)
	assert.True(t, strings.Contains(errorLog, "TestRequest_withPanic_Log"), "The stack trace has this test in the stack listing.")
}

func getParsedLogs(buffer *bytes.Buffer, t *testing.T) []map[string]interface{} {
	logs := strings.Split(buffer.String(), "\n")
	var parsedList []map[string]interface{}
	for _, item := range logs {
		if len(item) > 0 {
			itemMap := map[string]interface{}{}
			err := json.Unmarshal([]byte(item), &itemMap)
			parsedList = append(parsedList, itemMap)
			assert.Nil(t, err)
		}
	}
	return parsedList
}

func createTestLoggingApp() (*bytes.Buffer, *echo.Echo) {
	conf := config()
	buffer := &bytes.Buffer{}
	conf.LoggerProvider = func(config xecho.Config) *logrus.Logger { return createLogger(buffer) }
	echo := xecho.Echo(conf)
	echo.GET("/loggingTest", xecho.EchoHandler(func(c *xecho.Context) error {
		return c.JSONBlob(http.StatusOK, []byte(`{"some": "value"}`))
	}))
	echo.GET("/loggingTest/error", xecho.EchoHandler(func(c *xecho.Context) error {
		return errors.New("SomeErrorHappened")
	}))
	echo.GET("/loggingTest/panic", xecho.EchoHandler(func(c *xecho.Context) error {
		panic("This is a panic message")
	}))
	return buffer, echo
}

func createLogger(buffer *bytes.Buffer) *logrus.Logger {
	log := logrus.New()
	log.Out = buffer
	log.Formatter = &logrus.JSONFormatter{}
	return log
}

func config() xecho.Config {
	config := xecho.NewConfig()
	config.ProjectName = "acme"
	config.AppName = "login"
	config.EnvName = "dev"
	config.NewRelicLicense = "1111111111111111111111111111111111111111"
	config.NewRelicEnabled = false
	return config
}
