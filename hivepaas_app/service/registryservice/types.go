package registryservice

import (
	"context"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/systemappservice"
)

type SettingApplyReq struct {
	// Setting is the row that was just saved. Apply loads it itself when nil.
	Setting *entity.Setting
	// TriggerUserID is who asked, recorded on the first deployment the way the
	// template usecase records it. Empty when nothing asked - a reconcile that
	// runs on start-up, for instance.
	TriggerUserID string

	// Resources is what the registry's app runs under. It is not stored with the
	// settings: the app's service is where it lives, the same place the app's own
	// resource screen reads and writes. Nil leaves a running registry as it is,
	// and gives a new one the defaults.
	Resources *systemappservice.Resources

	// RemoveApp takes the registry's app and its service down. It is only read
	// when the configuration being applied is switched off, and switching off
	// without it is refused: the app has no screen of its own to be removed from,
	// so this save is the only place it can happen, and it has to be asked for.
	RemoveApp bool
	// RemoveStorage deletes the images along with the app. It is the app's own
	// directory inside the volume that goes, not the volume; a registry on S3
	// ignores it, because nothing here reaches into a bucket.
	RemoveStorage bool
}

type SettingApplyResp struct {
	// App is the registry's app, set whenever the registry is enabled.
	App *entity.App

	// DeploymentTask and CertTasks are created by provisioning and deliberately
	// left unscheduled: a task row can be picked up only once the transaction it
	// was written in has committed, and Apply runs inside the caller's. The
	// caller schedules them after the commit, the way the template usecase does.
	DeploymentTask *entity.Task
	CertTasks      []*entity.Task

	// Cleanup removes from docker what provisioning created there, for the caller
	// to run when its transaction does not commit. It is set when this call
	// provisioned the app, even if provisioning failed.
	Cleanup func(ctx context.Context) error

	// RemovedApp says the app was taken down by this call, and RemovedCredentialID
	// is the credential it was pushed with, which the caller deletes after the
	// transaction commits if nothing else still names it.
	RemovedApp          bool
	RemovedCredentialID string
}

type Status struct {
	// Provisioned says the app exists. Everything below is only meaningful then.
	Provisioned bool
	AppID       string
	// Reachable says the registry answered its own API. Unreachable is normal
	// while it restarts after a save.
	Reachable bool
	// Unreachable carries why, for the screen to show.
	Unreachable string
	// Repositories and StoredBytes come from zot's search extension, and are
	// zero when it could not be asked.
	Repositories int
	StoredBytes  int64
	// CredentialRotatedAt is when the password last changed, empty if never.
	CredentialRotatedAt time.Time
	// Resources is what the app's service runs under, or what a new registry
	// would get when there is none.
	Resources systemappservice.Resources
}

type RotateCredentialReq struct {
	Setting *entity.Setting
}

type RotateCredentialResp struct {
	// GraceEndsAt is when the previous password stops working.
	GraceEndsAt time.Time
}

type DomainProbe struct {
	// Proxied says a proxy answered instead of the registry.
	Proxied bool
	// Evidence names what said so: a header, or addresses that are not the
	// cluster's. It is what the dashboard shows; it is never a refusal.
	Evidence []string
	// Reached says the probe got an answer at all.
	Reached bool
}

type PushCheckReq struct {
	Setting *entity.Setting
	// Bytes is how large a blob to send. Zero means the default, which is above
	// the limit a proxy on a free plan imposes.
	Bytes int64
}

type PushCheckResult struct {
	OK bool
	// StatusCode is what the registry, or whatever is in front of it, answered.
	StatusCode int
	// Detail is the sentence the dashboard shows.
	Detail string
	// Elapsed is how long the upload took, which is the other thing an operator
	// wants to know about a proxy.
	Elapsed time.Duration
}
