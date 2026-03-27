package appdto

import (
	"time"

	"github.com/docker/docker/api/types/swarm"
	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/copier"
)

type GetAppReq struct {
	ProjectID string `json:"-"`
	AppID     string `json:"-"`
	GetStats  bool   `json:"-" mapstructure:"getStats"`
}

func NewGetAppReq() *GetAppReq {
	return &GetAppReq{}
}

func (req *GetAppReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectID")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appID")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetAppResp struct {
	Meta *basedto.Meta `json:"meta"`
	Data *AppResp      `json:"data"`
}

type AppResp struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Key       string         `json:"key"`
	Status    base.AppStatus `json:"status"`
	Token     string         `json:"token"`
	Note      string         `json:"note"`
	Tags      []string       `json:"tags" copy:"-"` // manual copy AppTag -> string
	UpdateVer int            `json:"updateVer"`

	// Stats of app, only returns when req.getStats=true
	Stats *AppStatsResp `json:"stats"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type AppUserAccessResp struct {
	*basedto.UserBaseResp
	Access base.AccessActions `json:"access"`
}

type AppStatsResp struct {
	RunningTasks   int `json:"runningTasks"`
	DesiredTasks   int `json:"desiredTasks"`
	CompletedTasks int `json:"completedTasks"`
}

type AppBaseResp struct {
	ID     string         `json:"id"`
	Name   string         `json:"name"`
	Key    string         `json:"key"`
	Status base.AppStatus `json:"status"`
}

type AppTransformationInput struct {
	SwarmServiceMap map[string]*swarm.Service
}

func TransformApp(app *entity.App, input *AppTransformationInput) (resp *AppResp, err error) {
	if err = copier.Copy(&resp, &app); err != nil {
		return nil, apperrors.Wrap(err)
	}
	resp.Tags = gofn.MapSlice(app.Tags, func(t *entity.AppTag) string { return t.Tag })
	resp.Stats = TransformAppStats(app, input)
	return resp, nil
}

func TransformAppStats(app *entity.App, input *AppTransformationInput) *AppStatsResp {
	if input == nil || input.SwarmServiceMap == nil {
		return nil
	}
	service := input.SwarmServiceMap[app.ID]
	if service == nil || service.ServiceStatus == nil {
		return nil
	}
	//nolint
	return &AppStatsResp{
		RunningTasks:   int(service.ServiceStatus.RunningTasks),
		DesiredTasks:   int(service.ServiceStatus.DesiredTasks),
		CompletedTasks: int(service.ServiceStatus.CompletedTasks),
	}
}

func TransformAppsBase(apps []*entity.App) []*AppBaseResp {
	return gofn.MapSlice(apps, func(app *entity.App) *AppBaseResp {
		return &AppBaseResp{
			ID:     app.ID,
			Name:   app.Name,
			Key:    app.Key,
			Status: app.Status,
		}
	})
}
