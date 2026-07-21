// Command grpc is the gRPC entry point.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/zhanbinb/go-clean-arch_demo/cmd"
	grpchandler "github.com/zhanbinb/go-clean-arch_demo/internal/interfaces/grpc"
	"github.com/zhanbinb/go-clean-arch_demo/pkg/server"
)

func main() {
	env := os.Getenv("APP_ENV")
	deps, err := wire.New(context.Background(), env)
	if err != nil {
		log.Fatalf("init: %v", err)
	}
	defer func() { _ = deps.Log.Sync() }()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	grpcSrv, err := server.NewGRPCServer(deps.Cfg.Server.GRPCAddr(), nil, deps.Log.Zap())
	if err != nil {
		deps.Log.Fatal("build grpc server", zap.Error(err))
	}

	gh := grpchandler.NewHandlers(deps.ArticleSvc)
	gh.Register(grpcSrv.Server())

	if err := grpcSrv.Run(ctx); err != nil {
		deps.Log.Fatal("grpc server exited", zap.Error(err))
	}
	deps.Log.Info("grpc server stopped")
}
