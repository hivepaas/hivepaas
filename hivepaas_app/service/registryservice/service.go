// Package registryservice runs the registry HivePaaS provisions for itself: the
// app, the credential HivePaaS pushes with, and the questions the dashboard asks
// about both.
package registryservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

type Service interface {
	// Apply makes the cluster match the stored configuration: it provisions the
	// registry on the first save and reconciles it on the rest. It is
	// idempotent, so a failed save leaves work the next save retries.
	Apply(ctx context.Context, db database.IDB, req *SettingApplyReq) (*SettingApplyResp, error)

	// Status says what is running and what the registry holds, for the settings
	// screen. It never fails the screen: a registry that cannot be reached comes
	// back as a status saying so.
	Status(ctx context.Context, db database.IDB, setting *entity.Setting) (*Status, error)

	// RotateCredential issues a new password, keeping the old one working until
	// the grace period ends, because a service that is not redeployed still
	// presents the credential it was deployed with.
	RotateCredential(ctx context.Context, db database.IDB, req *RotateCredentialReq) (
		*RotateCredentialResp, error)

	// ProbeDomain reports what can be seen from outside about the registry's
	// address - a proxy in front of it, above all. It is advisory.
	ProbeDomain(ctx context.Context, domain string) (*DomainProbe, error)

	// CheckPush uploads a large blob through the public domain and throws it
	// away, which is the only way to find a body limit before a build does.
	CheckPush(ctx context.Context, db database.IDB, req *PushCheckReq) (*PushCheckResult, error)

	// Validate refuses a configuration before it is written. The usecase calls it
	// in PrepareUpdate, so a bad save is a validation error rather than a stored
	// configuration Apply then fails on.
	Validate(next, current *entity.RegistrySettings) error
}
