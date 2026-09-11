package auditlogdto

import (
	"time"

	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/copier"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/projecthelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appuc/appdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/projectuc/projectdto"
)

type GetAuditLogReq struct {
	ID    string              `json:"-"`
	Scope *entity.ObjectScope `json:"-" mapstructure:"-"`
}

func NewGetAuditLogReq() *GetAuditLogReq {
	return &GetAuditLogReq{}
}

func (req *GetAuditLogReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ID, true, "id")...)
	// TODO: add validation
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetAuditLogResp struct {
	Meta *basedto.Meta `json:"meta"`
	Data *AuditLogResp `json:"data"`
}

type AuditLogResp struct {
	ID     string              `json:"id"`
	Type   base.AuditLogType   `json:"type"`
	Source base.AuditLogSource `json:"source,omitempty"`
	// Section survives in the listing where it used to be lost: it lived inside
	// Detail, which the listing drops once it passes a hundred characters.
	Section string              `json:"section,omitempty"`
	Result  base.AuditLogResult `json:"result"`

	ScopeProject *projectdto.ProjectBaseResp `json:"scopeProject,omitempty"`
	ScopeApp     *appdto.AppBaseResp         `json:"scopeApp,omitempty"`
	ScopeUser    *basedto.UserBaseResp       `json:"scopeUser,omitempty"`

	Actor    *ActorResp    `json:"actor"`
	Resource *ResourceResp `json:"resource,omitempty"`

	ViaAPIKey  bool   `json:"viaApiKey,omitempty"`
	SessionUID string `json:"sessionUid,omitempty"`
	ClientIP   string `json:"clientIp,omitempty"`
	RemoteAddr string `json:"remoteAddr,omitempty"`
	UserAgent  string `json:"userAgent,omitempty"`
	RequestID  string `json:"requestId,omitempty"`

	Detail    string `json:"detail,omitempty"`
	HasDetail bool   `json:"hasDetail,omitempty"`

	CreatedAt time.Time `json:"createdAt"`
}

type ActorResp struct {
	ID         string           `json:"id"`
	Type       base.SubjectType `json:"type"`
	Name       string           `json:"name,omitempty"`
	LoggedName string           `json:"loggedName,omitempty"`
	Photo      string           `json:"photo,omitempty"`
}

type ResourceResp struct {
	ID         string            `json:"id"`
	Type       base.ResourceType `json:"type"`
	Name       string            `json:"name,omitempty"`
	LoggedName string            `json:"loggedName,omitempty"`
	Photo      string            `json:"photo,omitempty"`
}

func TransformAuditLog(
	auditLog *entity.AuditLog,
	refObjects *entity.RefObjects,
) (resp *AuditLogResp, err error) {
	if err = copier.Copy(&resp, &auditLog); err != nil {
		return nil, hperrors.Wrap(err)
	}
	if refObjects == nil {
		refObjects = entity.NewRefObjects()
	}

	TransformAuditLogScopeObject(auditLog, refObjects, resp)
	resp.Actor = TransformAuditLogActor(auditLog, refObjects)
	resp.Resource = TransformAuditLogResource(auditLog, refObjects)

	// Return detail along with the item if it's not too big
	if len(auditLog.Detail) < 100 { //nolint:mnd
		resp.Detail = auditLog.Detail
	}
	resp.HasDetail = auditLog.Detail != ""

	return resp, nil
}

func TransformAuditLogScopeObject(
	auditLog *entity.AuditLog,
	refObjects *entity.RefObjects,
	resp *AuditLogResp,
) {
	if auditLog.ObjectID == "" {
		return
	}

	var projectID, appID, userID string
	switch auditLog.Scope {
	case base.ObjectScopeProject:
		projectID = auditLog.ObjectID
	case base.ObjectScopeProjectEnv:
		projectID, _ = projecthelper.ParseProjectEnvID(auditLog.ObjectID)
	case base.ObjectScopeApp:
		appID = auditLog.ObjectID
	case base.ObjectScopeUser:
		userID = auditLog.ObjectID
	case base.ObjectScopeGlobal, base.ObjectScopeHivepaas:
	}

	if appID != "" {
		ref := refObjects.RefApps[appID]
		refResp := appdto.TransformAppBase(ref)
		if refResp == nil {
			refResp = appdto.NewMissingApp(appID)
		}
		resp.ScopeApp = refResp
		if ref != nil && projectID == "" {
			projectID = ref.ProjectID
		}
	}
	if projectID != "" {
		refResp := projectdto.TransformProjectBase(refObjects.RefProjects[projectID])
		if refResp == nil {
			refResp = projectdto.NewMissingProject(projectID)
		}
		resp.ScopeProject = refResp
	}
	if userID != "" {
		refResp := basedto.TransformUserBase(refObjects.RefUsers[userID])
		if refResp == nil {
			refResp = basedto.NewMissingUser(userID)
		}
		resp.ScopeUser = refResp
	}
}

func TransformAuditLogActor(
	auditLog *entity.AuditLog,
	refObjects *entity.RefObjects,
) (resp *ActorResp) {
	if auditLog.ActorID == "" {
		return nil
	}
	resp = &ActorResp{
		ID:         auditLog.ActorID,
		Type:       auditLog.ActorType,
		LoggedName: auditLog.ActorName,
	}
	// Only users are looked up today. Every other actor type keeps the name that
	// was written into the record at the time, which is the point of storing it:
	// the object may have been renamed or deleted since, and the entry still has
	// to say who acted.
	if auditLog.ActorType == base.SubjectTypeUser {
		if refUser := refObjects.RefUsers[auditLog.ActorID]; refUser != nil {
			resp.Name = gofn.Coalesce(refUser.FullName, refUser.Username)
			resp.Photo = basedto.TransformObjectIcon(refUser.Photo)
		}
	}

	return resp
}

func TransformAuditLogResource(
	auditLog *entity.AuditLog,
	refObjects *entity.RefObjects,
) (resp *ResourceResp) {
	if auditLog.ResID == "" {
		return nil
	}
	resp = &ResourceResp{
		ID:         auditLog.ResID,
		Type:       auditLog.ResType,
		LoggedName: auditLog.ResName,
	}
	// As with the actor: only users are enriched, and everything else falls back to
	// the name recorded at the time.
	if auditLog.ResType == base.ResourceTypeUser {
		if refUser := refObjects.RefUsers[auditLog.ResID]; refUser != nil {
			resp.Name = gofn.Coalesce(refUser.FullName, refUser.Username)
			resp.Photo = basedto.TransformObjectIcon(refUser.Photo)
		}
	}

	return resp
}

// TransformAuditLogDetail is TransformAuditLog plus the Detail the listing omits.
func TransformAuditLogDetail(
	auditLog *entity.AuditLog,
	refObjects *entity.RefObjects,
) (resp *AuditLogResp, err error) {
	item, err := TransformAuditLog(auditLog, refObjects)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	item.Detail = auditLog.Detail
	item.HasDetail = item.Detail != ""
	return item, nil
}
