package envvarserviceimpl

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
)

const (
	secretRefPrefix = "secrets"
	secretMask      = "********"
)

var (
	reEnvOrSecretRef = regexp.MustCompile(`\$\{(?:([a-zA-Z0-9_]+)\.)?([a-zA-Z_][a-zA-Z0-9_]*)\}`)
)

type processRefsData struct {
	EnvStore    map[string]*envvarservice.EnvVar
	SecretStore map[string]*entity.Setting
	MaskSecrets bool

	ExternalRefsData     map[string]map[string]*envvarservice.EnvVar
	ExternalRefsLoadFunc func(refName string) (map[string]*envvarservice.EnvVar, error)
}

func (s *service) processRefs(
	env *envvarservice.EnvVar,
	data *processRefsData,
) error {
	return s.processRefsRecursively(env, data, "", nil)
}

//nolint:gocognit
func (s *service) processRefsRecursively(
	env *envvarservice.EnvVar,
	data *processRefsData,
	sharedEnv string,
	visitingMap map[string]struct{},
) (gErr error) {
	// Trivial case: env is literal or no ref in the value
	if env.IsLiteral || !s.HasRef(env.Value) {
		return nil
	}

	if env.IsShared {
		sharedEnv = env.Key
	}
	if visitingMap == nil {
		visitingMap = make(map[string]struct{})
	}

	replFunc := func(match string) string {
		if gErr != nil {
			return match
		}
		refName, varName := parseEnvRef(match) // env form: ${VAR} or ${secrets.NAME} or ${an_app.VAR}
		if refName == secretRefPrefix {
			refSecret, exists := data.SecretStore[varName]
			if !exists {
				env.Errors = append(env.Errors, fmt.Sprintf("secret '%s' not found", varName))
				return match
			}
			secret, err := refSecret.AsSecret()
			if err != nil {
				env.Errors = append(env.Errors, fmt.Sprintf("failed to parse secret '%s'", varName))
				return match
			}
			env.RefSecrets[secret] = struct{}{}

			if data.MaskSecrets {
				return secretMask
			}
			value, err := secret.Value.GetPlain()
			if err != nil {
				env.Errors = append(env.Errors, fmt.Sprintf("failed to decrypt secret '%s'", varName))
				return match
			}
			return value
		}

		if refName != "" { // external ref: e.g. reference another app env such as `db.HOST`
			if sharedEnv != "" {
				gErr = apperrors.Wrap(apperrors.ErrSharedEnvVarContainExternalReference).
					WithParam("Name", sharedEnv)
				return match
			}
			if data.ExternalRefsLoadFunc == nil {
				gErr = apperrors.Wrap(apperrors.ErrEnvVarExternalReferenceIsNotAllowed).
					WithParam("Name", env.Key)
				return match
			}
			externalVars, err := s.loadExternalRefEnvData(refName, data)
			if err != nil {
				gErr = apperrors.Wrap(err)
				return match
			}
			val, exists := externalVars[varName]
			if !exists {
				env.Errors = append(env.Errors, fmt.Sprintf("env '%s.%s' not found", refName, varName))
				return match
			}
			return val.Value
		}

		// Prevent infinite loop due to circular references
		if _, exists := visitingMap[varName]; exists {
			env.Errors = append(env.Errors, fmt.Sprintf("circular references detected at '%s'", varName))
			return match
		}
		visitingMap[varName] = struct{}{}

		refEnv, exists := data.EnvStore[varName]
		if !exists {
			env.Errors = append(env.Errors, fmt.Sprintf("env '%s' not found", varName))
			return match
		}
		if err := s.processRefsRecursively(refEnv, data, sharedEnv, visitingMap); err != nil {
			gErr = apperrors.Wrap(err)
			return ""
		}
		if len(refEnv.Errors) > 0 {
			env.Errors = append(env.Errors, refEnv.Errors...)
			return match
		}
		for secret := range refEnv.RefSecrets {
			env.RefSecrets[secret] = struct{}{}
		}
		return refEnv.Value
	}

	value := reEnvOrSecretRef.ReplaceAllStringFunc(env.Value, replFunc)
	if gErr != nil {
		return apperrors.Wrap(gErr)
	}
	env.Value = value
	return nil
}

func parseEnvRef(match string) (refName, envName string) {
	s := match[2 : len(match)-1]
	refName, envName, found := strings.Cut(s, ".")
	if !found {
		return "", s
	}
	return refName, envName
}

func (s *service) loadExternalRefEnvData(
	appKey string,
	data *processRefsData,
) (map[string]*envvarservice.EnvVar, error) {
	if appData, ok := data.ExternalRefsData[appKey]; ok {
		return appData, nil
	}
	refData, err := data.ExternalRefsLoadFunc(appKey)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	data.ExternalRefsData[appKey] = refData
	return refData, nil
}
