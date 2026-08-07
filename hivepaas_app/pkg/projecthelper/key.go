package projecthelper

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/slugify"
)

const (
	objectKeyMaxLen = 100
)

func CalcProjectKey(projectName string) string {
	return slugify.SlugifyEx(projectName, []string{"-", "_"}, objectKeyMaxLen)
}

func CalcAppKey(appName string) string {
	return slugify.SlugifyEx(appName, []string{"-", "_"}, objectKeyMaxLen)
}

func CalcAppGlobalKey(projectKey, appKey, env string) string {
	globalKey := projectKey
	if env != "" {
		globalKey += "_" + CalcProjectEnvKey(env)
	}
	globalKey += "_" + appKey
	return globalKey
}
