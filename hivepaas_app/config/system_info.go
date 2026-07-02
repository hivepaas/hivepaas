package config

import "github.com/hivepaas/hivepaas/hivepaas_app/base"

type SystemInfo struct {
	NextStep base.InstallationStep `toml:"-" env:"-"`
}
