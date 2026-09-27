package schedjobdto

import (
	"path"
	"regexp"
	"strings"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/copier"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/commandtemplateuc/commandtemplatedto"
)

const (
	maxDataBackupTags         = 20
	maxDataBackupSubpathLen   = 500
	dataBackupReservedTagsKey = "hivepaas."
)

var (
	// dataBackupFileNameRegex is a file name without a directory: the snapshot's
	// one file.
	dataBackupFileNameRegex = regexp.MustCompile(`^[A-Za-z0-9._-]{1,100}$`)
	// A tag is kopia's key:value: the key holds no ':', neither holds a space.
	dataBackupTagKeyRegex   = regexp.MustCompile(`^[A-Za-z0-9._-]{1,50}$`)
	dataBackupTagValueRegex = regexp.MustCompile(`^[A-Za-z0-9._:/@+=-]{1,100}$`)
)

// SchedJobDataBackupReq is what a data-backup job backs up, and where to.
type SchedJobDataBackupReq struct {
	Source base.SchedJobDataBackupSource `json:"source"`
	// SourceCommand runs in the job's app; its stdout is the file SourceFileName.
	SourceCommand  *commandtemplatedto.CommandTemplateBaseReq `json:"sourceCommand"`
	SourceFileName string                                     `json:"sourceFileName"`
	// SourceVolume is a volume the app mounts; SourceVolumeSubpath a path inside
	// what the app sees of it, "" for all of it.
	SourceVolume        basedto.ObjectIDReq `json:"sourceVolume"`
	SourceVolumeSubpath string              `json:"sourceVolumeSubpath"`
	TargetRepository    basedto.ObjectIDReq `json:"targetRepository"`
	Tags                map[string]string   `json:"tags"`
}

func (req *SchedJobDataBackupReq) ToEntity() *entity.SchedJobDataBackup {
	if req == nil {
		return nil
	}
	res := &entity.SchedJobDataBackup{
		Source:           req.Source,
		TargetRepository: entity.ObjectID{ID: req.TargetRepository.ID},
		Tags:             req.Tags,
	}
	switch req.Source {
	case base.SchedJobDataBackupSourceCommand:
		res.SourceCommand = req.SourceCommand.ToEntity()
		res.SourceFileName = req.SourceFileName
	case base.SchedJobDataBackupSourceVolume:
		res.SourceVolume = entity.ObjectID{ID: req.SourceVolume.ID}
		res.SourceVolumeSubpath = req.SourceVolumeSubpath
	}
	if len(res.Tags) == 0 {
		res.Tags = nil
	}
	return res
}

func (req *SchedJobDataBackupReq) modifyRequest() error {
	if req == nil {
		return nil
	}
	req.SourceFileName = strings.TrimSpace(req.SourceFileName)
	req.SourceVolumeSubpath = strings.TrimSpace(req.SourceVolumeSubpath)
	if req.SourceVolumeSubpath != "" && !strings.HasPrefix(req.SourceVolumeSubpath, "/") {
		req.SourceVolumeSubpath = strings.TrimSuffix(path.Clean(req.SourceVolumeSubpath), "/")
		if req.SourceVolumeSubpath == "." {
			req.SourceVolumeSubpath = ""
		}
	}
	if req.SourceCommand != nil {
		req.SourceCommand.Name = "-"
		req.SourceCommand.Kind = ""
		// Its stdout is the backup: a TTY would mix stderr into it.
		req.SourceCommand.TTY = false
		if err := req.SourceCommand.ModifyRequest(); err != nil {
			return hperrors.Wrap(err)
		}
	}
	return nil
}

func (req *SchedJobDataBackupReq) validate(field string) (res []vld.Validator) {
	res = append(res, basedto.ValidateCond(req != nil, field)...)
	if req == nil {
		return res
	}
	field += "."
	res = append(res, basedto.ValidateStrIn(&req.Source, true, base.AllSchedJobDataBackupSources, field+"source")...)
	isCommand := req.Source == base.SchedJobDataBackupSourceCommand
	isVolume := req.Source == base.SchedJobDataBackupSourceVolume

	res = append(res, basedto.ValidateCond(isCommand == (req.SourceCommand != nil), field+"sourceCommand")...)
	if req.SourceCommand != nil {
		res = append(res, req.SourceCommand.Validate(field+"sourceCommand")...)
	}
	res = append(res, basedto.ValidateCond(isCommand == dataBackupFileNameRegex.MatchString(req.SourceFileName),
		field+"sourceFileName")...)

	res = append(res, basedto.ValidateObjectIDReq(&req.SourceVolume, isVolume, field+"sourceVolume")...)
	res = append(res, basedto.ValidateCond(isVolume || req.SourceVolume.ID == "", field+"sourceVolume")...)
	res = append(res, basedto.ValidateCond(isVolume || req.SourceVolumeSubpath == "",
		field+"sourceVolumeSubpath")...)
	res = append(res, basedto.ValidateCond(isRelativeSubpath(req.SourceVolumeSubpath),
		field+"sourceVolumeSubpath")...)

	res = append(res, basedto.ValidateObjectIDReq(&req.TargetRepository, true, field+"targetRepository")...)
	res = append(res, basedto.ValidateCond(validDataBackupTags(req.Tags), field+"tags")...)
	return res
}

