package main

import (
	"context"
	"time"

	"go.uber.org/fx"

	"github.com/hivepaas/hivepaas/hivepaas_app/cmd/internal"
	"github.com/hivepaas/hivepaas/hivepaas_app/registry"
)

const (
	startTimeoutDefault = 60 * time.Second
)

func main() {
	provides := []any{
		context.Background,
	}
	provides = append(provides, registry.Provides...)

	app := fx.New(
		fx.StartTimeout(startTimeoutDefault),
		fx.Provide(provides...),
		fx.Invoke(internal.InitLogger),
		fx.Invoke(internal.InitConfig),
		fx.Invoke(internal.InitDBConnection),
		fx.Invoke(internal.InitCache),
		fx.Invoke(internal.InitDockerManager),
		fx.Invoke(internal.CompleteInstallation),
		fx.Invoke(internal.InitTaskQueue),
		fx.Invoke(internal.InitJWTSession),
		fx.Invoke(internal.InitHTTPServer),
		fx.Invoke(internal.InitUpdater),
		fx.Invoke(internal.FinalizeStartup),
	)

	app.Run()
}
