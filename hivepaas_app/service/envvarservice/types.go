package envvarservice

import "github.com/hivepaas/hivepaas/hivepaas_app/entity"

type EnvVar struct {
	*entity.EnvVar
	Errors []string
}

func (env *EnvVar) ToString(sep string) string {
	return env.Key + sep + env.Value
}