// isRelativeSubpath is a path inside a directory: not from the root, never
// above where it starts.
func isRelativeSubpath(subpath string) bool {
	if subpath == "" {
		return true
	}
	if len(subpath) > maxDataBackupSubpathLen || strings.HasPrefix(subpath, "/") {
		return false
	}
	cleaned := path.Clean(subpath)
	return cleaned != ".." && !strings.HasPrefix(cleaned, "../")
}

func validDataBackupTags(tags map[string]string) bool {
	if len(tags) > maxDataBackupTags {
		return false
	}
	for key, value := range tags {
		if !dataBackupTagKeyRegex.MatchString(key) || strings.HasPrefix(key, dataBackupReservedTagsKey) ||
			!dataBackupTagValueRegex.MatchString(value) {
			return false
		}
	}
	return true
}

// validateDataBackupFields is what a job's type says about its data backup: a
// data-backup has one, runs its source's command only and is its app's; any
// other type has none.
func (req *SchedJobBaseReq) validateDataBackupFields(field string) (res []vld.Validator) {
	if req.JobType != base.SchedJobTypeDataBackup {
		return basedto.ValidateCond(req.DataBackup == nil, field+"dataBackup")
	}
	res = append(res, req.DataBackup.validate(field+"dataBackup")...)
	res = append(res, basedto.ValidateCond(req.Command == nil, field+"command")...)
	res = append(res, basedto.ValidateCond(req.CommandOutput == nil, field+"commandOutput")...)
	res = append(res, basedto.ValidateObjectIDReq(&req.App, true, field+"app")...)
	return res
}

type SchedJobDataBackupResp struct {
	Source              base.SchedJobDataBackupSource           `json:"source"`
	SourceCommand       *commandtemplatedto.CommandTemplateResp `json:"sourceCommand,omitempty"`
	SourceFileName      string                                  `json:"sourceFileName,omitempty"`
	SourceVolume        *settings.BaseSettingResp               `json:"sourceVolume,omitempty"`
	SourceVolumeSubpath string                                  `json:"sourceVolumeSubpath,omitempty"`
	TargetRepository    *settings.BaseSettingResp               `json:"targetRepository"`
	Tags                map[string]string                       `json:"tags,omitempty"`
}

// TransformSchedJobDataBackup is a data backup with its volume and repository
// named, from the settings refObjects holds; "missing" for one that is gone.
func TransformSchedJobDataBackup(
	dataBackup *entity.SchedJobDataBackup,
	refObjects *entity.RefObjects,
) (*SchedJobDataBackupResp, error) {
	if dataBackup == nil {
		return nil, nil //nolint:nilnil // not a data backup
	}
	resp := &SchedJobDataBackupResp{
		Source:              dataBackup.Source,
		SourceFileName:      dataBackup.SourceFileName,
		SourceVolumeSubpath: dataBackup.SourceVolumeSubpath,
		TargetRepository:    namedSetting(refObjects, dataBackup.TargetRepository.ID, base.SettingTypeBackupRepo),
		Tags:                dataBackup.Tags,
	}
	if dataBackup.SourceVolume.ID != "" {
		resp.SourceVolume = namedSetting(refObjects, dataBackup.SourceVolume.ID, base.SettingTypeClusterVolume)
	}
	if dataBackup.SourceCommand != nil {
		if err := copier.Copy(&resp.SourceCommand, dataBackup.SourceCommand); err != nil {
			return nil, hperrors.Wrap(err)
		}
		commandtemplatedto.TransformScript(&dataBackup.SourceCommand.Script, refObjects, resp.SourceCommand)
	}
	return resp, nil
}

func namedSetting(refObjects *entity.RefObjects, id string, typ base.SettingType) *settings.BaseSettingResp {
	var setting *entity.Setting
	if refObjects != nil {
		setting = refObjects.RefSettings[id]
	}
	resp, _ := settings.TransformSettingBase(setting)
	if resp == nil {
		resp = settings.NewMissingSetting(id, typ)
	}
	return resp
}
