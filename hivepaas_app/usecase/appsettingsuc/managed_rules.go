package appsettingsuc

import (
	"context"
	"slices"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/services/docker/dockerhelper"
)

// A managed app's engine owns part of it: what its locks name, recorded when the
// app was made. These are the rules the settings' updates keep. Each refuses
// with ERR_APP_MANAGED_FIELD_LOCKED, naming the field; none silently puts a
// value back. Each passes anything on an app no engine looks after (nil locks).

// managedLocksOf is the locks of a managed app, from its app-managed setting
// among the settings loaded with it; nil for any other app.
func managedLocksOf(app *entity.App) (*entity.ManagedLocks, error) {
	setting := app.GetSettingByType(base.SettingTypeAppManaged)
	if setting == nil || setting.Status != base.SettingStatusActive {
		return nil, nil
	}
	managed, err := setting.AsAppManagedSettings()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &managed.Locks, nil
}

// loadManagedLocks is the locks of an app by its id, read on their own so that
// a use case need not load the setting with the app.
func (uc *UC) loadManagedLocks(ctx context.Context, db database.IDB, appID string) (*entity.ManagedLocks, error) {
	settings, _, err := uc.settingRepo.List(ctx, db, nil, nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeAppManaged),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
		bunex.SelectWhere("setting.object_id = ?", appID),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return managedLocksOf(&entity.App{Settings: settings})
}

func managedFieldLocked(field string) error {
	return hperrors.Wrap(hperrors.ErrAppManagedFieldLocked).WithParam("Field", field).ExposeParams("Field")
}

// checkManagedKind keeps a managed app what its engine looks after.
func checkManagedKind(locks *entity.ManagedLocks, current *entity.AppKindSettings,
	nextCategory base.AppCategory, nextEngine string) error {
	if locks == nil || current == nil {
		return nil
	}
	if nextCategory != current.Category || nextEngine != current.Engine {
		return managedFieldLocked("kind")
	}
	return nil
}

// checkManagedImage keeps the repository the app was made with, and lets the
// tag move within its version line and its family: another major would not open
// the data directory, and another base image may not run the artifacts.
func checkManagedImage(locks *entity.ManagedLocks, image string) error {
	if locks == nil || locks.Image == "" {
		return nil
	}
	class, err := templatemodel.ClassifyImageOverride(locks.ImageLine, locks.Image, image)
	if err != nil || class != templatemodel.ImageOverrideSameLine ||
		templatemodel.ImageTagFamily(image) != templatemodel.ImageTagFamily(locks.Image) {
		return managedFieldLocked("image")
	}
	return nil
}

// checkManagedCommand keeps the engine's arguments first: what a person adds
// comes after them. The two are compared as the arguments Docker runs, so
// spacing and quoting do not matter.
func checkManagedCommand(locks *entity.ManagedLocks, command string) error {
	if locks == nil || locks.Command == "" {
		return nil
	}
	owned, err := dockerhelper.CommandArgs(locks.Command)
	if err != nil {
		return hperrors.Wrap(err)
	}
	args, err := dockerhelper.CommandArgs(command)
	if err != nil || len(args) < len(owned) || !slices.Equal(args[:len(owned)], owned) {
		return managedFieldLocked("command")
	}
	return nil
}

// checkManagedDeployment keeps a managed app deployed from its image, by its
// entrypoint, with the engine's arguments.
func checkManagedDeployment(locks *entity.ManagedLocks, current, next *entity.AppDeploymentSettings) error {
	if locks == nil {
		return nil
	}
	if next.ActiveMethod != base.DeploymentMethodImage || next.ImageSource == nil {
		return managedFieldLocked("deploymentMethod")
	}
	if current != nil && next.Entrypoint != current.Entrypoint {
		return managedFieldLocked("entrypoint")
	}
	if err := checkManagedImage(locks, next.ImageSource.Image); err != nil {
		return err
	}
	return checkManagedCommand(locks, next.Command)
}

// checkManagedContainer keeps the same on the container screen, which writes the
// service directly, and the user the server runs as.
func checkManagedContainer(locks *entity.ManagedLocks, current *swarm.ContainerSpec,
	image, entrypoint, command, user string) error {
	if locks == nil || current == nil {
		return nil
	}
	entrypointArgs, err := dockerhelper.CommandArgs(entrypoint)
	if err != nil || !slices.Equal(entrypointArgs, current.Command) {
		return managedFieldLocked("entrypoint")
	}
	if user != current.User {
		return managedFieldLocked("user")
	}
	if err = checkManagedImage(locks, image); err != nil {
		return err
	}
	return checkManagedCommand(locks, command)
}

// checkManagedMounts keeps every mount the engine relies on as it is: its source,
// its options, writable. Mounts may be added.
func checkManagedMounts(locks *entity.ManagedLocks, current, final []mount.Mount) error {
	if locks == nil {
		return nil
	}
	byTarget := func(mounts []mount.Mount, target string) *mount.Mount {
		for i := range mounts {
			if mounts[i].Target == target {
				return &mounts[i]
			}
		}
		return nil
	}
	for _, target := range locks.Mounts {
		was, now := byTarget(current, target), byTarget(final, target)
		if was == nil {
			continue
		}
		if now == nil || now.Type != was.Type || now.Source != was.Source || now.ReadOnly != was.ReadOnly ||
			!sameVolumeSubpath(was, now) {
			return managedFieldLocked("storage")
		}
	}
	return nil
}

func sameVolumeSubpath(a, b *mount.Mount) bool {
	subpath := func(m *mount.Mount) string {
		if m.VolumeOptions == nil {
			return ""
		}
		return m.VolumeOptions.Subpath
	}
	return subpath(a) == subpath(b)
}

// checkManagedEnv keeps the variables the engine owns as they are, present or
// absent; the others are the person's.
func checkManagedEnv(locks *entity.ManagedLocks, current, next []*entity.EnvVar) error {
	if locks == nil {
		return nil
	}
	of := func(vars []*entity.EnvVar, key string) []entity.EnvVar {
		var out []entity.EnvVar
		for _, env := range vars {
			if env != nil && env.Key == key {
				out = append(out, *env)
			}
		}
		return out
	}
	for _, key := range locks.EnvKeys {
		if !slices.Equal(of(current, key), of(next, key)) {
			return managedFieldLocked("envVars")
		}
	}
	return nil
}
