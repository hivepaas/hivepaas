package entity

import "github.com/hivepaas/hivepaas/hivepaas_app/base"

// SpecDecision is a policy's answer about one setting.
type SpecDecision struct {
	Export bool
	// Reason explains a refusal, and is what the export report prints.
	Reason string
}

// SpecPolicy decides how one setting type appears in a configuration spec.
//
// It is registered beside the parser, and for the same reason: a setting type
// that gains a policy here cannot be forgotten by the exporter, and one that
// forgets to register is refused rather than exported blindly.
type SpecPolicy interface {
	// Decide reports whether this particular setting belongs in a spec.
	Decide(setting *Setting) SpecDecision
	// Strip removes values the target regenerates, or that would trigger an
	// action on import. It must not remove values that are merely specific to
	// the installation being exported: those are what make a restore exact.
	Strip(data SettingData)
}

var specPolicyMap = make(map[base.SettingType]SpecPolicy, 46) //nolint:mnd

// can happen in a package-level var block.
//
//nolint:unparam // mirrors registerSettingParser: the result exists so registration
func registerSpecPolicy(typ base.SettingType, policy SpecPolicy) bool {
	specPolicyMap[typ] = policy
	return true
}

func registerDefaultSpecPolicies(types ...base.SettingType) bool {
	for _, typ := range types {
		specPolicyMap[typ] = defaultSpecPolicy{}
	}
	return true
}

// SpecPolicyFor returns the policy for a type, or nil if none is registered.
func SpecPolicyFor(typ base.SettingType) SpecPolicy {
	return specPolicyMap[typ]
}

// SpecExportDecision is the exporter's entry point. An unregistered type is
// refused, so adding a setting type without thinking about the spec produces a
// report entry rather than a silent leak.
func SpecExportDecision(setting *Setting) SpecDecision {
	policy := SpecPolicyFor(setting.Type)
	if policy == nil {
		return SpecDecision{Reason: "no spec policy is registered for this setting type"}
	}
	return policy.Decide(setting)
}

// defaultSpecPolicy exports the setting whole and strips nothing.
type defaultSpecPolicy struct{}

func (defaultSpecPolicy) Decide(*Setting) SpecDecision { return SpecDecision{Export: true} }
func (defaultSpecPolicy) Strip(SettingData)            {}

// skipSpecPolicy never exports.
type skipSpecPolicy struct{ reason string }

func (p skipSpecPolicy) Decide(*Setting) SpecDecision { return SpecDecision{Reason: p.reason} }
func (skipSpecPolicy) Strip(SettingData)              {}

// clusterSpecPolicy exports only the rows HivePaaS authored.
//
// Every cluster type exists in two flavors. networks_sync.go, volumes_sync.go
// and docker_node_sync.go all hardcode ObjectScopeGlobal and reuse the Docker
// object id as the setting id; those rows are rediscovered by the next sync and
// mean nothing on another installation - worse, sync soft-deletes any row whose
// Docker object is missing, so importing one elsewhere is undone immediately.
//
// The rows created by networkservice/project_env.go and volumeservice/project.go
// sit at project or project-env scope with a real ULID, and record which project
// owns what. No sync can reconstruct that, because no sync ever writes a
// non-global scope.
//
// This also settles cluster-node without a special case: nothing outside
// docker_node_sync.go creates one, so there are no non-global rows to export.
type clusterSpecPolicy struct{}

func (clusterSpecPolicy) Decide(setting *Setting) SpecDecision {
	if setting.Scope == base.ObjectScopeGlobal {
		return SpecDecision{Reason: "discovered by cluster sync; recreated automatically"}
	}
	return SpecDecision{Export: true}
}

func (clusterSpecPolicy) Strip(SettingData) {}

// appRoutingSpecPolicy clears the reset flag, which is a command rather than
// configuration: importing Reset: true performs a reset.
type appRoutingSpecPolicy struct{}

func (appRoutingSpecPolicy) Decide(*Setting) SpecDecision { return SpecDecision{Export: true} }

