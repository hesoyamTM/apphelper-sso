package schedule

import (
	"context"
	"fmt"
	"log/slog"

	schedulev1 "github.com/hesoyamTM/apphelper-protos/gen/go/schedule"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	api schedulev1.ScheduleClient
	log *slog.Logger
}

func New(log *slog.Logger, addr string) (*Client, error) {
	const op = "schedule.New"

	cc, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("%s: %w")
	}

	return &Client{
		api: schedulev1.NewScheduleClient(cc),
		log: log,
	}, nil
}

func (c *Client) SetPublicKey(ctx context.Context, key string) error {
	const op = "schedule.SetPublicKey"

	_, err := c.api.SetPublicKey(ctx, &schedulev1.SetPublicKeyRequest{
		Key: key,
	})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
