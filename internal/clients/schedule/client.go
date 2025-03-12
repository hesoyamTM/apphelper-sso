package schedule

import (
	"context"
	"fmt"

	schedulev1 "github.com/hesoyamTM/apphelper-protos/gen/go/schedule"
	"github.com/hesoyamTM/apphelper-sso/internal/clients"
	"github.com/hesoyamTM/apphelper-sso/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	api schedulev1.ScheduleClient
	log *logger.Logger
}

func New(ctx context.Context, addr string) (*Client, error) {
	const op = "schedule.New"

	cc, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithDefaultServiceConfig(clients.RetryPolicy), grpc.WithMaxCallAttempts(10))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Client{
		api: schedulev1.NewScheduleClient(cc),
		log: logger.GetLoggerFromCtx(ctx),
	}, nil
}
