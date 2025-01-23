package suite

import (
	"context"
	"net"
	"sso/internal/config"
	"strconv"
	"testing"
	"time"

	ssov1 "github.com/hesoyamTM/apphelper-protos/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Suite struct {
	*testing.T
	Cfg        *config.Config
	AuthClient ssov1.AuthClient
}

func New(t *testing.T) (context.Context, *Suite) {
	t.Helper()
	t.Parallel()

	cfg := config.MustLoadByPath("./config/local_test.yaml")

	ctx, cancelCtx := context.WithTimeout(context.Background(), time.Second*30)

	t.Cleanup(func() {
		t.Helper()
		cancelCtx()
	})

	cc, err := grpc.NewClient(net.JoinHostPort("localhost", strconv.Itoa(cfg.Grpc.Port)), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("grpc server connection failed: %v", err)
	}

	return ctx, &Suite{
		T:          t,
		Cfg:        cfg,
		AuthClient: ssov1.NewAuthClient(cc),
	}
}
