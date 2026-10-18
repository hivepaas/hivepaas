// Package artifactservice puts a managed template's artifacts on the nodes that
// run its app: the file for the node's architecture, downloaded and kept by the
// node's agent. See docs/superpowers/specs/2026-10-17-managed-templates-design.md §2.
package artifactservice

import (
	"context"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

type Service interface {
	// EnsureOnNode has the artifact on a node, in the build for its architecture,
	// and says where the host has it. A node whose architecture the artifact has
	// no build for is refused with ERR_ARTIFACT_ARCH_UNSUPPORTED, before its agent
	// is asked anything.
	EnsureOnNode(ctx context.Context, req *EnsureOnNodeReq) (*EnsureOnNodeResp, error)

	// ArtifactsOf is an app's artifacts setting, or nil when it has none.
	ArtifactsOf(ctx context.Context, db database.IDB, appID string) (*entity.AppArtifactsSettings, error)

	// MountsFor has the artifacts on every node a task of spec could start on -
	// the Ready, active nodes its placement constraints allow - and returns the
	// mounts that bind them. One bind names one file, so nodes that would need
	// different builds are refused with ERR_ARTIFACT_ARCH_UNSUPPORTED before any
	// is asked; so is a placement no node satisfies, as ERR_ARTIFACT_DOWNLOAD_FAILED.
	// No artifacts asks nothing and returns none.
	MountsFor(ctx context.Context, artifacts *entity.AppArtifactsSettings, spec *swarm.ServiceSpec) (
		[]mount.Mount, error)

	// ReconcileAll has every app's artifacts on the nodes its tasks may start on
	// as its service stands: a node that joined after it was deployed, or one
	// whose cache lost them. It changes no service. Each app is had on its own:
	// one that fails is in the response, and the others go on.
	ReconcileAll(ctx context.Context, db database.IDB) (*ReconcileAllResp, error)
}

type ReconcileAllResp struct {
	// Apps is how many apps had their artifacts where their tasks may start.
	Apps int
	// Failed are the apps that did not, each with why.
	Failed []*ReconcileFailure
}

type ReconcileFailure struct {
	App *entity.App
	Err error
}

type EnsureOnNodeReq struct {
	NodeID   string
	Artifact *entity.AppArtifact
}

type EnsureOnNodeResp struct {
	// HostPath is the file as the host has it, which is what a bind mount names.
	HostPath   string
	Arch       base.CPUArch
	SHA256     string
	Downloaded bool
}
