package specserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// importPolicy is how import writes the settings of one type in a scope - the
// global one, a project's, an env's. An app's settings are built with the app.
//
// This is the export registry's safety rule turned around: no setting type is
// imported because somebody forgot it could not be. Writing a setting through
// its screen can do more than write its row - schedule tasks, initialize a
// repository, restart a service HivePaaS runs - and a type whose writing import
// does not reproduce yet is skipped with the reason, rather than written as a
// row that looks right and does nothing.
type importPolicy struct {
	// skip is why the type is not imported; empty for one that is.
	skip string
	// afterCommit does what writing settings of the type takes beyond their rows,
	// once the rows are committed. Nil for a type whose row is all there is.
	afterCommit func(s *service, settings []*entity.Setting) error
}

const (
	reasonSchedulesTasks = "writing it schedules tasks, which import does not do yet"
	reasonRunsAService   = "it configures a service HivePaaS runs, which import does not restart yet"
)

var importPolicies = map[base.SettingType]importPolicy{
	// A row is all there is to these.
	base.SettingTypeAccessToken:     {},
	base.SettingTypeAcmeDnsProvider: {},
	base.SettingTypeAppClone:        {},
	base.SettingTypeAppDeployment:   {},
	base.SettingTypeAppFeatures:     {},
	base.SettingTypeAppKind:         {},
	base.SettingTypeAppPlacement:    {},
	base.SettingTypeAppRouting:      {},
	base.SettingTypeAppTemplate:     {},
	base.SettingTypeBasicAuth:       {},
	base.SettingTypeCloudStorage:    {},
	base.SettingTypeClusterVolume:   {},
	base.SettingTypeCommandPipe:     {},
	base.SettingTypeCommandTemplate: {},
	base.SettingTypeConfigFile:      {},
	base.SettingTypeDomainSettings:  {},
	base.SettingTypeEmail:           {},
	base.SettingTypeEnvVar:          {},
	base.SettingTypeGithubApp:       {},
	base.SettingTypeIMService:       {},
	base.SettingTypeImageBuild:      {},
	base.SettingTypeNotification:    {},
	base.SettingTypeOAuth:           {},
	base.SettingTypeProject:         {},
	base.SettingTypeRegistryAuth:    {},
	base.SettingTypeRepoWebhook:     {},
	base.SettingTypeSSHKey:          {},
	base.SettingTypeSSLProvider:     {},
	base.SettingTypeScript:          {},
	base.SettingTypeSecret:          {},

	base.SettingTypeSSLCert: {afterCommit: func(s *service, settings []*entity.Setting) error {
		// Traefik reads certificates from files, not from the database.
		return hperrors.Wrap(s.sslService.WriteCertFiles(true, settings...))
	}},

	base.SettingTypeSchedJob:          {skip: reasonSchedulesTasks},
	base.SettingTypePeriodicJob:       {skip: reasonSchedulesTasks},
	base.SettingTypeBackupRepoCleanup: {skip: reasonSchedulesTasks},
	base.SettingTypeSSLRenewal:        {skip: reasonSchedulesTasks},
	base.SettingTypeSystemBackup:      {skip: reasonSchedulesTasks},
	base.SettingTypeSystemCleanup:     {skip: reasonSchedulesTasks},
	base.SettingTypeBackupRepo: {
		skip: "writing it initializes the repository, which import does not do yet",
	},
	base.SettingTypeLogging:         {skip: reasonRunsAService},
	base.SettingTypeRegistry:        {skip: reasonRunsAService},
	base.SettingTypeTraefikConfig:   {skip: reasonRunsAService},
	base.SettingTypeTraefikService:  {skip: reasonRunsAService},
	base.SettingTypeHivePaaSService: {skip: reasonRunsAService},
	base.SettingTypeClusterNetwork: {
		skip: "a network is created with an env's first app, under this installation's name",
	},

	// Export never writes these; a bundle holding one was written by hand.
	base.SettingTypeApp:            {skip: reasonNeverExported},
	base.SettingTypeAPIKey:         {skip: reasonNeverExported},
	base.SettingTypeBackupSnapshot: {skip: reasonNeverExported},
	base.SettingTypeClusterNode:    {skip: reasonNeverExported},
}

const reasonNeverExported = "export never writes this type"

// importPolicyFor is the policy of a type. A type with none is not imported: a
// setting type added later has to be thought about before import writes it.
func importPolicyFor(typ base.SettingType) importPolicy {
	policy, found := importPolicies[typ]
	if !found {
		return importPolicy{skip: "no import policy covers this type"}
	}
	return policy
}
