package report

import (
	"context"
	"fmt"

	reportv1 "github.com/hesoyamTM/apphelper-protos/gen/go/report"
	"github.com/hesoyamTM/apphelper-sso/internal/clients"
	"github.com/hesoyamTM/apphelper-sso/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	api reportv1.ReportClient
	log logger.Logger
}

func New(ctx context.Context, addr string) (*Client, error) {
	const op = "report.New"

	cc, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithDefaultServiceConfig(clients.RetryPolicy), grpc.WithMaxCallAttempts(10),
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Client{
		api: reportv1.NewReportClient(cc),
		log: *logger.GetLoggerFromCtx(ctx),
	}, nil
}
