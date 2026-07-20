package projecthelper

import (
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/slugify"
)

const (
	projectKeyMaxLen = 100
)

func CalcProjectKey(projectName string) string {
	return slugify.SlugifyEx(projectName, []string{"-", "_"}, projectKeyMaxLen)
}

func CalcProjectEnvKey(env string) string {
	env = strings.ToLower(env)
	switch env {
	case "development":
		return "dev"
	case "production":
		return "prod"
	default:
		return slugify.SlugifyEx(env, []string{"-", "_"}, projectKeyMaxLen)
	}
}

func CalcAppGlobalKey(projectKey, appKey, env string) string {
	globalKey := projectKey
	if env != "" {
		globalKey += "_" + CalcProjectEnvKey(env)
	}
	globalKey += "_" + appKey
	return globalKey
}
