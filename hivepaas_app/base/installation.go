package base

type InstallationStep string

const (
	InstallationStepNone       = ""
	InstallationStepInitData   = "hivepaas/init-data"
	InstallationStepGetStarted = "hivepaas/get-started"
)
