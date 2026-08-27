package envvarserviceimpl

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
)

func init() {
	if config.Current == nil {
		config.Current = &config.Config{Secret: "test_secret_key"}
	}
}

func newEnvVar(key, value string, isShared, isLiteral bool) *envvarservice.EnvVar {
	return &envvarservice.EnvVar{
		EnvVar: &entity.EnvVar{
			Key:       key,
			Value:     value,
			IsShared:  isShared,
			IsLiteral: isLiteral,
		},
	}
}

func createSecretSetting(key, plainVal string) *entity.Setting {
	return &entity.Setting{
		Type: base.SettingTypeSecret,
		Data: fmt.Sprintf(`{"key":%q,"value":%q}`, key, plainVal),
	}
}

func TestProcessRefs_Simple(t *testing.T) {
	s := &service{}

	envBase := newEnvVar("BASE_URL", "https://example.com", false, false)
	envApi := newEnvVar("API_URL", "${BASE_URL}/api/v1", false, false)

	data := &processRefsData{
		EnvStore: map[string]*envvarservice.EnvVar{
			"BASE_URL": envBase,
			"API_URL":  envApi,
		},
	}

	err := s.processRefs(envApi, data)
	assert.NoError(t, err)
	assert.Equal(t, "https://example.com/api/v1", envApi.Value)
	assert.Empty(t, envApi.Errors)
}

func TestProcessRefs_MultiStepRecursive(t *testing.T) {
	s := &service{}

	envPort := newEnvVar("PORT", "3000", false, false)
	envHost := newEnvVar("HOST", "localhost:${PORT}", false, false)
	envUrl := newEnvVar("URL", "http://${HOST}", false, false)

	data := &processRefsData{
		EnvStore: map[string]*envvarservice.EnvVar{
			"PORT": envPort,
			"HOST": envHost,
			"URL":  envUrl,
		},
	}

	err := s.processRefs(envUrl, data)
	assert.NoError(t, err)
	assert.Equal(t, "http://localhost:3000", envUrl.Value)
	assert.Equal(t, "localhost:3000", envHost.Value)
}

func TestProcessRefs_DuplicateInSameString(t *testing.T) {
	s := &service{}

	envHost := newEnvVar("HOST", "api.internal", false, false)
	envUrl := newEnvVar("URL", "https://${HOST}:8080/${HOST}", false, false)

	data := &processRefsData{
		EnvStore: map[string]*envvarservice.EnvVar{
			"HOST": envHost,
			"URL":  envUrl,
		},
	}

	err := s.processRefs(envUrl, data)
	assert.NoError(t, err)
	assert.Equal(t, "https://api.internal:8080/api.internal", envUrl.Value)
}

func TestProcessRefs_DiamondDependency(t *testing.T) {
	s := &service{}

	envRoot := newEnvVar("ROOT", "core", false, false)
	envB := newEnvVar("VAR_B", "${ROOT}_b", false, false)
	envC := newEnvVar("VAR_C", "${ROOT}_c", false, false)
	envA := newEnvVar("VAR_A", "${VAR_B} and ${VAR_C}", false, false)

	data := &processRefsData{
		EnvStore: map[string]*envvarservice.EnvVar{
			"ROOT":  envRoot,
			"VAR_B": envB,
			"VAR_C": envC,
			"VAR_A": envA,
		},
	}

	err := s.processRefs(envA, data)
	assert.NoError(t, err)
	assert.Equal(t, "core_b and core_c", envA.Value)
}

func TestProcessRefs_CircularReference(t *testing.T) {
	s := &service{}

	envA := newEnvVar("A", "prefix_${B}", false, false)
	envB := newEnvVar("B", "suffix_${A}", false, false)

	data := &processRefsData{
		EnvStore: map[string]*envvarservice.EnvVar{
			"A": envA,
			"B": envB,
		},
	}

	err := s.processRefs(envA, data)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, hperrors.ErrEnvVarCircularReference))
}

func TestProcessRefs_SelfReference(t *testing.T) {
	s := &service{}

	envA := newEnvVar("A", "prefix_${A}", false, false)

	data := &processRefsData{
		EnvStore: map[string]*envvarservice.EnvVar{
			"A": envA,
		},
	}

	err := s.processRefs(envA, data)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, hperrors.ErrEnvVarCircularReference))
}

func TestProcessRefs_MissingVar(t *testing.T) {
	s := &service{}

	envA := newEnvVar("A", "hello_${MISSING_VAR}", false, false)

	data := &processRefsData{
		EnvStore: map[string]*envvarservice.EnvVar{
			"A": envA,
		},
	}

	err := s.processRefs(envA, data)
	assert.NoError(t, err)
	assert.Equal(t, "hello_${MISSING_VAR}", envA.Value)
	if assert.Len(t, envA.Errors, 1) {
		assert.Equal(t, envvarservice.ParseErrorVarMissing, envA.Errors[0].Type)
		assert.Equal(t, "MISSING_VAR", envA.Errors[0].Name)
	}
}

