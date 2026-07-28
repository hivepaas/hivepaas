package projecthelper

import (
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/slugify"
)

const (
	projectEnvIDSep = ":"
)

func CalcProjectEnvKey(env string) string {
	_, after, found := strings.Cut(env, projectEnvIDSep)
	if found {
		env = after
	}
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

func IsProjectEnvID(env string) bool {
	return strings.IndexByte(env, projectEnvIDSep[0]) > 0
}

func CalcProjectEnvID(projectID, env string) string {
	if IsProjectEnvID(env) {
		return env
	}
	if projectID == "" || env == "" {
		return ""
	}
	return projectID + projectEnvIDSep + CalcProjectEnvKey(env)
}

func ParseProjectEnvID(projectEnvID string) (string, string) {
	before, after, found := strings.Cut(projectEnvID, projectEnvIDSep)
	if !found {
		return "", ""
	}
	return before, after
}
