package server

import (
	"context"
	"fmt"
	"net"

	"github.com/devathh/staffy-sso/internal/infrastructure/config"
	"github.com/devathh/staffy-sso/pkg/consts"
	"google.golang.org/grpc"
)

type Server struct {
	grpcServer *grpc.Server
	cfg        *config.Config
}

func (s *Server) Start() error {
	lis, err := net.Listen(
		s.cfg.Server.GRPC.Protocol,
		net.JoinHostPort(
			s.cfg.Server.GRPC.Host,
			s.cfg.Server.GRPC.Port,
		),
	)

	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}

	if err := s.grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve listener: %w", err)
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	stopped := make(chan struct{})
	go func() {
		s.grpcServer.GracefulStop()
		close(stopped)
	}()

	select {
	case <-ctx.Done():
		s.grpcServer.Stop()
		return ctx.Err()
	case <-stopped:
		return nil
	}
}

func NewServer(cfg *config.Config, grpcServer *grpc.Server) (*Server, error) {
	if cfg == nil {
		return nil, consts.ErrNilCfg
	}

	return &Server{
		grpcServer: grpcServer,
		cfg:        cfg,
	}, nil
}
