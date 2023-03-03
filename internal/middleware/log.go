package middleware

import (
	"bufio"
	"errors"
	"net"
	"net/http"
)

type LoggingMiddleware struct {
	logger logger
}

func NewLoggingMiddleware(l logger) *LoggingMiddleware {
	return &LoggingMiddleware{
		logger: l,
	}
}

type logger interface {
	Infof(format string, args ...interface{})
}

type responseWriterInterceptor struct {
	http.ResponseWriter
	statusCode int
}

// Logging middleware to log http requests
func (lm *LoggingMiddleware) Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		wi := &responseWriterInterceptor{
			statusCode:     http.StatusOK,
			ResponseWriter: w,
		}

		next.ServeHTTP(wi, r)
		msg := wi.Header().Get("err")
		if msg != "" {
			lm.logger.Infof("%s %s %d %s", r.Method, r.RequestURI, wi.statusCode, msg)
			return
		}

		lm.logger.Infof("%s %s %d", r.Method, r.RequestURI, wi.statusCode)
	})
}

func (w *responseWriterInterceptor) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseWriterInterceptor) Write(p []byte) (int, error) {
	return w.ResponseWriter.Write(p)
}

func (w *responseWriterInterceptor) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("type assertion failed http.ResponseWriter not a http.Hijacker")
	}
	return h.Hijack()
}

func (w *responseWriterInterceptor) Flush() {
	f, ok := w.ResponseWriter.(http.Flusher)
	if !ok {
		return
	}

	f.Flush()
}
