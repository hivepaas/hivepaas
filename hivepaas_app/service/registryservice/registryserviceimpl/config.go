// Package registryserviceimpl runs the system registry: it renders zot's
// configuration and the app document from the stored setting, provisions the app
// through the ordinary provisioning path, and keeps the credential HivePaaS
// pushes with.
package registryserviceimpl

import (
	"encoding/json"
	"fmt"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

const (
	registryPort     = "5000"
	registryRootDir  = "/var/lib/registry"
	registryConfDir  = "/etc/zot"
	registryS3Prefix = "/zot"

	// gcDelay shields a blob younger than itself from collection, so it has to
	// be longer than the slowest push. gcInterval is how often a pass starts; a
	// repository waits up to about twice that.
	registryGCDelay    = "2h"
	registryGCInterval = "1h"

	hoursPerDay = 24
)

// zotConfigInput is what the configuration needs that the setting does not hold.
type zotConfigInput struct {
	// S3 is nil for a registry on a volume.
	S3 *zotS3Input
}

type zotS3Input struct {
	Bucket    string
	Region    string
	Endpoint  string
	AccessKey string
	SecretKey string
	Secure    bool
}

// renderZotConfig writes the configuration file zot is started with.
//
// It is built as maps rather than a text template because two of its keys are
// conditional and one of them is a list whose length depends on the settings; a
// template with that many branches is a template nobody can read.
func renderZotConfig(cfg *entity.RegistrySettings, in zotConfigInput) ([]byte, error) {
	if cfg == nil {
		return nil, hperrors.Wrap(hperrors.ErrRegistryNotConfigured)
	}

	storage := map[string]any{
		"rootDirectory": registryRootDir,
		// zot refuses to start with dedupe on and a remote store unless a remote
		// cache is configured, which HivePaaS does not run.
		"dedupe":     cfg.Storage.Type != base.RegistryStorageTypeS3,
		"gc":         true,
		"gcDelay":    registryGCDelay,
		"gcInterval": registryGCInterval,
	}
	if cfg.Cleanup.Enabled {
		storage["retention"] = retentionPolicy(cfg.Cleanup)
	}
	if cfg.Storage.Type == base.RegistryStorageTypeS3 {
		if in.S3 == nil {
			return nil, hperrors.Wrap(hperrors.ErrRegistrySettingsInvalid).
				WithExtraDetail("The registry is set to S3, but no bucket was found for it.")
		}
		storage["storageDriver"] = map[string]any{
			"name":           "s3",
			"rootdirectory":  registryS3Prefix,
			"bucket":         in.S3.Bucket,
			"region":         in.S3.Region,
			"regionendpoint": in.S3.Endpoint,
			"secure":         in.S3.Secure,
			"forcepathstyle": true,
			"accesskey":      in.S3.AccessKey,
			"secretkey":      in.S3.SecretKey,
		}
	}

	doc := map[string]any{
		"distSpecVersion": "1.1.1",
		"storage":         storage,
		"http": map[string]any{
			"address": "0.0.0.0",
			"port":    registryPort,
			// Every daemon on the classic image store pushes docker v2s2
			// manifests, which zot answers 415 to without this.
			"compat": []string{"docker2s2"},
			"auth":   map[string]any{"htpasswd": map[string]any{"path": registryConfDir + "/htpasswd"}},
		},
		"log": map[string]any{"level": "info"},
		"extensions": map[string]any{
			// search answers with each repository's size and last update, which
			// is what the dashboard's status section reads.
			"search": map[string]any{"enable": true},
			"ui":     map[string]any{"enable": true},
		},
	}

	raw, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return raw, nil
}

// retentionPolicy turns the two numbers into zot's rules.
//
// The rules are a union: a tag survives if any of them keeps it. pulledWithin is
// there for the image a long-running service fetched when it was last
// rescheduled, which may be older than anything pushedWithin would save.
func retentionPolicy(cleanup entity.RegistryCleanup) map[string]any {
	window := fmt.Sprintf("%dh", cleanup.KeepDays*hoursPerDay)
	return map[string]any{
		"dryRun": false,
		"policies": []any{map[string]any{
			"repositories":    []string{"**"},
			"deleteUntagged":  true,
			"deleteReferrers": true,
			"keepTags": []any{
				keepTagsRule("mostRecentlyPushedCount", cleanup.KeepLast),
				keepTagsRule("pushedWithin", window),
				keepTagsRule("pulledWithin", window),
			},
		}},
	}
}

// keepTagsRule is one of zot's retention rules. They differ only in what limits
// them - a count, or a window - so the pattern they all apply to is written once.
func keepTagsRule(limit string, value any) map[string]any {
	return map[string]any{"patterns": []string{".*"}, limit: value}
}
