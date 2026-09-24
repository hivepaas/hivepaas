package specmodel

import (
	"strings"
	"time"
)

// Issue codes of the import plan, beside the ones issue.go declares for both
// halves.
const (
	CodeKeyMismatch         = "KEY_MISMATCH"
	CodeNameInUse           = "NAME_IN_USE"
	CodeSettingVersionNewer = "SETTING_VERSION_NEWER"
	CodeTypeNotImportable   = "TYPE_NOT_IMPORTABLE"

	CodeDomainInUse             = "DOMAIN_IN_USE"
	CodePortInUse               = "PORT_IN_USE"
	CodeNodeNotFound            = "NODE_NOT_FOUND"
	CodeStorageNotEmpty         = "STORAGE_NOT_EMPTY"
	CodeStorageUnchecked        = "STORAGE_UNCHECKED"
	CodeCapabilityNotPermitted  = "CAPABILITY_NOT_PERMITTED"
	CodeSharedMountNotPermitted = "SHARED_MOUNT_NOT_PERMITTED"
	CodeSecretOmitted           = "SECRET_OMITTED"
	CodeOwnerNotPermitted       = "OWNER_NOT_PERMITTED"
)

// Note codes. A note says what import does, and needs no acceptance.
const (
	CodeSecretGenerated = "SECRET_GENERATED"
	CodeCredentialKept  = "CREDENTIAL_KEPT" //nolint:gosec // an issue code, not a credential
	CodeOwnerNotFound   = "OWNER_NOT_FOUND"
)

// SeverityWarning clears nothing, but the result may not be what the operator
// expects, so applying still needs their acceptance.
const SeverityWarning Severity = "warning"

// Selection is the part of a bundle an import takes: node paths, each naming
// its node and everything below it. A segment may be "*". Exclude wins, and an
// empty Include selects everything.
type Selection struct {
	Include []string `json:"include"`
	Exclude []string `json:"exclude"`
}

// Selects reports whether the node at path is part of the selection.
func (s Selection) Selects(path string) bool {
	for _, pattern := range s.Exclude {
		if covers(pattern, path) {
			return false
		}
	}
	if len(s.Include) == 0 {
		return true
	}
	for _, pattern := range s.Include {
		if covers(pattern, path) {
			return true
		}
	}
	return false
}

// covers reports whether a selection path names the node at path or one of its
// ancestors: every segment of the pattern matches the path's segment there.
func covers(pattern, path string) bool {
	want := strings.Split(strings.Trim(pattern, "/"), "/")
	have := strings.Split(path, "/")
	if len(want) > len(have) {
		return false
	}
	for i, segment := range want {
		if segment != "*" && segment != have[i] {
			return false
		}
	}
	return true
}

// Existing says what import does with an object the installation already has.
type Existing string

const (
	// ExistingUpdate makes it match the bundle.
	ExistingUpdate Existing = "update"
	// ExistingKeep leaves it alone: only what is missing is created.
	ExistingKeep Existing = "keep"
)

// ImportOptions are the choices an operator makes on the plan screen.
type ImportOptions struct {
	Existing            Existing `json:"existing"`
	DeployCreated       bool     `json:"deployCreated"`
	DeployChangedSource bool     `json:"deployChangedSource"`
}

// NodeKind is what a plan node stands for.
type NodeKind string

const (
	NodeKindGlobal   NodeKind = "global"
	NodeKindProject  NodeKind = "project"
	NodeKindEnv      NodeKind = "env"
	NodeKindSettings NodeKind = "settings"
	NodeKindApp      NodeKind = "app"
)

// NodeAction is what import does to a node.
type NodeAction string

const (
	ActionCreate    NodeAction = "create"
	ActionUpdate    NodeAction = "update"
	ActionUnchanged NodeAction = "unchanged"
	// ActionKeep is an object the installation has that the operator chose to
	// leave alone.
	ActionKeep NodeAction = "keep"
	// ActionSkip is an object an issue keeps out of the import.
	ActionSkip NodeAction = "skip"
)

// MatchedBy says how a node found the object it stands for.
type MatchedBy string

const (
	MatchedByID  MatchedBy = "id"
	MatchedByKey MatchedBy = "key"
)

// PlanNode is one node of the plan tree.
type PlanNode struct {
	Path       string     `json:"path"`
	Kind       NodeKind   `json:"kind"`
	Key        string     `json:"key,omitempty"`
	Name       string     `json:"name,omitempty"`
	Selected   bool       `json:"selected"`
	SelectedBy string     `json:"selectedBy,omitempty"`
	Action     NodeAction `json:"action"`
	MatchedBy  MatchedBy  `json:"matchedBy,omitempty"`
	// TargetID is the object matched on this installation; it stays on the server.
	TargetID string   `json:"-"`
	Changes  []string `json:"changes,omitempty"`
	Restart  bool     `json:"restart"`
	Deploy   bool     `json:"deploy"`
	Issues   []Issue  `json:"issues,omitempty"`
	Notes    []Issue  `json:"notes,omitempty"`
}

// BundleInfo is what the plan says about the bundle itself.
type BundleInfo struct {
	APIVersion       string      `json:"apiVersion"`
	Scope            string      `json:"scope"`
	ExportedAt       time.Time   `json:"exportedAt"`
	SourceAppVersion string      `json:"sourceAppVersion"`
	SecretsMode      SecretsMode `json:"secretsMode"`
}

// ImportPlan is what validate answers, and what apply is bound to by PlanHash.
type ImportPlan struct {
	Bundle   BundleInfo     `json:"bundle"`
	Nodes    []*PlanNode    `json:"nodes"`
	Summary  map[string]int `json:"summary"`
	PlanHash string         `json:"planHash"`
}
