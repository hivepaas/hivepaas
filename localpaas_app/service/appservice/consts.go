package appservice

import "github.com/localpaas/localpaas/services/docker"

const (
	LabelAppNamespace = docker.StackLabelNamespace
	LabelAppName      = "localpaas.app.name"
	LabelAppEnv       = "localpaas.app.env"
)
