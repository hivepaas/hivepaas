package appdto

import (
	"time"

	"github.com/moby/moby/api/types/swarm"
	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/copier"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/projectuc/projectdto"
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
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetAppResp struct {
	Meta *basedto.Meta `json:"meta"`
	Data *AppResp      `json:"data"`
}

type AppResp struct {
	ID        string                      `json:"id"`
	Name      string                      `json:"name"`
	Project   *projectdto.ProjectBaseResp `json:"project"`
	ParentApp *AppBaseResp                `json:"parentApp"`
	Key       string                      `json:"key"`
	LocalKey  string                      `json:"localKey"`
	Status    base.AppStatus              `json:"status"`
	Env       string                      `json:"env"`
	Note      string                      `json:"note"`
	Tags      []string                    `json:"tags" copy:"-"` // manual copy AppTag -> string
	UpdateVer int                         `json:"updateVer"`

	// Stats of app, only returns when req.getStats=true
	Stats *AppStatsResp `json:"stats"`

	// AccessLinks external links to access the app
	AccessLinks []string `json:"accessLinks,omitempty"`

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
	ID       string         `json:"id"`
	Name     string         `json:"name"`
	Key      string         `json:"key"`
	LocalKey string         `json:"localKey"`
	Status   base.AppStatus `json:"status"`
	Env      string         `json:"env"`
}

type AppTransformationInput struct {
	SwarmServiceMap map[string]*swarm.Service
}

func TransformApp(app *entity.App, input *AppTransformationInput) (resp *AppResp, err error) {
	if app == nil {
		return nil, nil
	}
	if err = copier.Copy(&resp, &app); err != nil {
		return nil, apperrors.New(err)
	}
	resp.Tags = gofn.MapSlice(app.Tags, func(t *entity.AppTag) string { return t.Tag })
	resp.Stats = TransformAppStats(app, input)
	resp.AccessLinks = TransformAppAccessLinks(app)
	if app.ParentID != "" {
		resp.ParentApp = gofn.Coalesce(TransformAppBase(app.ParentApp), &AppBaseResp{ID: app.ParentID})
	} else {
		resp.ParentApp = nil
	}
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

func TransformAppAccessLinks(app *entity.App) (resp []string) {
	setting := app.GetSettingByType(base.SettingTypeAppHttp)
	if setting == nil {
		return nil
	}
	for _, domain := range setting.MustAsAppHttpSettings().GetActiveDomainNames() {
		resp = append(resp, "https://"+domain)
	}
	return resp
}

func TransformAppBase(app *entity.App) *AppBaseResp {
	if app == nil {
		return nil
	}
	return &AppBaseResp{
		ID:       app.ID,
		Name:     app.Name,
		Key:      app.Key,
		LocalKey: app.LocalKey,
		Status:   app.Status,
		Env:      app.Env,
	}
}

func TransformAppsBase(apps []*entity.App) []*AppBaseResp {
	return gofn.MapSlice(apps, TransformAppBase)
}
