package metrics

import (
	"net/http"
	"strconv"
	"time"
)

type responseWriter struct {
	http.ResponseWriter

	statusCode int
}

func (m *Metrics) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rw := responseWriter{ResponseWriter: w}
			next.ServeHTTP(&rw, r)

			duration := time.Since(start).Seconds()
			statusCode := rw.statusCode
			if statusCode == 0 {
				statusCode = http.StatusOK
			}

			m.RequestCounter.WithLabelValues(r.Method, r.URL.Path, strconv.Itoa(statusCode)).Inc()
			m.RequestDuration.WithLabelValues(r.Method, r.URL.Path, strconv.Itoa(statusCode)).Observe(duration)
		})
	}
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
