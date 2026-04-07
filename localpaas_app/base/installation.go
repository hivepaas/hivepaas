package base

type InstallationStep string

const (
	InstallationStepNone         = "none"
	InstallationStepInitData     = "localpaas/init-data"
	InstallationStepObtainAppSSL = "localpaas/obtain-ssl"
)
