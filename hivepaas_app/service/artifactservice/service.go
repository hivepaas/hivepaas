// Package artifactservice puts a managed template's artifacts on the nodes that
// run its app: the file for the node's architecture, downloaded and kept by the
// node's agent. See docs/superpowers/specs/2026-10-17-managed-templates-design.md §2.
package artifactservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
)

type Service interface {
	// EnsureOnNode has the artifact on a node, in the build for its architecture,
	// and says where the host has it. A node whose architecture the artifact has
	// no build for is refused with ERR_ARTIFACT_ARCH_UNSUPPORTED, before its agent
	// is asked anything.
	EnsureOnNode(ctx context.Context, req *EnsureOnNodeReq) (*EnsureOnNodeResp, error)
}

type EnsureOnNodeReq struct {
	NodeID   string
	Artifact *templatemodel.Artifact
}

type EnsureOnNodeResp struct {
	// HostPath is the file as the host has it, which is what a bind mount names.
	HostPath   string
	Arch       base.CPUArch
	SHA256     string
	Downloaded bool
}
