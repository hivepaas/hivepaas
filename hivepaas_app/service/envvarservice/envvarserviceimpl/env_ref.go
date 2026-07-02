package envvarserviceimpl

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
)

var (
	reEnvOrSecretRef = regexp.MustCompile(`\$\{(secrets\.)?(\w)+\}`)
	// reSecretRef = regexp.MustCompile(`\$\{secrets\.(\w)+\}`)
)

func (s *service) processRefs(
	env *envvarservice.EnvVar,
	envStore map[string]*entity.EnvVar,
	secretStore map[string]*entity.Secret,
	usedSecrets *[]*entity.Secret,
) {
	processRefs(env, envStore, secretStore, usedSecrets, make(map[string]struct{}))
}

func processRefs(
	env *envvarservice.EnvVar,
	envStore map[string]*entity.EnvVar,
	secretStore map[string]*entity.Secret,
	usedSecrets *[]*entity.Secret,
	visitMap map[string]struct{},
) {
	if env.IsLiteral {
		return
	}

	replFunc := func(match string) string {
		envName, isSecret := parseEnvName(match) // env form: ${NAME} or ${secrets.NAME}
		if isSecret {
			refSecret, exists := secretStore[envName]
			if !exists {
				env.Errors = append(env.Errors, fmt.Sprintf("secret '%s' not found", envName))
				return match
			}
			value, err := refSecret.Value.GetPlain()
			if err != nil {
				env.Errors = append(env.Errors, fmt.Sprintf("failed to parse secret '%s'", envName))
				return match
			}
			*usedSecrets = append(*usedSecrets, refSecret)
			return value
		}

		// Prevent infinite loop due to circular references
		if _, exists := visitMap[envName]; exists {
			env.Errors = append(env.Errors, fmt.Sprintf("circular references detected at '%s'", envName))
			return match
		}
		visitMap[envName] = struct{}{}

		val, exists := envStore[envName]
		if !exists {
			env.Errors = append(env.Errors, fmt.Sprintf("env '%s' not found", envName))
			return match
		}
		refEnv := &envvarservice.EnvVar{EnvVar: val}
		processRefs(refEnv, envStore, secretStore, usedSecrets, visitMap)
		if len(refEnv.Errors) > 0 {
			env.Errors = append(env.Errors, refEnv.Errors...)
			return match
		}
		return refEnv.Value
	}

	env.Value = reEnvOrSecretRef.ReplaceAllStringFunc(env.Value, replFunc)
}

func parseEnvName(match string) (string, bool) {
	envName := match[2 : len(match)-1]
	if after, ok := strings.CutPrefix(envName, "secrets."); ok {
		return after, true
	}
	return envName, false
}
