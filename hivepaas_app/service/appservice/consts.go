package appservice

import "github.com/hivepaas/hivepaas/services/docker"

const (
	LabelAppNamespace = docker.StackLabelNamespace
	LabelAppKey       = "hivepaas.app.key"
	LabelAppName      = "hivepaas.app.name"
	LabelAppEnv       = "hivepaas.app.env"
)
