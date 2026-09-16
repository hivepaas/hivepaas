package specmodel

import "github.com/hivepaas/hivepaas/hivepaas_app/base"

// singletonBlockNames maps every setting type that can hold at most one setting
// per scope to the name its block takes in a spec document.
//
// These need no key: the block name is the identity. That is not a convenience.
// env-var and app-routing settings carry no name at all in a real installation
// - they were the only two types with an empty name in every row - so a key
// derived from the name would be the empty string.
//
// A type belongs here if it is read through SettingRepo.GetSingle with no extra
// filters, or through the setting_*_unique.go usecases. Two types look like
// singletons and are not: notification calls GetSingle with
// `is_default = TRUE`, and sched-job calls it with a kind and a target filter.
// Both hold many rows per scope. TestNotificationAndSchedJobAreCollections
// guards that distinction, because getting it wrong loses every row but one.
var singletonBlockNames = map[base.SettingType]string{
	base.SettingTypeApp:           "app",
	base.SettingTypeAppClone:      "clone",
	base.SettingTypeAppDeployment: "source", // lifted into deployment.source
	base.SettingTypeAppFeatures:   "features",
	base.SettingTypeAppKind:       "kind",
	base.SettingTypeAppPlacement:  "placement",
	base.SettingTypeAppRouting:    "routing",
	base.SettingTypeEnvVar:        "envVars",
	base.SettingTypeProject:       "project",

	base.SettingTypeBackupRepoCleanup: "backupRepoCleanup",
	base.SettingTypeDomainSettings:    "domainSettings",
	base.SettingTypeHivePaaSService:   "hivepaasService",
	base.SettingTypeImageBuild:        "imageBuild",
	base.SettingTypeLogging:           "logging",
	base.SettingTypeSSLRenewal:        "sslRenewal",
	base.SettingTypeSystemBackup:      "systemBackup",
	base.SettingTypeSystemCleanup:     "systemCleanup",
	base.SettingTypeTraefikConfig:     "traefikConfig",
	base.SettingTypeTraefikService:    "traefikService",
}

// collectionBlockNames names the block each collection type's keyed map takes.
var collectionBlockNames = map[base.SettingType]string{
	base.SettingTypeAccessToken:     "accessTokens",
	base.SettingTypeAcmeDnsProvider: "acmeDnsProviders",
	base.SettingTypeAPIKey:          "apiKeys",
	base.SettingTypeBackupRepo:      "backupRepos",
	base.SettingTypeBackupSnapshot:  "backupSnapshots",
	base.SettingTypeBasicAuth:       "basicAuths",
	base.SettingTypeCloudStorage:    "cloudStorages",
	base.SettingTypeClusterNetwork:  "networks",
	base.SettingTypeClusterNode:     "nodes",
	base.SettingTypeClusterVolume:   "volumes",
	base.SettingTypeCommandPipe:     "commandPipes",
	base.SettingTypeCommandTemplate: "commandTemplates",
	base.SettingTypeConfigFile:      "configFiles",
	base.SettingTypeEmail:           "emails",
	base.SettingTypeGithubApp:       "githubApps",
	base.SettingTypeIMService:       "imServices",
	base.SettingTypeNotification:    "notifications",
	base.SettingTypeOAuth:           "oauths",
	base.SettingTypePeriodicJob:     "periodicJobs",
	base.SettingTypeRegistryAuth:    "registryAuths",
	base.SettingTypeRepoWebhook:     "repoWebhooks",
	base.SettingTypeSchedJob:        "schedJobs",
	base.SettingTypeScript:          "scripts",
	base.SettingTypeSecret:          "secrets",
	base.SettingTypeSSHKey:          "sshKeys",
	base.SettingTypeSSLCert:         "sslCerts",
	base.SettingTypeSSLProvider:     "sslProviders",
}

// IsSingletonType reports whether a type holds at most one setting per scope.
func IsSingletonType(typ base.SettingType) bool {
	_, ok := singletonBlockNames[typ]
	return ok
}

// SingletonBlockName is the YAML key a singleton type occupies. Empty if the
// type is not a singleton.
func SingletonBlockName(typ base.SettingType) string { return singletonBlockNames[typ] }

// CollectionBlockName is the YAML key a collection type's keyed map occupies.
// Empty if the type is not a collection.
func CollectionBlockName(typ base.SettingType) string { return collectionBlockNames[typ] }
