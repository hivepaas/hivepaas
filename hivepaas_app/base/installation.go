package base

type InstallationStep string

const (
	InstallationStepNone         = ""
	InstallationStepInitData     = "hivepaas/init-data"
	InstallationStepObtainAppSSL = "hivepaas/obtain-ssl"
)
