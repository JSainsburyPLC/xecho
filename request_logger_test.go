package xecho

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var _ TimeProvider = (&testTimeProvider{}).Next

type testTimeProvider struct {
	index int
	calls []time.Time
}

func (t *testTimeProvider) Next() time.Time {
	i := t.index
	t.index++
	return t.calls[i]
}

type testContext struct {
	Context
	response    *echo.Response
	queryParams url.Values
	request     *http.Request
}

var _ echo.Context = &testContext{}

func (t *testContext) Response() *echo.Response {
	return t.response
}

func (t *testContext) Request() *http.Request {
	return t.request
}

func (t *testContext) Path() string {
	return t.request.URL.Path
}

func (t *testContext) QueryParams() url.Values {
	return t.queryParams
}

const urlTo = "https://this.is.a.domain/this/is/the/site?aparam=aValue&aparam2:avalue2"

func TestRequestLogger_LogTest(t *testing.T) {
	buffer := &bytes.Buffer{}
	URL, _ := url.Parse(urlTo)
	writer, _ := NewWriter()
	context := createTestContext(writer, URL, buffer)
	now := time.Now()
	provider := testTimeProvider{calls: []time.Time{now, now.Add(755 * time.Millisecond)}}
	nextCalled := false
	nextPtr := &nextCalled
	var next echo.HandlerFunc = func(context echo.Context) error {
		*nextPtr = true
		context.Response().WriteHeader(200)
		return nil
	}
	err := RequestLoggerMiddleware(provider.Next)(next)(context)
	assert.Nil(t, err)
	fields := getLogFields(buffer, err, t)

	assert.Equal(t, "[GET] /this/is/the/site 200", fields["msg"])
	assert.Equal(t, "set_one", fields["correlation_id"])
	assert.Equal(t, "request", fields["scope"])
	assert.Equal(t, float64(755), fields["duration_ms"])

	request, ok := fields["request"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "this.is.a.domain", request["host_name"])
	assert.Equal(t, "GET", request["method"])
	//assert.Equal(t, []interface{}{"ONE", "ONE", "TWO"}, request["cookies"])

	response, ok := fields["response"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, float64(200), response["status_code"])
	assert.Equal(t, float64(150), response["content_length"])

}

func TestRequestLogger_HealthNoLogTest(t *testing.T) {
	buffer := &bytes.Buffer{}
	URL, _ := url.Parse("http://somedomain/health")
	writer, _ := NewWriter()
	context := createTestContext(writer, URL, buffer)
	now := time.Now()
	provider := testTimeProvider{calls: []time.Time{now, now.Add(755 * time.Millisecond)}}
	nextCalled := false
	nextPtr := &nextCalled
	var next echo.HandlerFunc = func(context echo.Context) error {
		*nextPtr = true
		context.Response().WriteHeader(200)
		return nil
	}
	err := RequestLoggerMiddleware(provider.Next)(next)(context)
	assert.Nil(t, err)
	fields := getLogFields(buffer, err, t)

	assert.Equal(t, len(fields), 0)

}

func createTestContext(writer *responseWriter, URL *url.URL, buffer *bytes.Buffer) *Context {
	return &Context{
		Context: &testContext{
			queryParams: url.Values{
				"aparam":  []string{"aValue"},
				"aparam2": []string{"avalue2"},
			},
			response: createResponse(writer),
			request:  createRequest(URL),
		},
		logger: &Logger{Entry: createLogger(buffer).
			WithField("correlation_id", "set_one").
			WithField("scope", "request")},
	}
}

func createResponse(writer *responseWriter) *echo.Response {
	return &echo.Response{
		Writer: writer,
		Status: 200,
		Size:   150,
	}
}
func createRequest(URL *url.URL) *http.Request {
	return &http.Request{
		Method:        "GET",
		URL:           URL,
		Header:        http.Header{"Correlation-Id": []string{"set_one"}, "User-Agent": []string{"ELB-HealthChecker/2.0"}},
		ContentLength: 34567,
		Host:          "this.is.a.domain",
		RemoteAddr:    "",
		RequestURI:    urlTo,
	}
}

func getLogFields(buffer *bytes.Buffer, err error, t *testing.T) logrus.Fields {
	fields := logrus.Fields{}
	logStatement := buffer.Bytes()
	if len(logStatement) > 0 {
		err = json.Unmarshal(logStatement, &fields)
		assert.Nil(t, err)
	}
	return fields
}

func createLogger(buffer *bytes.Buffer) *logrus.Logger {
	log := logrus.New()
	log.Out = buffer
	log.Formatter = &logrus.JSONFormatter{}
	return log
}
