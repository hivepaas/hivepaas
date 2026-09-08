package internal

import (
	"testing"
	"time"

	"go.uber.org/fx"

	"github.com/hivepaas/hivepaas/hivepaas_app/registry"
)

// TestFxGraphResolves builds the app's dependency graph without starting it.
//
// fx resolves constructors lazily at run time, so a provider left out of
// registry.Provides is not a compile error and not a test failure anywhere else:
// it is a process that dies on boot, or - worse, for anything constructed behind
// an interface - a request that panics much later. This is the cheap check that
// every constructor the startup path asks for can actually be built.
//
// It mirrors cmd/app/main.go. An fx.Invoke added there belongs here too, or the
// dependencies it drags in go unchecked.
func TestFxGraphResolves(t *testing.T) {
	const startTimeout = 60 * time.Second

	err := fx.ValidateApp(
		fx.StartTimeout(startTimeout),
		fx.Provide(registry.Provides...),
		fx.Invoke(InitLogger),
		fx.Invoke(InitConfig),
		fx.Invoke(InitDBConnection),
		fx.Invoke(InitDataKey),
		fx.Invoke(InitCache),
		fx.Invoke(InitDockerManager),
		fx.Invoke(SystemInstallation),
		fx.Invoke(InitSystemSettings),
		fx.Invoke(InitSystemEventBus),
		fx.Invoke(InitTaskQueue),
		fx.Invoke(InitWorkerHeartbeat),
		fx.Invoke(InitJWTSession),
		fx.Invoke(InitSettingsProbation),
		fx.Invoke(InitHTTPServer),
		fx.Invoke(InitUpdater),
		fx.Invoke(FinalizeStartup),
	)
	if err != nil {
		t.Fatalf("the app's fx graph does not resolve: %v", err)
	}
}
