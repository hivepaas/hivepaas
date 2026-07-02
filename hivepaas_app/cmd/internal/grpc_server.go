package internal

import (
	"context"
	"net"
	"os"
	"strconv"
	"time"

	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/logging"
)

const (
	maxConnectionIdle     = 15 * time.Minute
	maxConnectionAge      = 30 * time.Minute
	maxConnectionAgeGrace = 5 * time.Second
	keepaliveTime         = 1 * time.Minute
	keepaliveTimeout      = 20 * time.Second
	keepaliveMinTime      = 5 * time.Second
)

// GrpcRegistrar defines a callback type for registering gRPC services.
type GrpcRegistrar func(s *grpc.Server)

func InitGrpcServer(
	lc fx.Lifecycle,
	cfg *config.Config,
	logger logging.Logger,
	registerFn GrpcRegistrar,
) {
	// 1. Setup keepalive parameters to maintain healthy long-lived connections
	keepaliveParams := keepalive.ServerParameters{
		MaxConnectionIdle:     maxConnectionIdle, // Close connections if idle
		MaxConnectionAge:      maxConnectionAge,  // Helps naturally load balance clients
		MaxConnectionAgeGrace: maxConnectionAgeGrace,
		Time:                  keepaliveTime,    // Ping client after 1 minute of silence
		Timeout:               keepaliveTimeout, // Wait 20s for client response
	}

	keepalivePolicy := keepalive.EnforcementPolicy{
		MinTime:             keepaliveMinTime, // Minimum duration between client pings
		PermitWithoutStream: true,             // Allow pings even if no active RPCs
	}

	// 2. Setup server options including keepalives and chained custom interceptors
	opts := []grpc.ServerOption{
		grpc.KeepaliveParams(keepaliveParams),
		grpc.KeepaliveEnforcementPolicy(keepalivePolicy),
		grpc.ChainUnaryInterceptor(
			unaryAuthInterceptor(),
			unaryLoggingAndRecoveryInterceptor(logger),
		),
		grpc.ChainStreamInterceptor(
			streamAuthInterceptor(),
		),
	}

	server := grpc.NewServer(opts...)

	// Invoke the register function to register services
	registerFn(server)

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			grpcPort := strconv.Itoa(cfg.Agent.Port)
			lis, err := net.Listen("tcp", ":"+grpcPort)
			if err != nil {
				return apperrors.New(err)
			}
			logger.Infof("gRPC Server listening on port %s ...", grpcPort)
			go func() {
				if err := server.Serve(lis); err != nil {
					logger.Errorf("gRPC Server stopped: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("stopping gRPC Server ...")
			server.GracefulStop()
			return nil
		},
	})
}

// Unary interceptor for logging request metadata and recovering from potential panics
func unaryLoggingAndRecoveryInterceptor(logger logging.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {
		startTime := time.Now()

		// Protect the server from crashing due to unexpected panics
		defer func() {
			if r := recover(); r != nil {
				logger.Errorf("[gRPC Panic] Recovered from panic in method %s: %v", info.FullMethod, r)
				err = status.Errorf(codes.Internal, "internal server error: panic recovered")
			}
		}()

		logger.Infof("[gRPC Request] Start - Method: %s", info.FullMethod)
		resp, err = handler(ctx, req)

		duration := time.Since(startTime)
		if err != nil {
			err = apperrors.ToGRPCError(err)
			logger.Errorf("[gRPC Request] Fail - Method: %s, Duration: %s, Error: %v", info.FullMethod, duration, err)
		} else {
			logger.Infof("[gRPC Request] Success - Method: %s, Duration: %s", info.FullMethod, duration)
		}

		return resp, apperrors.New(err)
	}
}

func validateAuthToken(ctx context.Context) error {
	secretToken := os.Getenv("AGENT_SECRET_TOKEN")
	if secretToken == "" {
		return nil // No token configured, skip authentication (for local dev)
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Errorf(codes.Unauthenticated, "missing metadata")
	}

	tokens := md.Get("authorization")
	if len(tokens) == 0 {
		return status.Errorf(codes.Unauthenticated, "missing authorization token")
	}

	if tokens[0] != secretToken {
		return status.Errorf(codes.Unauthenticated, "invalid authorization token")
	}

	return nil
}

func unaryAuthInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		if err := validateAuthToken(ctx); err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

func streamAuthInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv any,
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		if err := validateAuthToken(ss.Context()); err != nil {
			return err
		}
		return handler(srv, ss)
	}
}
