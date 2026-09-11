package placementserviceimpl

import (
	"context"
	"testing"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/placementservice"
)

// scopeCapturingSettingRepo stands in for the real repository and records the
// scope it was queried with. Only List is exercised by loadVolumePins; every
// other method is left to the embedded nil interface and would panic if called,
// which is the point - a test relying on them is testing the wrong thing.
type scopeCapturingSettingRepo struct {
	repository.SettingRepo
	gotScope *entity.ObjectScope
	settings []*entity.Setting
}

func (r *scopeCapturingSettingRepo) List(_ context.Context, _ database.IDB, scope *entity.ObjectScope,
	_ *basedto.Paging, _ ...bunex.SelectQueryOption) ([]*entity.Setting, *basedto.PagingMeta, error) {
	r.gotScope = scope
	return r.settings, nil, nil
}

// storedClusterVolumeSetting builds a *entity.Setting the way a row read back
// from the database looks. SetData caches the parsed struct on the setting it
// is called on (setting.go:128), and AsClusterVolume returns that cache
// before ever touching Data (setting.go:156) - so a setting built by calling
// MustSetData on itself and handed straight to a test never exercises the
// JSON round-trip a real row goes through. This writes Data through a
// throwaway setting and copies it onto a fresh one built only from meta's
// storage fields, which never had SetData called on it and so carries no
// cache.
func storedClusterVolumeSetting(t *testing.T, meta *entity.Setting, data *entity.ClusterVolume) *entity.Setting {
	t.Helper()

	tmp := &entity.Setting{Type: meta.Type}
	require.NoError(t, tmp.SetData(data))

	fresh := *meta
	fresh.Data = tmp.Data
	return &fresh
}

func mountingService(refID string) *swarm.Service {
	return &swarm.Service{
		Spec: swarm.ServiceSpec{
			TaskTemplate: swarm.TaskSpec{
				ContainerSpec: &swarm.ContainerSpec{
					Mounts: []mount.Mount{{Type: mount.TypeVolume, Source: refID}},
				},
			},
		},
	}
}

// A volume most commonly mounted by an app is not one the app itself defined -
// it is the project's default volume, created once and shared by every app in
// the project. loadVolumePins has to find that pin through the app's own scope,
// not just settings the app owns directly, or the pin silently stops being
// enforced for the common case.
//
// This is the same query storage_settings_update.go relies on (via ListByIDs,
// which calls List with the same app scope) to decide whether an app may mount
// a volume at all: applyAppFilter walks up to the parent app, the project env,
// the project, and global for any setting marked Inheritable - which is exactly
// how CreateProjectDefaultVolume creates the project's volume
// (Inheritable: true, see volumeserviceimpl/project.go). So a volume the app is
// allowed to mount is a volume this query finds.
func TestLoadVolumePinsUsesAppScopeSoParentScopeVolumesAreFound(t *testing.T) {
	app := &entity.App{ID: "app-1", ProjectID: "proj-1"}

	projectVolume := storedClusterVolumeSetting(t, &entity.Setting{
		ID:          "setting-1",
		Scope:       base.ObjectScopeProject,
		ObjectID:    "proj-1",
		Type:        base.SettingTypeClusterVolume,
		Name:        "default",
		RefID:       "vol-ref-1",
		Inheritable: true,
	}, &entity.ClusterVolume{NodeID: "node-1"})

	repo := &scopeCapturingSettingRepo{settings: []*entity.Setting{projectVolume}}
	svc := &service{settingRepo: repo}

	data := &placementSettingsData{
		ApplyPlacementSettingsReq: &placementservice.ApplyPlacementSettingsReq{
			App:     app,
			Service: mountingService("vol-ref-1"),
		},
	}

	pins, err := svc.loadVolumePins(context.Background(), nil, data)

	require.NoError(t, err)
	assert.Equal(t, app.GetObjectScope(), repo.gotScope,
		"the query must use the app's own scope - that is what makes the scope"+
			" filter walk up to settings the project marked inheritable")
	assert.False(t, repo.gotScope.NoInherited,
		"NoInherited would drop the project's volume from the result")
	assert.Equal(t, []placementservice.VolumePin{{VolumeName: "default", NodeID: "node-1"}}, pins)
}

