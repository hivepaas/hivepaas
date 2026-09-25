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
	files, _, err := s.resolve(ctx, db, app)
	return files, err
}

func (s *service) EntryStates(
	ctx context.Context, db database.IDB, app *entity.App,
) (map[string]*settingmountservice.EntryState, error) {
	_, states, err := s.resolve(ctx, db, app)
	return states, err
}

// resolve is the files the app should have, and for each of its entries what of
// it is mounted and why nothing is when nothing is.
func (s *service) resolve(
	ctx context.Context, db database.IDB, app *entity.App,
) ([]*settingmountservice.File, map[string]*settingmountservice.EntryState, error) {
	all, err := s.loadEntries(ctx, db, app.ID)
	if err != nil {
		return nil, nil, hperrors.Wrap(err)
	}
	states := make(map[string]*settingmountservice.EntryState, len(all))
	entries := make([]*entity.Setting, 0, len(all))
	for _, e := range all {
		state := &settingmountservice.EntryState{Mounted: []string{}}
		states[e.ID] = state
		switch {
		case e.Status != base.SettingStatusActive:
			state.Reason = settingmountservice.ReasonEntryDisabled
		case !settingmountservice.ValidEntryKey(e.Name):
			state.Reason = settingmountservice.ReasonKeyInvalid
		default:
			entries = append(entries, e)
		}
	}
	// Key order, so that the first claim to a path is the same one every time.
	slices.SortFunc(entries, func(a, b *entity.Setting) int { return strings.Compare(a.Name, b.Name) })

	mounts := make(map[string]*entity.AppSettingMount, len(entries))
	sourceIDs := make([]string, 0, len(entries))
	for _, e := range entries {
		mount, err := e.AsAppSettingMount()
		if err != nil {
			return nil, nil, hperrors.Wrap(err)
		}
		mounts[e.ID] = mount
		sourceIDs = append(sourceIDs, mount.Source.ID)
	}
	sourceIDs = gofn.ToSet(gofn.ToSliceSkippingZero(sourceIDs...))
	byID := map[string]*entity.Setting{}
	if len(sourceIDs) > 0 {
		sources, err := s.loadSources(ctx, db, app, sourceIDs)
		if err != nil {
			return nil, nil, hperrors.Wrap(err)
		}
		for _, source := range sources {
			byID[source.ID] = source
		}
	}

	key := s.rotationKey()
	claimed := map[string]bool{}
	var files []*settingmountservice.File
	for _, e := range entries {
		entryFiles, reason, err := entryFiles(key, e.Name, mounts[e.ID], byID[mounts[e.ID].Source.ID])
		if err != nil {
			return nil, nil, hperrors.Wrap(err)
		}
		state := states[e.ID]
		state.Reason = reason
		for _, file := range entryFiles {
			if claimed[file.Path] {
				continue
			}
			claimed[file.Path] = true
			files = append(files, file)
			state.Mounted = append(state.Mounted, file.Path)
		}
		if len(entryFiles) > 0 && len(state.Mounted) == 0 {
			state.Reason = settingmountservice.ReasonPathsTaken
		}
	}
	return files, states, nil
}

// entryFiles is what one entry mounts: nothing, with the reason, when its
// source is missing, disabled or unusable, and none of the files naming a part
// the source does not offer or a path no file may have.
func entryFiles(
	key []byte, entryKey string, mount *entity.AppSettingMount, source *entity.Setting,
) ([]*settingmountservice.File, string, error) {
	if source == nil || source.Status != base.SettingStatusActive ||
		!settingmountservice.IsSourceType(source.Type) {
		return nil, settingmountservice.ReasonSourceUnavailable, nil
	}
	values, err := settingmountservice.Values(source)
	if err != nil {
		return nil, "", hperrors.Wrap(err)
	}
	if !settingmountservice.Usable(source.Type, values) {
		return nil, settingmountservice.ReasonSourceIncomplete, nil
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
			return nil, "", hperrors.Wrap(err)
		}
		files = append(files, &settingmountservice.File{
			Entry:    entryKey,
			Part:     part.Name,
			Path:     f.Path,
			UID:      gofn.Coalesce(f.UID, defaultUID),
			GID:      gofn.Coalesce(f.GID, defaultGID),
			Mode:     gofn.Coalesce(f.Mode, defaultMode),
			Secret:   part.Secret,
			Data:     data,
			Rotation: settingmountservice.RotationKey(key, source.Type, part, values),
		})
	}
	return files, "", nil
}
