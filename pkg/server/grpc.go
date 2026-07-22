package server

import (
	"context"
	"errors"
	"net"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// GRPCServer wraps grpc.Server with graceful shutdown.
type GRPCServer struct {
	srv *grpc.Server
	lis net.Listener
	log *zap.Logger
}

// NewGRPCServer builds a gRPC server. Pass reflection.Register to enable
// service reflection (so grpcurl / grpcui can discover services).
func NewGRPCServer(addr string, opts []grpc.ServerOption, log *zap.Logger) (*GRPCServer, error) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	srv := grpc.NewServer(opts...)
	reflection.Register(srv) // enable grpcurl/grpcui debugging
	return &GRPCServer{srv: srv, lis: lis, log: log}, nil
}

// Server exposes the underlying grpc.Server for service registration.
func (g *GRPCServer) Server() *grpc.Server { return g.srv }

// Addr returns the actual bound address (useful when port = 0).
func (g *GRPCServer) Addr() string { return g.lis.Addr().String() }

// Run blocks until the server stops or ctx is cancelled.
func (g *GRPCServer) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		g.log.Info("grpc server starting", zap.String("addr", g.lis.Addr().String()))
		errCh <- g.srv.Serve(g.lis)
	}()

	select {
	case err := <-errCh:
		if errors.Is(err, grpc.ErrServerStopped) {
			return nil
		}
		return err
	case <-ctx.Done():
		g.log.Info("grpc server graceful stop initiated")
		stopped := make(chan struct{})
		go func() {
			g.srv.GracefulStop()
			close(stopped)
		}()
		select {
		case <-stopped:
			g.log.Info("grpc server stopped")
			return nil
		case <-time.After(30 * time.Second):
			g.log.Warn("grpc graceful stop timeout, forcing")
			g.srv.Stop()
			return nil
		}
	}
}