func TestLoadVolumePinsSkipsTheQueryWhenTheServiceHasNoMounts(t *testing.T) {
	app := &entity.App{ID: "app-1"}
	repo := &scopeCapturingSettingRepo{}
	svc := &service{settingRepo: repo}

	data := &placementSettingsData{
		ApplyPlacementSettingsReq: &placementservice.ApplyPlacementSettingsReq{
			App: app,
			Service: &swarm.Service{
				Spec: swarm.ServiceSpec{
					TaskTemplate: swarm.TaskSpec{
						ContainerSpec: &swarm.ContainerSpec{},
					},
				},
			},
		},
	}

	pins, err := svc.loadVolumePins(context.Background(), nil, data)

	require.NoError(t, err)
	assert.Nil(t, pins)
	assert.Nil(t, repo.gotScope, "a service with no mounts at all must not query settings")
}

// A bind mount that does not come from any volume HivePaaS knows about (the
// user may have added it directly) still has to be checked against whatever
// cluster volumes the app can see - loadVolumePins can no longer tell in
// advance, from the mount alone, that a bind is unrelated to any volume. It
// must still come back with no pins once VolumePinsForMounts finds nothing to
// match.
func TestLoadVolumePinsIgnoresABindMountThatMatchesNoVolume(t *testing.T) {
	app := &entity.App{ID: "app-1"}
	repo := &scopeCapturingSettingRepo{}
	svc := &service{settingRepo: repo}

	data := &placementSettingsData{
		ApplyPlacementSettingsReq: &placementservice.ApplyPlacementSettingsReq{
			App: app,
			Service: &swarm.Service{
				Spec: swarm.ServiceSpec{
					TaskTemplate: swarm.TaskSpec{
						ContainerSpec: &swarm.ContainerSpec{
							Mounts: []mount.Mount{{Type: mount.TypeBind, Source: "/host/path"}},
						},
					},
				},
			},
		},
	}

	pins, err := svc.loadVolumePins(context.Background(), nil, data)

	require.NoError(t, err)
	assert.Empty(t, pins)
	assert.NotNil(t, repo.gotScope, "a bind mount has no ref id to pre-filter on, so it must still query")
}

// This is the case Task 6b exists for: a volume pinned to a node is mounted as
// a plain bind (useBindMountIfAppropriate rewrites a local/type=none volume
// into one), so loadVolumePins has to find its pin by matching the mount's host
// path against the volume's recorded device, not by ref id.
func TestLoadVolumePinsFindsAPinBehindABindMount(t *testing.T) {
	app := &entity.App{ID: "app-1"}

	bindVolume := storedClusterVolumeSetting(t, &entity.Setting{
		ID:   "setting-1",
		Type: base.SettingTypeClusterVolume,
		Name: "webroot",
	}, &entity.ClusterVolume{
		NodeID:     "node-1",
		Managed:    true,
		DriverOpts: map[string]string{"type": "none", "device": "/srv/data"},
	})

	repo := &scopeCapturingSettingRepo{settings: []*entity.Setting{bindVolume}}
	svc := &service{settingRepo: repo}

	data := &placementSettingsData{
		ApplyPlacementSettingsReq: &placementservice.ApplyPlacementSettingsReq{
			App: app,
			Service: &swarm.Service{
				Spec: swarm.ServiceSpec{
					TaskTemplate: swarm.TaskSpec{
						ContainerSpec: &swarm.ContainerSpec{
							Mounts: []mount.Mount{{Type: mount.TypeBind, Source: "/srv/data/shop/prod/web"}},
						},
					},
				},
			},
		},
	}

	pins, err := svc.loadVolumePins(context.Background(), nil, data)

	require.NoError(t, err)
	assert.Equal(t, []placementservice.VolumePin{{VolumeName: "webroot", NodeID: "node-1"}}, pins)
}
