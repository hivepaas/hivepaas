package envvarserviceimpl

import (
	"slices"
	"sort"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
)

// inherited is what a scope takes from the scope above it - a project's env, an
// env's app, an app's preview.
type inherited struct {
	vars    []*envvarservice.EnvVar
	secrets []*entity.Setting
	// withheld are the names of the secrets held back.
	withheld map[string]struct{}
}

// inherit keeps the secrets marked inheritable, as the settings repository does
// for every other setting, and the variables. A variable whose value was built
// from a secret that is not inheritable holds that secret, so it is refused: it
// comes down emptied, with an error that names the secret.
//
// The variables of the scope above are left as they are: they are that scope's
// own result, and may be shared with the other scopes below it.
func inherit(vars []*envvarservice.EnvVar, secrets []*entity.Setting) *inherited {
	res := &inherited{
		vars:     make([]*envvarservice.EnvVar, 0, len(vars)),
		secrets:  make([]*entity.Setting, 0, len(secrets)),
		withheld: map[string]struct{}{},
	}
	for _, secret := range secrets {
		if secret.Inheritable {
			res.secrets = append(res.secrets, secret)
			continue
		}
		res.withheld[secret.Name] = struct{}{}
	}

	for _, v := range vars {
		var usedWithheld []string
		for _, setting := range v.RefSecretSettings {
			if !setting.Inheritable {
				usedWithheld = append(usedWithheld, setting.Name)
			}
		}
		if len(usedWithheld) == 0 {
			res.vars = append(res.vars, v)
			continue
		}
		sort.Strings(usedWithheld)
		envVar := *v.EnvVar
		envVar.Value = ""
		res.vars = append(res.vars, &envvarservice.EnvVar{
			EnvVar: &envVar,
			Errors: append(slices.Clone(v.Errors), &envvarservice.ParseError{
				Type:    envvarservice.ParseErrorVarUsesWithheldSecret,
				Name:    v.Key,
				Secrets: usedWithheld,
			}),
		})
	}
	return res
}
