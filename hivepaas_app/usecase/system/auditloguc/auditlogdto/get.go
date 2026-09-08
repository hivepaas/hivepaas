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
	Meta *basedto.Meta       `json:"meta"`
	Data *AuditLogDetailResp `json:"data"`
}

// AuditLogDetailResp is one entry with its Detail.
//
// Detail is the free-form half of a record - what a capability revoke took away,
// which fields a settings change touched - and it is only worth sending when
// somebody has asked for that one entry. The listing leaves it out on purpose:
// it is unbounded, and a page of them is a large response nobody reads.
type AuditLogDetailResp struct {
	AuditLogResp

	Detail string `json:"detail,omitempty"`
}

type AuditLogResp struct {
	ID     string              `json:"id"`
	Type   base.AuditLogType   `json:"type"`
	Source base.AuditLogSource `json:"source,omitempty"`
	Result base.AuditLogResult `json:"result"`

	Actor    *ActorResp    `json:"actor"`
	Resource *ResourceResp `json:"resource,omitempty"`

	ViaAPIKey  bool   `json:"viaApiKey,omitempty"`
	SessionUID string `json:"sessionUid,omitempty"`
	ClientIP   string `json:"clientIp,omitempty"`
	RemoteAddr string `json:"remoteAddr,omitempty"`
	UserAgent  string `json:"userAgent,omitempty"`
	RequestID  string `json:"requestId,omitempty"`

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

	resp.Actor = TransformAuditLogActor(auditLog, refObjects)
	resp.Resource = TransformAuditLogResource(auditLog, refObjects)

	return resp, nil
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
) (resp *AuditLogDetailResp, err error) {
	item, err := TransformAuditLog(auditLog, refObjects)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &AuditLogDetailResp{AuditLogResp: *item, Detail: auditLog.Detail}, nil
}
