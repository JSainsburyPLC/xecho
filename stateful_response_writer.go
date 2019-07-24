package xecho

import "net/http"

type statefulResponseWriter struct {
	w          http.ResponseWriter
	statusCode int
}

func (lrw *statefulResponseWriter) Header() http.Header {
	return lrw.w.Header()
}

func (lrw *statefulResponseWriter) Write(b []byte) (int, error) {
	return lrw.w.Write(b)
}

func (lrw *statefulResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.w.WriteHeader(code)
}
