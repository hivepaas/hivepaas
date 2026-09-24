package internal

import (
	"testing"
	"time"

	"go.uber.org/fx"
	"google.golang.org/grpc"

	agentserver "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/server"
	"github.com/hivepaas/hivepaas/hivepaas_app/registry"
)

// TestAgentFxGraphResolves is TestFxGraphResolves for the agent: it mirrors
// cmd/agent/main.go, and an fx.Invoke added there belongs here too.
func TestAgentFxGraphResolves(t *testing.T) {
	const startTimeout = 60 * time.Second

	provides := make([]any, 0, len(registry.Provides)+1)
	provides = append(provides, func(_ *agentserver.AgentServer) GrpcRegistrar {
		return func(*grpc.Server) {}
	})
	provides = append(provides, registry.Provides...)

	err := fx.ValidateApp(
		fx.StartTimeout(startTimeout),
		fx.Provide(provides...),
		fx.Invoke(InitLogger),
		fx.Invoke(InitConfig),
		fx.Invoke(InitDBConnection),
		fx.Invoke(InitCache),
		fx.Invoke(InitDockerManager),
		fx.Invoke(InitSystemSettings),
		fx.Invoke(InitSystemEventBus),
		fx.Invoke(InitGrpcServer),
		fx.Invoke(InitDockerAPIHost),
	)
	if err != nil {
		t.Fatalf("the agent's fx graph does not resolve: %v", err)
	}
}
