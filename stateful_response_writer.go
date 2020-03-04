package xecho

import "net/http"

type statefulResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *statefulResponseWriter) WriteHeader(code int) {
	lrw.ResponseWriter.WriteHeader(code)
	lrw.statusCode = code
}
