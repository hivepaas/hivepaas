package envhelper

import "github.com/hivepaas/hivepaas/hivepaas_app/entity"

func ToMap(envVars []*entity.EnvVar) map[string]*entity.EnvVar {
	result := make(map[string]*entity.EnvVar, len(envVars))
	for _, envVar := range envVars {
		result[envVar.Key] = envVar
	}
	return result
}
