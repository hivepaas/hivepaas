package appservice

import "github.com/hivepaas/hivepaas/services/docker"

const (
	LabelAppNamespace = docker.StackLabelNamespace
	LabelAppInfo      = "hivepaas.app.info"
	// LabelAppPrevServiceMode holds the service mode captured when an app is stopped, so that
	// starting it again restores the previous mode instead of guessing one.
	LabelAppPrevServiceMode = "hivepaas.app.prevServiceMode"
)
