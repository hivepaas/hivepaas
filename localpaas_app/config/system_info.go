package config

import "github.com/localpaas/localpaas/localpaas_app/base"

type SystemInfo struct {
	NextStep base.InstallationStep `toml:"-" env:"-"`
}
