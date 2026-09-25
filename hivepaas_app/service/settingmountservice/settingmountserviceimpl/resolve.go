package settingmountserviceimpl

import (
	"context"
	"slices"
	"strings"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/fileutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingmountservice"
)

const (
	defaultUID  = "0"
	defaultGID  = "0"
	defaultMode = fileutil.FileMode(0o444)
)

func (s *service) Resolve(ctx context.Context, db database.IDB, app *entity.App) ([]*settingmountservice.File, error) {
	entries, err := s.loadEntries(ctx, db, app.ID)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	entries = gofn.Filter(entries, func(e *entity.Setting) bool {
		return e.Status == base.SettingStatusActive && settingmountservice.ValidEntryKey(e.Name)
	})
	// Key order, so that the first claim to a path is the same one every time.
	slices.SortFunc(entries, func(a, b *entity.Setting) int { return strings.Compare(a.Name, b.Name) })

	mounts := make(map[string]*entity.AppSettingMount, len(entries))
	sourceIDs := make([]string, 0, len(entries))
	for _, e := range entries {
		mount, err := e.AsAppSettingMount()
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		mounts[e.ID] = mount
		sourceIDs = append(sourceIDs, mount.Source.ID)
	}
	sourceIDs = gofn.ToSet(gofn.ToSliceSkippingZero(sourceIDs...))
	if len(sourceIDs) == 0 {
		return nil, nil
	}
	sources, err := s.loadSources(ctx, db, app, sourceIDs)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	byID := make(map[string]*entity.Setting, len(sources))
	for _, source := range sources {
		byID[source.ID] = source
	}

	key := s.rotationKey()
	claimed := map[string]bool{}
	var files []*settingmountservice.File
	for _, e := range entries {
		entryFiles, err := entryFiles(key, e.Name, mounts[e.ID], byID[mounts[e.ID].Source.ID])
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		for _, file := range entryFiles {
			if claimed[file.Path] {
				continue
			}
			claimed[file.Path] = true
			files = append(files, file)
		}
	}
	return files, nil
}

// entryFiles is what one entry mounts: nothing when its source is missing,
// disabled or unusable, and none of the files naming a part the source does not
// offer or a path no file may have.
func entryFiles(
	key []byte, entryKey string, mount *entity.AppSettingMount, source *entity.Setting,
) ([]*settingmountservice.File, error) {
	if source == nil || source.Status != base.SettingStatusActive ||
		!settingmountservice.IsSourceType(source.Type) {
		return nil, nil
	}
	values, err := settingmountservice.Values(source)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if !settingmountservice.Usable(source.Type, values) {
		return nil, nil
	}
	files := make([]*settingmountservice.File, 0, len(mount.Files))
	seenParts := map[string]bool{}
	for _, f := range mount.Files {
		part := settingmountservice.PartOf(source.Type, f.Part)
		if part == nil || seenParts[f.Part] || !settingmountservice.ValidPath(f.Path) || part.Empty(values) {
			continue
		}
		seenParts[f.Part] = true
		data, err := part.RenderFrom(values)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		files = append(files, &settingmountservice.File{
			Entry:     entryKey,
			Part:      part.Name,
			Path:      f.Path,
			UID:       gofn.Coalesce(f.UID, defaultUID),
			GID:       gofn.Coalesce(f.GID, defaultGID),
			Mode:      gofn.Coalesce(f.Mode, defaultMode),
			Sensitive: part.Sensitive,
			Data:      data,
			Rotation:  settingmountservice.RotationKey(key, source.Type, part, values),
		})
	}
	return files, nil
}
