package entity

import (
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/base"
)

const (
	CurrentFileVersion = 1
)

var _ = registerSettingParser(base.SettingTypeFile, &fileParser{})

type fileParser struct {
}

func (s *fileParser) New() SettingData {
	return &File{}
}

type File struct {
	FileKind    base.FileKind        `json:"fileKind"`
	StorageType base.FileStorageType `json:"storageType"`
	Storage     ObjectID             `json:"storage,omitzero"`
	Bucket      string               `json:"bucket,omitempty"`
	Mimetype    string               `json:"mimetype"`
	Size        int64                `json:"size"`
	Path        string               `json:"path"`
	Deleted     bool                 `json:"deleted,omitempty"`
}

func (s *File) GetType() base.SettingType {
	return base.SettingTypeFile
}

func (s *File) GetRefObjectIDs() *RefObjectIDs {
	refIDs := &RefObjectIDs{}
	return refIDs
}

func (s *File) IsInLocalStorage() bool {
	return s.StorageType == base.FileStorageLocal
}

func (s *File) IsInCloudStorage() bool {
	return s.StorageType == base.FileStorageCloud
}

func (s *File) Migrate(setting *Setting) (hasChange bool, err error) {
	if CurrentFileVersion == setting.Version {
		return false, nil
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentFileVersion
	setting.MustSetData(s)
	return true, nil
}

func (s *Setting) AsFile() (*File, error) {
	return parseSettingAs[*File](s)
}

func (s *Setting) MustAsFile() *File {
	return gofn.Must(s.AsFile())
}
