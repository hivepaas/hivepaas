package projecthelper

import (
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/slugify"
)

func CalcProjectEnvKey(env string) string {
	env = strings.ToLower(env)
	switch env {
	case "dev", "development", "develop":
		return "dev"
	case "prod", "production":
		return "prod"
	case "stg", "staging":
		return "stg"
	case "test", "testing":
		return "test"
	default:
		return slugify.SlugifyEx(env, []string{"-", "_"}, projectKeyMaxLen)
	}
}

func CalcProjectEnvID(projectID, env string) string {
	if projectID == "" || env == "" {
		return ""
	}
	return projectID + ":" + CalcProjectEnvKey(env)
}

func ParseProjectEnvID(projectEnvID string) (string, string) {
	before, after, found := strings.Cut(projectEnvID, ":")
	if !found {
		return projectEnvID, ""
	}
	return before, after
}
