package envvarserviceimpl

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
)

func scopedSecret(id, key, value string, inheritable bool) *entity.Setting {
	return &entity.Setting{
		ID: id, Type: base.SettingTypeSecret, Name: key, Inheritable: inheritable,
		Data: fmt.Sprintf(`{"key":%q,"value":%q}`, key, value),
	}
}

// buildEnvUnderProject builds a project and then its env the way a deployment
// does: the env inherits what the project build returns.
func buildEnvUnderProject(
	t *testing.T, projectVars []*envvarservice.EnvVar, projectSecrets []*entity.Setting,
	envVars []*envvarservice.EnvVar, envSecrets []*entity.Setting,
) (*envvarservice.BuildEnvVarsInProjectResp, *envvarservice.BuildEnvVarsInProjectEnvResp) {
	t.Helper()
	s := &service{}
	project := &entity.Project{ID: "p1", Name: "p1"}
	projectResp, err := s.BuildEnvVarsInProject(context.Background(), nil, &envvarservice.BuildEnvVarsInProjectReq{
		Project:      project,
		DataLoadFunc: envvarservice.NewStaticEnvLoadFunc(projectVars, projectSecrets),
	})
	assert.NoError(t, err)
	envResp, err := s.BuildEnvVarsInProjectEnv(context.Background(), nil, &envvarservice.BuildEnvVarsInProjectEnvReq{
		ProjectEnv:            &entity.ProjectEnv{ID: "p1:dev", ProjectID: "p1", Project: project},
		DataLoadFunc:          envvarservice.NewStaticEnvLoadFunc(envVars, envSecrets),
		InheritedDataLoadFunc: envvarservice.NewStaticEnvLoadFunc(projectResp.EnvVars, projectResp.Secrets),
	})
	assert.NoError(t, err)
	return projectResp, envResp
}

func varNamed(vars []*envvarservice.EnvVar, key string) *envvarservice.EnvVar {
	for _, v := range vars {
		if v.Key == key {
			return v
		}
	}
	return nil
}

func errorTypes(v *envvarservice.EnvVar) []envvarservice.ParseErrorType {
	var types []envvarservice.ParseErrorType
	for _, err := range v.Errors {
		types = append(types, err.Type)
	}
	return types
}

// A secret the scope above does not make inheritable is used where it is
// defined, and is not there for the scope below.
func TestASecretThatIsNotInheritableStaysInItsScope(t *testing.T) {
	projectResp, envResp := buildEnvUnderProject(t,
		[]*envvarservice.EnvVar{newEnvVar("AT_PROJECT", "${secrets.HIDDEN}", false, false)},
		[]*entity.Setting{
			scopedSecret("s1", "HIDDEN", "h1dden", false),
			scopedSecret("s2", "SHARED", "sh4red", true),
		},
		[]*envvarservice.EnvVar{
			newEnvVar("USES_SHARED", "${secrets.SHARED}", false, false),
			newEnvVar("USES_HIDDEN", "${secrets.HIDDEN}", false, false),
		},
		nil,
	)

	atProject := varNamed(projectResp.EnvVars, "AT_PROJECT")
	assert.Equal(t, "h1dden", atProject.Value, "the project itself reads its own secret")
	assert.Empty(t, atProject.Errors)

	assert.Equal(t, "sh4red", varNamed(envResp.EnvVars, "USES_SHARED").Value)
	assert.Equal(t, []envvarservice.ParseErrorType{envvarservice.ParseErrorSecretWithheld},
		errorTypes(varNamed(envResp.EnvVars, "USES_HIDDEN")),
		"the env is told the secret is held back, not that it is missing")

	names := []string{}
	for _, secret := range envResp.Secrets {
		names = append(names, secret.Name)
	}
	assert.Equal(t, []string{"SHARED"}, names)
	assert.Len(t, envResp.InheritedSecrets, 1)
}

// A variable of the scope above carries the value of the secrets it used. When
// one of them is held back, the variable is refused below: its value is that
// secret.
func TestAnInheritedVariableBuiltFromAHeldBackSecretIsRefused(t *testing.T) {
	projectResp, envResp := buildEnvUnderProject(t,
		[]*envvarservice.EnvVar{
			newEnvVar("DB_URL", "postgres://u:${secrets.HIDDEN}@db", false, false),
			newEnvVar("API_URL", "https://${secrets.SHARED}.example.com", false, false),
		},
		[]*entity.Setting{
			scopedSecret("s1", "HIDDEN", "h1dden", false),
			scopedSecret("s2", "SHARED", "sh4red", true),
		},
		nil, nil,
	)

	dbURL := varNamed(envResp.EnvVars, "DB_URL")
	assert.Equal(t, []envvarservice.ParseErrorType{envvarservice.ParseErrorVarUsesWithheldSecret}, errorTypes(dbURL))
	assert.NotContains(t, dbURL.Value, "h1dden")
	assert.Equal(t, "https://sh4red.example.com", varNamed(envResp.EnvVars, "API_URL").Value)

	original := varNamed(projectResp.EnvVars, "DB_URL")
	assert.Equal(t, "postgres://u:h1dden@db", original.Value, "the project's own result is left as it was")
	assert.Empty(t, original.Errors)
}

// A scope's own secret wins over an inherited one of the same name, and an own
// secret is used whatever its flag.
func TestAnOwnSecretIsUsedWhateverItsFlag(t *testing.T) {
	_, envResp := buildEnvUnderProject(t, nil,
		[]*entity.Setting{scopedSecret("s1", "TOKEN", "from-project", true)},
		[]*envvarservice.EnvVar{newEnvVar("T", "${secrets.TOKEN}", false, false)},
		[]*entity.Setting{scopedSecret("s2", "TOKEN", "from-env", false)},
	)
	assert.Equal(t, "from-env", varNamed(envResp.EnvVars, "T").Value)
}

// A variable refused at one boundary is reported once, however far down it goes.
func TestARefusedVariableIsReportedOnce(t *testing.T) {
	refused := inherit([]*envvarservice.EnvVar{func() *envvarservice.EnvVar {
		v := newEnvVar("DB_URL", "x", false, false)
		v.AddRefSecretSetting(scopedSecret("s1", "HIDDEN", "h", false))
		return v
	}()}, nil)
	again := inherit(refused.vars, nil)
	assert.Len(t, again.vars[0].Errors, 1)
}
