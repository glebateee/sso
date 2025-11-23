package suite

import (
	"context"
	"net"
	"sso/internal/config"
	"strconv"
	"testing"

	ssov1 "github.com/glebateee/protos/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Suite struct {
	*testing.T
	Cfg        *config.Config
	AuthClient ssov1.AuthClient
}

const (
	gRPCHost = "localhost"
)

func New(t *testing.T) (context.Context, *Suite) {
	t.Helper()
	t.Parallel()

	cfg := config.MustLoadByPath("../config/local.yaml")
	ctx, cancel := context.WithTimeout(context.Background(), cfg.GRPC.Timeout)

	t.Cleanup(func() {
		t.Helper()
		cancel()
	})
	client, err := grpc.NewClient(grpcAddress(cfg), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("grpc server coonection failed: %v", err)
	}
	return ctx, &Suite{
		T:          t,
		Cfg:        cfg,
		AuthClient: ssov1.NewAuthClient(client),
	}
}

func grpcAddress(cfg *config.Config) string {
	return net.JoinHostPort(gRPCHost, strconv.Itoa(cfg.GRPC.Port))
}
