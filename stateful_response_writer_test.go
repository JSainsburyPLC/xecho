package xecho

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type responseWriter struct {
	called map[string]int
	status int
}

func (r responseWriter) Header() http.Header {
	r.incCall("Header")

	return http.Header{}
}

func (r *responseWriter) incCall(methodName string) {
	if val, ok := r.called[methodName]; ok {
		r.called[methodName] = val + 1
	} else {
		r.called[methodName] = 1
	}
}

func (r *responseWriter) Write([]byte) (int, error) {
	r.incCall("Write")
	return 0, nil
}

func (r *responseWriter) WriteHeader(statusCode int) {
	r.incCall("WriteHeader")
	r.status = statusCode
}

func NewWriter() (*responseWriter, *statefulResponseWriter) {
	response := responseWriter{called: map[string]int{}}
	writer := statefulResponseWriter{ResponseWriter: &response}
	return &response, &writer
}

var _ http.ResponseWriter = &responseWriter{}

func TestStatefulResponseWriter_WriteHeaderTest(t *testing.T) {
	response, writer := NewWriter()
	writer.WriteHeader(505)
	assert.Equal(t, response.called["WriteHeader"], 1)
	assert.Equal(t, writer.statusCode, 505)
	assert.Equal(t, response.status, 505)
}

func TestStatefulResponseWriter_WriteTest(t *testing.T) {
	response, writer := NewWriter()
	_, _ = writer.Write([]byte{})
	assert.Equal(t, response.called["Write"], 1)
}

func TestStatefulResponseWriter_HeaderTest(t *testing.T) {
	response, writer := NewWriter()
	writer.Header()
	assert.Equal(t, response.called["Header"], 1)
}
