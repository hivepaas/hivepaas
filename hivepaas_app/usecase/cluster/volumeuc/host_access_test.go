package volumeuc

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/volumeuc/volumedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/services/docker"
)

// answeringPermissionManager answers as told and keeps what it was asked.
type answeringPermissionManager struct {
	permission.Manager
	allow   bool
	checked []permission.AccessCheck
}

func (f *answeringPermissionManager) CheckAccess(
	_ context.Context, _ database.IDB, _ *basedto.Auth, check permission.AccessCheck,
) (bool, error) {
	f.checked = append(f.checked, check)
	return f.allow, nil
}

func hostAccessUC(allow bool) (*UC, *answeringPermissionManager) {
	perms := &answeringPermissionManager{allow: allow}
	return &UC{BaseUC: &settings.BaseUC{PermissionManager: perms}}, perms
}

func volumeReq() *volumedto.VolumeBaseReq {
	return &volumedto.VolumeBaseReq{Name: "data", Driver: docker.VolumeDriverLocal}
}

// A volume of docker's own storage, and a bind whose directory HivePaaS
// chooses, are the ordinary case: nothing is asked.
func TestVolumesThatReachNoPathOfTheirOwnAskNothing(t *testing.T) {
	for name, req := range map[string]*volumedto.VolumeBaseReq{
		"docker's own storage": volumeReq(),
		"a managed bind": func() *volumedto.VolumeBaseReq {
			r := volumeReq()
			r.BindOptions = &volumedto.VolumeBindOptionsReq{}
			return r
		}(),
	} {
		uc, perms := hostAccessUC(false)

		err := uc.checkVolumeHostAccess(context.Background(), &basedto.Auth{}, req)

		assert.NoError(t, err, name)
		assert.Empty(t, perms.checked, name)
	}
}

// A directory of the node is the decision Write on the Cluster module stands
// for, whether it is described by bind options or by raw driver options.
func TestAVolumeNamingAPathOfTheNodeNeedsWriteOnTheCluster(t *testing.T) {
	bind := volumeReq()
	bind.BindOptions = &volumedto.VolumeBindOptionsReq{Directory: "/srv/backups"}
	raw := volumeReq()
	raw.Options = map[string]string{"type": "none", "o": "bind,rw", "device": "/srv/backups"}

	for name, req := range map[string]*volumedto.VolumeBaseReq{"bind options": bind, "driver options": raw} {
		for allowed, wantRefusal := range map[bool]bool{false: true, true: false} {
			uc, perms := hostAccessUC(allowed)

			err := uc.checkVolumeHostAccess(context.Background(), &basedto.Auth{}, req)

			assert.Equal(t, wantRefusal, errors.Is(err, hperrors.ErrUnauthorized), "%s: %v", name, err)
			if assert.Len(t, perms.checked, 1, name) {
				check, ok := perms.checked[0].(*permission.ModuleAccessCheck)
				if assert.True(t, ok, name) {
					assert.Equal(t, base.ResourceModuleCluster, check.Module, name)
					assert.Equal(t, base.ActionTypeWrite, check.Action, name)
				}
			}
		}
	}
}

// The Docker API is granted through an app's Docker API settings, which say
// what it may do with it. A volume holding the socket would go around them, so
// it is refused to everyone - permission is not what is missing.
func TestAVolumeReachingTheDockerSocketIsRefusedToEveryone(t *testing.T) {
	for _, directory := range []string{"/var/run/docker.sock", "/var/run", "/"} {
		req := volumeReq()
		req.BindOptions = &volumedto.VolumeBindOptionsReq{Directory: directory}
		uc, perms := hostAccessUC(true)

		err := uc.checkVolumeHostAccess(context.Background(), &basedto.Auth{}, req)

		assert.True(t, errors.Is(err, hperrors.ErrArgumentInvalid), "%s: %v", directory, err)
		assert.Empty(t, perms.checked, "%s: no permission makes this allowed", directory)
	}

	raw := volumeReq()
	raw.Options = map[string]string{"device": "/var/run/docker.sock"}
	uc, _ := hostAccessUC(true)
	assert.True(t, errors.Is(uc.checkVolumeHostAccess(context.Background(), &basedto.Auth{}, raw),
		hperrors.ErrArgumentInvalid))
}
