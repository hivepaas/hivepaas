package imagebuildserviceimpl

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/moby/moby/api/types/registry"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
)

// calcBuildRegistryAuths is the project's registries by address, with their
// passwords opened, and those passwords.
func (s *service) calcBuildRegistryAuths(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
) (map[string]registry.AuthConfig, []string, error) {
	settings, _, err := s.settingRepo.List(ctx, db, app.Project.GetObjectScope(), nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeRegistryAuth),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
	)
	if err != nil {
		return nil, nil, hperrors.Wrap(err)
	}

	result := make(map[string]registry.AuthConfig, len(settings))
	secrets := make([]string, 0, len(settings))
	for _, setting := range settings {
		auth, err := s.registryAuthService.AuthConfig(ctx, setting)
		if err != nil {
			// An ECR credential whose keys AWS refuses is left out rather than
			// failing a build that may not pull from it; one that does fails
			// on the pull, which says so.
			if regAuth, _ := setting.AsRegistryAuth(); regAuth != nil && regAuth.Kind == base.RegistryAuthKindAWSECR {
				logging.Warnf("image build: registry credential %s left out: %v", setting.ID, err)
				continue
			}
			return nil, nil, hperrors.Wrap(err)
		}
		if auth.Password != "" {
			secrets = append(secrets, auth.Password)
		}
		result[auth.ServerAddress] = *auth
	}

	return result, secrets, nil
}

func (s *service) prepareDockerConfigDir(
	data *imageBuildData,
) (string, func(), error) {
	if len(data.RegistryAuths) == 0 {
		return "", func() {}, nil
	}

	configDir, err := os.MkdirTemp(data.TempDir, "docker-config-*")
	if err != nil {
		return "", func() {}, hperrors.Wrap(err)
	}
	cleanup := func() {
		_ = os.RemoveAll(configDir)
	}

	type dockerAuthItem struct {
		Username string `json:"username,omitempty"`
		Password string `json:"password,omitempty"`
		Auth     string `json:"auth,omitempty"`
	}
	type dockerConfig struct {
		Auths map[string]dockerAuthItem `json:"auths"`
	}

	cfg := dockerConfig{
		Auths: make(map[string]dockerAuthItem, len(data.RegistryAuths)),
	}
	for addr, auth := range data.RegistryAuths {
		encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth.Username + ":" + auth.Password))
		cfg.Auths[addr] = dockerAuthItem{
			Username: auth.Username,
			Password: auth.Password,
			Auth:     encodedAuth,
		}
	}

	content, err := json.Marshal(cfg)
	if err != nil {
		cleanup()
		return "", func() {}, hperrors.Wrap(err)
	}

	err = os.WriteFile(filepath.Join(configDir, "config.json"), content, 0600) //nolint:mnd
	if err != nil {
		cleanup()
		return "", func() {}, hperrors.Wrap(err)
	}

	// Symlink host docker directories into temporary configDir so plugins,
	// buildx instances, contexts, and sockets remain available.
	if homeDir, err := os.UserHomeDir(); err == nil {
		dirsToLink := []string{"cli-plugins", "buildx", "contexts", "run"}
		for _, dirName := range dirsToLink {
			hostDir := filepath.Join(homeDir, ".docker", dirName)
			if info, err := os.Stat(hostDir); err == nil && info.IsDir() {
				_ = os.Symlink(hostDir, filepath.Join(configDir, dirName))
			}
		}
	}

	return configDir, cleanup, nil
}