func TestProcessRefs_Secrets(t *testing.T) {
	s := &service{}

	secretSetting := createSecretSetting("DB_PASSWORD", "super_secret_123")
	envDbUrl := newEnvVar("DATABASE_URL", "postgres://user:${secrets.DB_PASSWORD}@localhost:5432/db", false, false)

	data := &processRefsData{
		EnvStore: map[string]*envvarservice.EnvVar{
			"DATABASE_URL": envDbUrl,
		},
		SecretStore: map[string]*entity.Setting{
			"DB_PASSWORD": secretSetting,
		},
	}

	// 1. Unmasked resolution
	err := s.processRefs(envDbUrl, data)
	assert.NoError(t, err)
	assert.Equal(t, "postgres://user:super_secret_123@localhost:5432/db", envDbUrl.Value)
	assert.Len(t, envDbUrl.RefSecrets, 1)

	// 2. Masked resolution
	envMasked := newEnvVar("DATABASE_URL", "postgres://user:${secrets.DB_PASSWORD}@localhost:5432/db", false, false)
	dataMasked := &processRefsData{
		EnvStore: map[string]*envvarservice.EnvVar{
			"DATABASE_URL": envMasked,
		},
		SecretStore: map[string]*entity.Setting{
			"DB_PASSWORD": secretSetting,
		},
		BuildOptions: envvarservice.EnvBuildOptions{MaskSecrets: true},
	}
	err = s.processRefs(envMasked, dataMasked)
	assert.NoError(t, err)
	assert.Equal(t, "postgres://user:********@localhost:5432/db", envMasked.Value)

	// 3. Missing secret
	envMissingSecret := newEnvVar("KEY", "${secrets.NON_EXISTING}", false, false)
	dataMissing := &processRefsData{
		EnvStore: map[string]*envvarservice.EnvVar{
			"KEY": envMissingSecret,
		},
		SecretStore: map[string]*entity.Setting{},
	}
	err = s.processRefs(envMissingSecret, dataMissing)
	assert.NoError(t, err)
	assert.Equal(t, "${secrets.NON_EXISTING}", envMissingSecret.Value)
	if assert.Len(t, envMissingSecret.Errors, 1) {
		assert.Equal(t, envvarservice.ParseErrorSecretMissing, envMissingSecret.Errors[0].Type)
	}
}

func TestProcessRefs_ExternalRefs(t *testing.T) {
	s := &service{}

	envApp := newEnvVar("BACKEND_URL", "http://${db_service.HOST}:${db_service.PORT}", false, false)

	data := &processRefsData{
		EnvStore: map[string]*envvarservice.EnvVar{
			"BACKEND_URL": envApp,
		},
		ExternalRefsLoadFunc: func(refName string) (map[string]*envvarservice.EnvVar, error) {
			if refName == "db_service" {
				return map[string]*envvarservice.EnvVar{
					"HOST": newEnvVar("HOST", "db.local", false, false),
					"PORT": newEnvVar("PORT", "5432", false, false),
				}, nil
			}
			return nil, nil
		},
	}

	err := s.processRefs(envApp, data)
	assert.NoError(t, err)
	assert.Equal(t, "http://db.local:5432", envApp.Value)

	// Shared env containing external ref should return error
	envShared := newEnvVar("SHARED_VAR", "${db_service.HOST}", true, false)
	dataShared := &processRefsData{
		EnvStore: map[string]*envvarservice.EnvVar{
			"SHARED_VAR": envShared,
		},
		ExternalRefsLoadFunc: data.ExternalRefsLoadFunc,
	}
	err = s.processRefs(envShared, dataShared)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, hperrors.ErrSharedEnvVarContainExternalReference))
}

func TestProcessRefs_LiteralAndNoRef(t *testing.T) {
	s := &service{}

	// Literal flag true
	envLiteral := newEnvVar("LITERAL", "${DO_NOT_RESOLVE}", false, true)
	data := &processRefsData{
		EnvStore: map[string]*envvarservice.EnvVar{
			"LITERAL": envLiteral,
		},
	}
	err := s.processRefs(envLiteral, data)
	assert.NoError(t, err)
	assert.Equal(t, "${DO_NOT_RESOLVE}", envLiteral.Value)

	// No reference
	envPlain := newEnvVar("PLAIN", "just_plain_text", false, false)
	dataPlain := &processRefsData{
		EnvStore: map[string]*envvarservice.EnvVar{
			"PLAIN": envPlain,
		},
	}
	err = s.processRefs(envPlain, dataPlain)
	assert.NoError(t, err)
	assert.Equal(t, "just_plain_text", envPlain.Value)
}
