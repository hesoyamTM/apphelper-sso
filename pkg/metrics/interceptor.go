package metrics

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"google.golang.org/grpc"
)

func (m *Metrics) Interceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		resp, err := handler(ctx, req)
		status := http.StatusOK
		if err != nil {
			status = http.StatusInternalServerError
		}

		m.RequestCounter.WithLabelValues(info.FullMethod, info.FullMethod, strconv.Itoa(status)).Inc()
		m.RequestDuration.WithLabelValues(info.FullMethod, info.FullMethod, strconv.Itoa(status)).Observe(time.Since(start).Seconds())

		return resp, err
	}
}
