package appservice

import "github.com/hivepaas/hivepaas/services/docker"

const (
	LabelAppNamespace = docker.StackLabelNamespace
	LabelAppName      = "hivepaas.app.name"
	LabelAppEnv       = "hivepaas.app.env"
)
