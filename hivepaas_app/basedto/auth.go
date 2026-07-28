package basedto

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/jwtsession"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/projecthelper"
)

type User struct {
	*entity.User
	AuthClaims *jwtsession.AuthClaims
}

func (u *User) Entity() *entity.User {
	if u != nil {
		return u.User
	}
	return nil
}

type Auth struct {
	User *User

	// AllowedResources resources which the current user can access to
	AllowedResources map[base.ResourceType][]string
}

func (auth *Auth) AllowedUsers(inIDs []string) (allowAll bool, allowed []string) {
	return auth.calcResIntersection(auth.AllowedResources[base.ResourceTypeUser], inIDs)
}

func (auth *Auth) AllowedProjects(inIDs []string) (allowAll bool, allowed []string) {
	return auth.calcResIntersection(auth.AllowedResources[base.ResourceTypeProject], inIDs)
}

func (auth *Auth) AllowedProjectEnvs(inEnvs []string) (allowAll bool, allowed []string) {
	for i, env := range inEnvs {
		envKey := projecthelper.CalcProjectEnvKey(env)
		if envKey != "" {
			inEnvs[i] = envKey
		}
	}
	allowedIDs := auth.AllowedResources[base.ResourceTypeProjectEnv]
	for i, id := range allowedIDs {
		_, envKey := projecthelper.ParseProjectEnvID(id)
		if envKey != "" {
			allowedIDs[i] = envKey
		}
	}
	return auth.calcResIntersection(allowedIDs, inEnvs)
}

func (auth *Auth) AllowedApps(inIDs []string) (allowAll bool, allowed []string) {
	return auth.calcResIntersection(auth.AllowedResources[base.ResourceTypeApp], inIDs)
}

func (auth *Auth) AllowedDeployments(inIDs []string) (allowAll bool, allowed []string) {
	return auth.calcResIntersection(auth.AllowedResources[base.ResourceTypeDeployment], inIDs)
}

func (auth *Auth) AllowedImages(inIDs []string) (allowAll bool, allowed []string) {
	return auth.calcResIntersection(auth.AllowedResources[base.ResourceTypeImage], inIDs)
}

func (auth *Auth) AllowedFiles(inIDs []string) (allowAll bool, allowed []string) {
	return auth.calcResIntersection(auth.AllowedResources[base.ResourceTypeFile], inIDs)
}

func (auth *Auth) AllowedSettings(inIDs []string) (allowAll bool, allowed []string) {
	return auth.calcResIntersection(auth.AllowedResources[base.ResourceTypeSetting], inIDs)
}

// calcResIntersection calculates the final allowed resource IDs.
func (auth *Auth) calcResIntersection(allowedIDs, checkIDs []string) (allowAll bool, allowed []string) {
	if len(checkIDs) == 0 {
		checkIDs = nil
	}
	if allowedIDs == nil {
		return checkIDs == nil, checkIDs
	}
	if len(allowedIDs) == 0 {
		return false, nil // no access on any object of the resource type
	}
	// checkIDs:   a b c
	// allowedIDs: a b d -> result: a b
	return false, gofn.FilterIN(checkIDs, allowedIDs...)
}
