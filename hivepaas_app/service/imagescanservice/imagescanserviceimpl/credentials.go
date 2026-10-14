package imagescanserviceimpl

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"

	"github.com/distribution/reference"
	"github.com/moby/moby/api/types/registry"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

// dockerHub is Docker Hub's domain as a reference names it, and hubAuthKey the
// key a config.json keeps its credentials under.
const (
	dockerHub  = "docker.io"
	hubAuthKey = "https://index.docker.io/v1/"
)

// pullAuth is the credential the app's service pulls its image with: its
// image's, or that of the registry its builds are pushed to. Nil when it pulls
// without one, or the credential cannot be read: the scan then tries the
// registry as anybody may.
func (s *service) pullAuth(ctx context.Context, app *entity.App) (*registry.AuthConfig, error) {
	deployments, _, err := s.settingRepo.List(ctx, s.db, nil, nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeAppDeployment),
		bunex.SelectWhere("setting.object_id = ?", app.ID),
		bunex.SelectLimit(1),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if len(deployments) == 0 {
		return nil, nil
	}
	deployment, err := deployments[0].AsAppDeploymentSettings()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	authID := deployment.ServiceRegistryAuthID()
	if authID == "" {
		return nil, nil
	}
	setting, err := s.settingRepo.GetByID(ctx, s.db, nil, base.SettingTypeRegistryAuth, authID, true)
	if err != nil || setting == nil {
		return nil, nil //nolint:nilerr // a scan of a public image needs none
	}
	auth, err := s.registryAuthService.AuthConfig(ctx, setting)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return auth, nil
}

// dockerConfig is a Docker config.json giving auth's credentials to the
// registry of ref; nil when auth is for another registry, or there is none.
func dockerConfig(auth *registry.AuthConfig, ref string) ([]byte, error) {
	if auth == nil || ref == "" || auth.Username == "" {
		return nil, nil
	}
	named, err := reference.ParseNormalizedNamed(ref)
	if err != nil {
		return nil, nil //nolint:nilerr // no reference, no registry to give credentials to
	}
	domain := reference.Domain(named)
	if registryDomain(auth.ServerAddress) != domain {
		return nil, nil
	}
	key := domain
	if domain == dockerHub {
		key = hubAuthKey
	}
	encoded := base64.StdEncoding.EncodeToString([]byte(auth.Username + ":" + auth.Password))
	data, err := json.Marshal(map[string]any{"auths": map[string]any{key: map[string]string{"auth": encoded}}})
	return data, hperrors.Wrap(err)
}

// registryDomain is a registry's address as a reference's domain would name
// it: no scheme, no path, Docker Hub's names as one.
func registryDomain(address string) string {
	address = strings.TrimPrefix(strings.TrimPrefix(address, "https://"), "http://")
	address, _, _ = strings.Cut(address, "/")
	switch address {
	case "", "index.docker.io", "registry-1.docker.io":
		return dockerHub
	}
	return address
}
