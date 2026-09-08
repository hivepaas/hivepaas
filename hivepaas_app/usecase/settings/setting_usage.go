package settings

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

// SettingUsage is one object that references a setting.
//
// It carries enough to build the exact page that object lives on, not just its
// name. A list saying "used by 10 things" and leaving the operator to find them
// is barely better than the refusal it explains - and with app-scoped settings
// the object could be on any of hundreds of pages.
//
// Everything past Type and ID is best effort. The link table stores ids, and the
// object it points at may have been deleted, or be of a kind this does not know
// how to place. A row with an id and nothing else is still worth showing: it is
// the difference between "something uses this" and "these three things do".
type SettingUsage struct {
	// Type is what holds the reference. In practice almost always a setting -
	// every implementation of GetResourceLinks passes ResourceTypeSetting - but
	// the column allows more, so this is reported rather than assumed.
	Type base.ResourceType
	ID   string
	Name string

	// SettingType and Scope are set when Type is a setting: which kind it is, and
	// which of the dashboard's setting pages it therefore lives on.
	SettingType base.SettingType
	Scope       base.ObjectScopeType

	// Where the owning object lives. ProjectEnvKey rather than the env id, because
	// that is what the URL carries - see projecthelper.CalcProjectEnvID.
	ProjectID     string
	ProjectEnvKey string
	AppID         string
	AppName       string
	UserID        string
}

// GetSettingUsages is LoadSettingUsages for a caller with no transaction of its
// own - the read-only endpoint behind the dashboard's "what uses this" modal.
func (uc *BaseUC) GetSettingUsages(ctx context.Context, settingID string) ([]*SettingUsage, error) {
	usages, err := uc.LoadSettingUsages(ctx, uc.DB, settingID)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return usages, nil
}

// LoadSettingUsages reports what still references a setting.
//
// Read from res_links rather than by scanning every setting's payload: the links
// are written whenever a setting is saved, so they are the index that already
// exists for this question.
func (uc *BaseUC) LoadSettingUsages(
	ctx context.Context,
	db database.IDB,
	settingID string,
) ([]*SettingUsage, error) {
	links, _, err := uc.ResLinkRepo.List(ctx, db, nil,
		bunex.SelectWhere("res_link.dst_type = ?", base.ResourceTypeSetting),
		bunex.SelectWhere("res_link.dst_id = ?", settingID),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	usages := make([]*SettingUsage, 0, len(links))
	srcSettingIDs := make([]string, 0, len(links))
	for _, link := range links {
		// A setting linking to itself is not a use of it by something else, and
		// showing it would tell the operator to go and edit the thing they are
		// already trying to delete.
		if link.SrcID == settingID {
			continue
		}
		usages = append(usages, &SettingUsage{Type: link.SrcType, ID: link.SrcID})
		if link.SrcType == base.ResourceTypeSetting {
			srcSettingIDs = append(srcSettingIDs, link.SrcID)
		}
	}
	if len(srcSettingIDs) == 0 {
		return usages, nil
	}

	if err = uc.describeSettingUsages(ctx, db, usages, srcSettingIDs); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return usages, nil
}

// describeSettingUsages fills in what each referencing setting is and where it
// lives, in two batched reads rather than one per row.
//
// Missing rows are left described only by their id. A reference to something
// already deleted is exactly the state this whole endpoint exists to report on,
// so it must not be the thing that makes reporting fail.
func (uc *BaseUC) describeSettingUsages(
	ctx context.Context,
	db database.IDB,
	usages []*SettingUsage,
	srcSettingIDs []string,
) error {
	srcSettings, err := uc.SettingRepo.ListByIDs(ctx, db, nil, srcSettingIDs, false)
	if err != nil {
		return hperrors.Wrap(err)
	}
	settingByID := make(map[string]*entity.Setting, len(srcSettings))
	appIDs := make([]string, 0, len(srcSettings))
	for _, setting := range srcSettings {
		settingByID[setting.ID] = setting
		if setting.Scope == base.ObjectScopeApp && setting.ObjectID != "" {
			appIDs = append(appIDs, setting.ObjectID)
		}
	}

	appByID := map[string]*entity.App{}
	if len(appIDs) > 0 {
		apps, e := uc.AppService.LoadAppsSkipMissing(ctx, db, "", appIDs, false, false,
			bunex.SelectRelation("ProjectEnv"),
		)
		if e != nil {
			return hperrors.Wrap(e)
		}
		for _, app := range apps {
			appByID[app.ID] = app
		}
	}

	for _, usage := range usages {
		setting := settingByID[usage.ID]
		if setting == nil {
			continue
		}
		usage.Name = setting.Name
		usage.SettingType = setting.Type
		usage.Scope = setting.Scope

		switch setting.Scope {
		case base.ObjectScopeApp:
			usage.AppID = setting.ObjectID
			if app := appByID[setting.ObjectID]; app != nil {
				usage.AppName = app.Name
				usage.ProjectID = app.ProjectID
				if app.ProjectEnv != nil {
					usage.ProjectEnvKey = app.ProjectEnv.Key
				}
			}
		case base.ObjectScopeProject, base.ObjectScopeProjectEnv:
			usage.ProjectID = setting.ObjectID
		case base.ObjectScopeUser:
			usage.UserID = setting.ObjectID
		case base.ObjectScopeGlobal, base.ObjectScopeHivepaas:
			// Nothing to place: these live on a single page of their own.
		}
	}

	return nil
}

// ensureSettingNotInUse refuses a deletion that would leave dangling references.
//
// The references live inside the referencing objects' own payloads, so removing
// this row does not remove them - it turns them into ids that resolve to nothing.
// And the apply path loads references strictly: the next time any of those
// objects is applied, for any reason, it fails with ErrSettingNotFound. That
// failure would arrive days later, attached to an unrelated deploy, with nothing
// pointing back to the deletion that caused it.
func (uc *BaseUC) ensureSettingNotInUse(
	ctx context.Context,
	db database.IDB,
	setting *entity.Setting,
) error {
	usages, err := uc.LoadSettingUsages(ctx, db, setting.ID)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if len(usages) == 0 {
		return nil
	}

	return hperrors.Wrap(hperrors.ErrSettingInUse).
		WithParam("Name", setting.Name).
		WithParam("Count", len(usages))
}
