package backupsnapshotdto

import (
	"strings"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

const (
	maxFilterValues = 50
	maxTagLen       = 200
)

type ListBackupSnapshotReq struct {
	Scope *entity.ObjectScope `json:"-" mapstructure:"-"`

	RepoIDs []string `json:"-" mapstructure:"repo"`
	AppIDs  []string `json:"-" mapstructure:"app"`
	// Tags are key:value, all of which a snapshot must carry.
	Tags     []string      `json:"-" mapstructure:"tag"`
	FromDate timeutil.Date `json:"-" mapstructure:"fromDate"`
	ToDate   timeutil.Date `json:"-" mapstructure:"toDate"`
	Search   string        `json:"-" mapstructure:"search"`

	Paging basedto.Paging `json:"-"`
}

func NewListBackupSnapshotReq() *ListBackupSnapshotReq {
	return &ListBackupSnapshotReq{}
}

// PagingReq is where a handler parses the page asked for.
func (req *ListBackupSnapshotReq) PagingReq() *basedto.Paging {
	return &req.Paging
}

func (req *ListBackupSnapshotReq) ModifyRequest() error {
	req.Search = strings.TrimSpace(req.Search)
	// Newest first, always: a snapshot's time is not a column to sort by.
	req.Paging.Sort = nil
	return nil
}

func (req *ListBackupSnapshotReq) Validate() hperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateIDSlice(req.RepoIDs, true, 0, "repo")...)
	validators = append(validators, basedto.ValidateIDSlice(req.AppIDs, true, 0, "app")...)
	validators = append(validators, basedto.ValidateCond(len(req.RepoIDs) <= maxFilterValues &&
		len(req.AppIDs) <= maxFilterValues && len(req.Tags) <= maxFilterValues, "repo|app|tag")...)
	for _, tag := range req.Tags {
		key, _, found := strings.Cut(tag, ":")
		validators = append(validators, basedto.ValidateCond(found && key != "" && len(tag) <= maxTagLen, "tag")...)
	}
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type ListBackupSnapshotResp struct {
	Meta *basedto.ListMeta     `json:"meta"`
	Data []*BackupSnapshotResp `json:"data"`
	// Repos are the repositories the view reaches, for its filter and their Sync.
	Repos []*settings.BaseSettingResp `json:"repos"`
}
