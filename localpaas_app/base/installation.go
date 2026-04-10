package base

type InstallationStep string

const (
	InstallationStepNone         = ""
	InstallationStepInitData     = "localpaas/init-data"
	InstallationStepObtainAppSSL = "localpaas/obtain-ssl"
)
