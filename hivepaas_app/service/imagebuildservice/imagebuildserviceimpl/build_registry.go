package imagebuildserviceimpl

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/moby/moby/api/types/registry"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

func (s *service) calcBuildRegistryAuths(
	ctx context.Context,
	db database.IDB,
	data *imageBuildData,
) (map[string]registry.AuthConfig, error) {
	settings, _, err := s.settingRepo.List(ctx, db, data.App.Project.GetObjectScope(), nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeRegistryAuth),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
	)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	result := make(map[string]registry.AuthConfig, len(settings))
	secrets := make([]string, 0, len(settings))
	for _, setting := range settings {
		regAuth, err := setting.AsRegistryAuth()
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		password, err := regAuth.Password.GetPlain()
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		if password != "" {
			secrets = append(secrets, password)
		}
		result[regAuth.Address] = registry.AuthConfig{
			Username:      regAuth.Username,
			Password:      password,
			ServerAddress: regAuth.Address,
		}
	}

	if data.LogStore != nil && len(secrets) > 0 {
		data.LogStore.UpdateRedactorAddSecrets(secrets)
	}

	return result, nil
}

func (s *service) prepareDockerConfigDir(
	data *imageBuildData,
) (string, func(), error) {
	if len(data.RegistryAuths) == 0 {
		return "", func() {}, nil
	}

	configDir, err := os.MkdirTemp(data.TempDir, "docker-config-*")
	if err != nil {
		return "", func() {}, apperrors.Wrap(err)
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
		return "", func() {}, apperrors.Wrap(err)
	}

	err = os.WriteFile(filepath.Join(configDir, "config.json"), content, 0600) //nolint:mnd
	if err != nil {
		cleanup()
		return "", func() {}, apperrors.Wrap(err)
	}

	return configDir, cleanup, nil
}
