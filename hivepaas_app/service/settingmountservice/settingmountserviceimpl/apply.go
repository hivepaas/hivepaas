package settingmountserviceimpl

import (
	"context"
	"errors"
	"reflect"
	"slices"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingmountservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

const serviceUpdateRetryMax = 2

// mounted is what Docker holds of mounted settings, by id and by name.
type mounted struct {
	secrets, configs     map[string]string // id -> name
	secretIDs, configIDs map[string]string // name -> id
}

func (s *service) ApplyToService(
	ctx context.Context, db database.IDB, app *entity.App, spec *swarm.ServiceSpec,
) error {
	contSpec := spec.TaskTemplate.ContainerSpec
	if contSpec == nil {
		return nil
	}
	files, err := s.Resolve(ctx, db, app)
	if err != nil {
		return hperrors.Wrap(err)
	}
	// Every mounted object, whichever app it is labeled for: a preview or a
	// clone starts from another app's spec, and what that app mounts is not
	// this one's.
	have, err := s.listMounted(ctx, settingmountservice.LabelEntry)
	if err != nil {
		return hperrors.Wrap(err)
	}

	// What stays: every reference that is not a mounted setting's. Its targets
	// are taken, and a file claiming one is left out.
	secrets := slices.DeleteFunc(slices.Clone(contSpec.Secrets), func(ref *swarm.SecretReference) bool {
		_, ok := have.secrets[ref.SecretID]
		return ok
	})
	configs := slices.DeleteFunc(slices.Clone(contSpec.Configs), func(ref *swarm.ConfigReference) bool {
		_, ok := have.configs[ref.ConfigID]
		return ok
	})
	taken := map[string]bool{}
	for _, ref := range secrets {
		if ref.File != nil {
			taken[settingmountservice.SecretTarget(ref.File.Name)] = true
		}
	}
	for _, ref := range configs {
		if ref.File != nil {
			taken[settingmountservice.ConfigTarget(ref.File.Name)] = true
		}
	}

	for _, file := range files {
		if taken[file.Path] {
			continue
		}
		name := settingmountservice.ObjectName(app.GlobalKey, file.Entry, file.Part, file.Rotation)
		if file.Sensitive {
			id, err := s.ensureSecret(ctx, have, app, file, name)
			if err != nil {
				return hperrors.Wrap(err)
			}
			secrets = append(secrets, &swarm.SecretReference{SecretID: id, SecretName: name,
				File: &swarm.SecretReferenceFileTarget{Name: file.Path, UID: file.UID, GID: file.GID,
					Mode: file.Mode.ToFileMode()}})
			continue
		}
		id, err := s.ensureConfig(ctx, have, app, file, name)
		if err != nil {
			return hperrors.Wrap(err)
		}
		configs = append(configs, &swarm.ConfigReference{ConfigID: id, ConfigName: name,
			File: &swarm.ConfigReferenceFileTarget{Name: file.Path, UID: file.UID, GID: file.GID,
				Mode: file.Mode.ToFileMode()}})
	}
	contSpec.Secrets, contSpec.Configs = secrets, configs
	return nil
}

// ensureSecret is the id of the secret named name, created when Docker has
// none. A create that finds the name taken - a refresh running beside a
// deployment - reads it back rather than failing.
func (s *service) ensureSecret(
	ctx context.Context, have *mounted, app *entity.App, file *settingmountservice.File, name string,
) (string, error) {
	if id, ok := have.secretIDs[name]; ok {
		return id, nil
	}
	resp, err := s.dockerManager.SecretCreate(ctx, name, file.Data, func(opts *client.SecretCreateOptions) {
		opts.Spec.Labels = settingmountservice.Labels(app.ID, file.Entry, file.Part)
	})
	if err == nil {
		have.secretIDs[name] = resp.ID
		return resp.ID, nil
	}
	if !isConflict(err) {
		return "", hperrors.Wrap(err)
	}
	again, listErr := s.listMounted(ctx, settingmountservice.LabelEntry)
	if listErr != nil {
		return "", hperrors.Wrap(listErr)
	}
	if id, ok := again.secretIDs[name]; ok {
		return id, nil
	}
	return "", hperrors.Wrap(err)
}

// ensureConfig is ensureSecret for a config.
func (s *service) ensureConfig(
	ctx context.Context, have *mounted, app *entity.App, file *settingmountservice.File, name string,
) (string, error) {
	if id, ok := have.configIDs[name]; ok {
		return id, nil
	}
	resp, err := s.dockerManager.ConfigCreate(ctx, name, file.Data, func(opts *client.ConfigCreateOptions) {
		opts.Spec.Labels = settingmountservice.Labels(app.ID, file.Entry, file.Part)
	})
	if err == nil {
		have.configIDs[name] = resp.ID
		return resp.ID, nil
	}
	if !isConflict(err) {
		return "", hperrors.Wrap(err)
	}
	again, listErr := s.listMounted(ctx, settingmountservice.LabelEntry)
	if listErr != nil {
		return "", hperrors.Wrap(listErr)
	}
	if id, ok := again.configIDs[name]; ok {
		return id, nil
	}
	return "", hperrors.Wrap(err)
}

func isConflict(err error) bool {
	return errors.Is(err, hperrors.ErrInfraConflict) || errors.Is(err, hperrors.ErrInfraAlreadyExists)
}

// listMounted lists the secrets and configs carrying every label given, each
// "key" (present) or "key=value" (equal).
func (s *service) listMounted(ctx context.Context, labels ...string) (*mounted, error) {
	have := &mounted{secrets: map[string]string{}, configs: map[string]string{},
		secretIDs: map[string]string{}, configIDs: map[string]string{}}
	secrets, err := s.dockerManager.SecretList(ctx, func(opts *client.SecretListOptions) {
		for _, label := range labels {
			docker.FilterAdd(&opts.Filters, "label", label)
		}
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	for _, secret := range secrets.Items {
		have.secrets[secret.ID], have.secretIDs[secret.Spec.Name] = secret.Spec.Name, secret.ID
	}
	configs, err := s.dockerManager.ConfigList(ctx, func(opts *client.ConfigListOptions) {
		for _, label := range labels {
			docker.FilterAdd(&opts.Filters, "label", label)
		}
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	for _, config := range configs.Items {
		have.configs[config.ID], have.configIDs[config.Spec.Name] = config.Spec.Name, config.ID
	}
	return have, nil
}

// Refresh brings the app's service to its mounted files in one update, which
// restarts it only when a reference changed, and then sweeps.
func (s *service) Refresh(ctx context.Context, db database.IDB, app *entity.App) error {
	if app.ServiceID == "" {
		return nil
	}
	err := s.dockerManager.ServiceUpdateFunc(ctx, app.ServiceID, nil,
		func(_ int, svc *swarm.Service) (bool, error) {
			contSpec := svc.Spec.TaskTemplate.ContainerSpec
			if contSpec == nil {
				return false, nil
			}
			secrets, configs := slices.Clone(contSpec.Secrets), slices.Clone(contSpec.Configs)
			if err := s.ApplyToService(ctx, db, app, &svc.Spec); err != nil {
				return false, hperrors.Wrap(err)
			}
			changed := !reflect.DeepEqual(secrets, contSpec.Secrets) || !reflect.DeepEqual(configs, contSpec.Configs)
			return changed, nil
		}, serviceUpdateRetryMax, 0)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if err = s.Sweep(ctx, app); err != nil {
		s.logger.Warnf("setting mounts of app %s: %v", app.ID, err)
	}
	return nil
}