func (appRoutingSpecPolicy) Strip(data SettingData) {
	if routing, ok := data.(*AppRoutingSettings); ok {
		routing.Reset = false
	}
}

// loggingSpecPolicy drops the endpoints of a managed backend. The setting's own
// comment says they are "derived from the deployed service when Managed, and
// ignored", so carrying them makes every spec disagree with the system as soon
// as it is applied.
type loggingSpecPolicy struct{}

func (loggingSpecPolicy) Decide(*Setting) SpecDecision { return SpecDecision{Export: true} }

func (loggingSpecPolicy) Strip(data SettingData) {
	logging, ok := data.(*LoggingSettings)
	if !ok || !logging.Backend.Managed {
		return
	}
	logging.Backend.Ingest = nil
	logging.Backend.Query = nil
}

// Nothing here strips Secret.SwarmRef, ConfigFile.SwarmRef, env-var entries
// flagged IsSystem, an image digest, or cluster-volume.NodeID. Each is specific
// to the installation rather than regenerated, and each is what makes a restore
// onto the same installation exact. The env-var case is the sharpest:
// HIVEPAAS_ROOT_PASSWORD is generated once and the database volume was
// initialized with it, so regenerating it on restore leaves the application
// holding a password the database will not accept.
var (
	_ = registerSpecPolicy(base.SettingTypeAPIKey, skipSpecPolicy{
		reason: "the secret is a one-way hash and authenticates nobody once imported"})
	_ = registerSpecPolicy(base.SettingTypeBackupSnapshot, skipSpecPolicy{
		reason: "backup history, rediscovered by scanning the repository"})
	_ = registerSpecPolicy(base.SettingTypeApp, skipSpecPolicy{
		reason: "a declared type with no parser and no rows"})

	_ = registerSpecPolicy(base.SettingTypeClusterNetwork, clusterSpecPolicy{})
	_ = registerSpecPolicy(base.SettingTypeClusterVolume, clusterSpecPolicy{})
	_ = registerSpecPolicy(base.SettingTypeClusterNode, clusterSpecPolicy{})

	_ = registerSpecPolicy(base.SettingTypeAppRouting, appRoutingSpecPolicy{})
	_ = registerSpecPolicy(base.SettingTypeLogging, loggingSpecPolicy{})

	// Everything else is exported whole. Keeping this list explicit rather than
	// defaulting to "export" is what makes a new setting type show up in the
	// report instead of appearing in specs unnoticed.
	_ = registerDefaultSpecPolicies(
		base.SettingTypeAccessToken,
		base.SettingTypeAcmeDnsProvider,
		base.SettingTypeAppClone,
		base.SettingTypeAppDeployment,
		base.SettingTypeAppFeatures,
		base.SettingTypeAppKind,
		base.SettingTypeAppPlacement,
		base.SettingTypeBackupRepo,
		base.SettingTypeBackupRepoCleanup,
		base.SettingTypeBasicAuth,
		base.SettingTypeCloudStorage,
		base.SettingTypeCommandPipe,
		base.SettingTypeCommandTemplate,
		base.SettingTypeConfigFile,
		base.SettingTypeDomainSettings,
		base.SettingTypeEmail,
		base.SettingTypeEnvVar,
		base.SettingTypeGithubApp,
		base.SettingTypeHivePaaSService,
		base.SettingTypeIMService,
		base.SettingTypeImageBuild,
		base.SettingTypeNotification,
		base.SettingTypeOAuth,
		base.SettingTypePeriodicJob,
		base.SettingTypeProject,
		base.SettingTypeRegistryAuth,
		base.SettingTypeRepoWebhook,
		base.SettingTypeSSHKey,
		base.SettingTypeSSLCert,
		base.SettingTypeSSLProvider,
		base.SettingTypeSSLRenewal,
		base.SettingTypeSchedJob,
		base.SettingTypeScript,
		base.SettingTypeSecret,
		base.SettingTypeSystemBackup,
		base.SettingTypeSystemCleanup,
		base.SettingTypeTraefikConfig,
		base.SettingTypeTraefikService,
	)
)
