package report

import (
	"context"
	"fmt"
	"log/slog"

	reportv1 "github.com/hesoyamTM/apphelper-protos/gen/go/report"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	api reportv1.ResultsClient
	log *slog.Logger
}

func New(log *slog.Logger, addr string) (*Client, error) {
	const op = "report.New"

	cc, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Client{
		api: reportv1.NewResultsClient(cc),
		log: log,
	}, nil
}

func (c *Client) SetPublicKey(ctx context.Context, key string) error {
	const op = "report.SetPublicKey"

	_, err := c.api.SetPublicKey(ctx, &reportv1.SetPublicKeyRequest{
		Key: key,
	})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
