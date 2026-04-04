package main

import (
	"context"

	"go.uber.org/fx"

	"github.com/localpaas/localpaas/localpaas_app/cmd/internal"
	"github.com/localpaas/localpaas/localpaas_app/registry"
)

func main() {
	provides := []any{
		context.Background,
	}
	provides = append(provides, registry.Provides...)

	app := fx.New(
		fx.Provide(provides...),
		fx.Invoke(internal.InitLogger),
		fx.Invoke(internal.InitConfig),
		fx.Invoke(internal.InitDBConnection),
		fx.Invoke(internal.MigrateData), // Migrate data structure of JSON columns
		fx.Invoke(internal.InitCache),
		fx.Invoke(internal.InitTaskQueue),
		fx.Invoke(internal.InitDockerManager),
		fx.Invoke(internal.InitSystemData),
		fx.Invoke(internal.InitJWTSession),
		fx.Invoke(internal.InitHTTPServer),
	)

	app.Run()
}
