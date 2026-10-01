package main

import (
	"time"

	"go.uber.org/fx"
	"google.golang.org/grpc"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/cmd/internal"
	agentproto "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/proto"
	agentserver "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/server"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/fileutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/registry"
)

const (
	startTimeoutDefault = 60 * time.Second
	stopTimeoutDefault  = 10 * time.Minute
)

func main() {
	provides := make([]any, 0, len(registry.Provides)+1)
	provides = append(provides,
		func(agentSrv *agentserver.AgentServer) internal.GrpcRegistrar {
			return func(s *grpc.Server) {
				agentproto.RegisterAgentServiceServer(s, agentSrv)
				agentproto.RegisterContainerServiceServer(s, agentSrv)
				agentproto.RegisterDockerAPIServiceServer(s, agentSrv)
				agentproto.RegisterNodeCleanupServiceServer(s, agentSrv)
				agentproto.RegisterImageBuildServiceServer(s, agentSrv)
				agentproto.RegisterNodeServiceServer(s, agentSrv)
				agentproto.RegisterRepoServerServiceServer(s, agentSrv)
				agentproto.RegisterVolumeServiceServer(s, agentSrv)
				agentproto.RegisterFileServiceServer(s, agentSrv)
				agentproto.RegisterFunctionServiceServer(s, agentSrv)
			}
		})
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
		fx.Invoke(sweepTempDirs),
		fx.Invoke(internal.InitGrpcServer),
		fx.Invoke(internal.InitDockerAPIHost),
	)

	app.Run()
}

// sweepTempDirs removes the temporary directories an agent killed during a
// build left behind. It runs before the agent serves: no build is running yet,
// so every day's directory is stale, today's included.
func sweepTempDirs() {
	_, _ = fileutil.RemoveDatedTempDirs(base.BaseTempDirDefault, time.Now())
}
