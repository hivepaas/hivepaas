package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/fileutil"
)

const (
	CurrentAppSettingMountVersion = 1
)

var _ = registerSettingParser(base.SettingTypeAppSettingMount, &appSettingMountParser{})

type appSettingMountParser struct {
}

func (s *appSettingMountParser) New() SettingData {
	return &AppSettingMount{}
}

// AppSettingMount mounts parts of another setting - a certificate and its key, a
// basic auth pair as htpasswd - as files in the app's containers, which follow
// that setting from then on. The entry's key is the setting's name. See
// docs/superpowers/specs/2026-09-25-setting-mounts-design.md.
type AppSettingMount struct {
	// Source is the setting the files come from, every one of them.
	Source ObjectID               `json:"source"`
	Files  []*AppSettingMountFile `json:"files"`
}

// AppSettingMountFile puts one part of the source at a path. UID, GID and Mode
// are those of secrets and config files when empty.
type AppSettingMountFile struct {
	Part string            `json:"part"`
	Path string            `json:"path"`
	UID  string            `json:"uid,omitempty"`
	GID  string            `json:"gid,omitempty"`
	Mode fileutil.FileMode `json:"mode,omitempty"`
}

func (s *AppSettingMount) GetType() base.SettingType {
	return base.SettingTypeAppSettingMount
}

func (s *AppSettingMount) GetRefObjectIDs() *RefObjectIDs {
	refIDs := &RefObjectIDs{}
	if s.Source.ID != "" {
		refIDs.RefSettingIDs = append(refIDs.RefSettingIDs, s.Source.ID)
	}
	return refIDs
}

func (s *AppSettingMount) GetResourceLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().GetResourceLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *Setting) AsAppSettingMount() (*AppSettingMount, error) {
	return parseSettingAs[*AppSettingMount](s)
}

func (s *Setting) MustAsAppSettingMount() *AppSettingMount {
	return gofn.Must(s.AsAppSettingMount())
}
