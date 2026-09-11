package auditserviceimpl

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/reqinfo"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
)

type service struct {
	auditLogRepo repository.AuditLogRepo
}

func New(
	auditLogRepo repository.AuditLogRepo,
) auditservice.Service {
	return &service{
		auditLogRepo: auditLogRepo,
	}
}

// Record implements auditservice.Service.
func (s *service) Record(ctx context.Context, db database.IDB, entry *auditservice.Entry) error {
	if entry == nil {
		return hperrors.NewArgumentInvalid("audit entry")
	}

	auditLog := &entity.AuditLog{
		ID:        gofn.Must(ulid.NewStringULID()),
		Scope:     entry.Scope,
		ObjectID:  entry.ObjectID,
		Type:      entry.Type,
		Source:    entry.Source,
		Result:    entry.Result,
		ResType:   entry.ResType,
		ResID:     entry.ResID,
		ResName:   entry.ResName,
		Detail:    entry.Detail,
		CreatedAt: timeutil.NowUTC(),
	}
	applyActor(auditLog, entry.Auth)
	applyRequest(auditLog, ctx)

	if err := s.auditLogRepo.Insert(ctx, db, auditLog); err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

// applyActor copies who acted. Every step is guarded because an audit write must
// not be the thing that panics: it runs on paths that are already refusing a
// request, where the auth may be less complete than the happy path assumes.
func applyActor(auditLog *entity.AuditLog, auth *basedto.Auth) {
	auditLog.ActorType = base.SubjectTypeUser
	if auth == nil || auth.User == nil {
		return
	}
	if user := auth.User.Entity(); user != nil {
		auditLog.ActorID = user.ID
		auditLog.ActorName = user.Username
	}
	if claims := auth.User.AuthClaims; claims != nil {
		auditLog.SessionUID = claims.UID
		auditLog.ViaAPIKey = claims.IsAPIKey
		if auditLog.ActorID == "" {
			auditLog.ActorID = claims.UserID
		}
	}
}

// applyRequest copies where the action came from. Work started outside a request
// - a worker, a scheduled job - carries none of this, and that is not an error.
func applyRequest(auditLog *entity.AuditLog, ctx context.Context) {
	info := reqinfo.From(ctx)
	if info == nil {
		return
	}
	auditLog.RequestID = info.RequestID
	auditLog.ClientIP = info.ClientIP
	auditLog.RemoteAddr = info.RemoteAddr
	auditLog.UserAgent = info.UserAgent
}
