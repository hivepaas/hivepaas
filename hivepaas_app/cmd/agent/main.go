package main

import (
	"context"
	"time"

	"go.uber.org/fx"
	"google.golang.org/grpc"

	"github.com/hivepaas/hivepaas/hivepaas_app/cmd/internal"
	agentproto "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/proto"
	agentserver "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/server"
	"github.com/hivepaas/hivepaas/hivepaas_app/registry"
)

const (
	startTimeoutDefault = 60 * time.Second
	stopTimeoutDefault  = 10 * time.Minute
)

func main() {
	provides := []any{
		context.Background,
		func(agentSrv *agentserver.AgentServer) internal.GrpcRegistrar {
			return func(s *grpc.Server) {
				agentproto.RegisterAgentServiceServer(s, agentSrv)
				agentproto.RegisterContainerServiceServer(s, agentSrv)
				agentproto.RegisterNodeCleanupServiceServer(s, agentSrv)
				agentproto.RegisterImageBuildServiceServer(s, agentSrv)
			}
		},
	}
	provides = append(provides, registry.Provides...)

	app := fx.New(
		fx.StartTimeout(startTimeoutDefault),
		fx.StopTimeout(stopTimeoutDefault),
		fx.Provide(provides...),
		fx.Invoke(internal.InitLogger),
		fx.Invoke(internal.InitConfig),
		fx.Invoke(internal.InitDBConnection),
		fx.Invoke(internal.InitCache),
		fx.Invoke(internal.InitDockerManager),
		fx.Invoke(internal.InitSystemSettings),
		fx.Invoke(internal.InitSystemEventBus),
		fx.Invoke(internal.InitGrpcServer),
	)

	app.Run()
}
