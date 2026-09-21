package apptemplateuc

import (
	"context"
	"fmt"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/apptemplateuc/apptemplatedto"
)

// sharedMount is one app of this request reaching the storage of an app that is
// not its own.
type sharedMount struct {
	// Borrower is the app being created that carries the mount.
	Borrower string
	// OwnerKey is the app whose directory it reaches.
	OwnerKey string
	Target   string
	Write    bool
}

func (m sharedMount) describe() string {
	verb := "read"
	if m.Write {
		verb = "read and change"
	}
	return fmt.Sprintf("%s would %s the files of %s at %s", m.Borrower, verb, m.OwnerKey, m.Target)
}

// checkSharedMounts refuses a request whose apps would be given the storage of
// an app the caller may not write to.
//
// An app named here is an app that already exists: the person creating this one
// is asking for its data, in full, for as long as the mount lasts. The gate is
// therefore the one the storage settings screen applies to the same mount -
// Write on that app - rather than anything about the project as a whole.
//
// An app of this very request is exempt. A template whose dependency mounts the
// app it was created for names nothing the caller does not already have: both
// apps are being created by the same person, in the same click, and there is no
// third party whose data is being reached.
func (uc *UC) checkSharedMounts(
	ctx context.Context,
	auth *basedto.Auth,
	req *apptemplatedto.CreateAppFromTemplateReq,
	apps []*appToProvision,
) error {
	created := make(map[string]bool, len(apps))
	for _, target := range apps {
		created[target.key()] = true
	}

	var wanted []sharedMount
	for _, target := range apps {
		for _, mnt := range sharedMountsOf(target) {
			if !created[mnt.OwnerKey] {
				wanted = append(wanted, mnt)
			}
		}
	}
	if len(wanted) == 0 {
		return nil
	}

	for _, mnt := range wanted {
		if err := uc.checkSharedMountOwner(ctx, auth, req, mnt); err != nil {
			return err
		}
	}
	return nil
}

// checkSharedMountOwner answers for one named app: that it is here, and that the
// caller may have what the mount would give.
func (uc *UC) checkSharedMountOwner(
	ctx context.Context,
	auth *basedto.Auth,
	req *apptemplatedto.CreateAppFromTemplateReq,
	mnt sharedMount,
) error {
	apps, _, err := uc.appRepo.List(ctx, uc.db, req.ProjectID, nil,
		bunex.SelectWhere("app.project_env_id = ?", req.ProjectEnvID),
		bunex.SelectWhere("app.key = ?", mnt.OwnerKey),
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
	)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if len(apps) == 0 {
		return hperrors.Wrap(hperrors.ErrAppTemplateInvalid).WithExtraDetail(
			"%s: this environment has no app called %q", mnt.Borrower, mnt.OwnerKey)
	}

	owner := apps[0]
	hasPerm, err := uc.permissionManager.CheckAccess(ctx, uc.db, auth, &permission.AppAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{Action: base.ActionTypeWrite},
		AppID:           owner.ID,
		ParentID:        owner.ParentID,
		ProjectID:       owner.ProjectID,
		ProjectEnv:      owner.ProjectEnvID,
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	if !hasPerm {
		return hperrors.Wrap(hperrors.ErrUnauthorized).
			WithExtraDetail("%s: this needs Write permission on %s", mnt.describe(), owner.Name).
			WithMsgLog("creating an app from a template that mounts another app's storage " +
				"requires Write on that app")
	}
	return nil
}

// sharedMountsOf names, for a person to read, what one app's document reaches of
// another app's storage.
func sharedMountsOf(target *appToProvision) []sharedMount {
	doc := target.rendered.Result.Doc
	if doc == nil || doc.Deployment == nil || doc.Deployment.Storage == nil {
		return nil
	}

	var out []sharedMount
	for mountTarget, mnt := range doc.Deployment.Storage.Mounts {
		src := mnt.SourceApp
		if src == nil || src.App == "" || src.App == target.key() {
			continue
		}
		out = append(out, sharedMount{
			Borrower: target.name,
			OwnerKey: src.App,
			Target:   mountTarget,
			Write:    src.Write,
		})
	}
	return out
}
