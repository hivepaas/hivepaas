package envvarserviceimpl

import (
	"context"
	"sort"
	"strconv"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/settinghelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
)

func (s *service) BuildSystemEnvVarsInApp(
	ctx context.Context,
	db database.IDB,
	req *envvarservice.BuildSystemEnvVarsInAppReq,
) ([]*envvarservice.EnvVar, error) {
	settings, _, err := s.settingRepo.List(ctx, db, nil, nil,
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
		bunex.SelectWhereIn("setting.type IN (?)", base.SettingTypeAppRouting, base.SettingTypeAppKind),
		bunex.SelectWhere("setting.object_id = ?", req.App.ID),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	routingSetting := settinghelper.FindSettingByType(settings, base.SettingTypeAppRouting)
	var routingSettings *entity.AppRoutingSettings
	if routingSetting != nil {
		routingSettings = routingSetting.MustAsAppRoutingSettings()
	}

	portStr := ""
	activeDomain := ""
	appURL := ""
	if routingSettings != nil {
		if routingSettings.Port > 0 {
			portStr = strconv.Itoa(routingSettings.Port)
		}
		activeDomain = gofn.FirstOr(routingSettings.GetActiveDomainNames(), "")
		appURL = routingSettings.GetAppURL()
	}

	kindSetting := settinghelper.FindSettingByType(settings, base.SettingTypeAppKind)
	var kindSettings *entity.AppKindSettings
	if kindSetting != nil {
		kindSettings = kindSetting.MustAsAppKindSettings()
	}

	result := []*envvarservice.EnvVar{
		{
			EnvVar: &entity.EnvVar{
				Key:      base.AppSystemEnvVarHost,
				Value:    req.App.Key,
				IsShared: true,
			},
		},
		{
			EnvVar: &entity.EnvVar{
				Key:      base.AppSystemEnvVarPort,
				Value:    portStr,
				IsShared: true,
			},
		},
		{
			EnvVar: &entity.EnvVar{
				Key:      base.AppSystemEnvVarDomain,
				Value:    activeDomain,
				IsShared: true,
			},
		},
		{
			EnvVar: &entity.EnvVar{
				Key:      base.AppSystemEnvVarAppURL,
				Value:    appURL,
				IsShared: true,
			},
		},
		{
			EnvVar: &entity.EnvVar{
				Key:      base.AppSystemEnvVarEnv,
				Value:    req.App.ProjectEnv.Name,
				IsShared: true,
			},
		},
		{
			EnvVar: &entity.EnvVar{
				Key:      base.AppSystemEnvVarName,
				Value:    req.App.Name,
				IsShared: true,
			},
		},
		{
			EnvVar: &entity.EnvVar{
				Key:      base.AppSystemEnvVarID,
				Value:    req.App.ID,
				IsShared: true,
			},
		},
	}

	kindEnvs, err := kindEnvVars(kindSettings)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	result = append(result, kindEnvs...)

	for _, env := range result {
		env.IsLiteral = true
		env.IsSystem = true
	}

	// Mask the secrets if instructed
	if req.MaskSecrets {
		for _, env := range result {
			if base.IsAppSecretEnv(env.Key) {
				env.Value = basedto.MaskedSecret
			}
		}
	}

	if req.Sort {
		sort.Slice(result, func(i, j int) bool {
			return result[i].Key < result[j].Key
		})
	}

	return result, nil
}

// kindEnvVars is the part of an app's system environment that comes from its
// kind: the credentials other apps connect with. It is a function of the kind
// setting alone, so the rules of who may read what are readable in one place.
//
// A webapp contributes nothing, and neither does a kind whose sub-block is
// missing - an app half configured is not a reason to publish empty credentials
// other apps would then refer to.
func kindEnvVars(kindSettings *entity.AppKindSettings) ([]*envvarservice.EnvVar, error) {
	if kindSettings == nil {
		return nil, nil
	}

	switch {
	case kindSettings.Category == base.AppCategoryDatabase && kindSettings.Database != nil:
		if err := kindSettings.Decrypt(); err != nil {
			return nil, hperrors.Wrap(err)
		}
		db := kindSettings.Database
		return []*envvarservice.EnvVar{
			sharedEnv(base.AppSystemEnvVarUser, db.Username),
			sharedEnv(base.AppSystemEnvVarPassword, gofn.Must(db.Password.GetPlain())),
			{
				EnvVar: &entity.EnvVar{
					Key:      base.AppSystemEnvVarRootPassword,
					Value:    gofn.Must(db.RootPassword.GetPlain()),
					IsShared: false, // NOTE: do not share this env with other apps
				},
			},
			sharedEnv(base.AppSystemEnvVarDatabaseName, db.DbName),
			sharedEnv(base.AppSystemEnvVarSSLMode, string(db.SSLMode)),
		}, nil

	case kindSettings.Category == base.AppCategoryStorage && kindSettings.Storage != nil:
		if err := kindSettings.Decrypt(); err != nil {
			return nil, hperrors.Wrap(err)
		}
		// A store authenticates with a pair - an S3 access key and secret key, or a
		// user and password over HTTP - and a client needs both halves plus somewhere
		// to put the objects. Both halves are shared: unlike a database's root
		// password, the pair is the only way in, so an app that may use the store at
		// all needs all of it.
		store := kindSettings.Storage
		return []*envvarservice.EnvVar{
			sharedEnv(base.AppSystemEnvVarKeyID, store.KeyID),
			sharedEnv(base.AppSystemEnvVarSecret, gofn.Must(store.Secret.GetPlain())),
			sharedEnv(base.AppSystemEnvVarBucket, store.Bucket),
			sharedEnv(base.AppSystemEnvVarRegion, store.Region),
		}, nil

	case kindSettings.Category == base.AppCategoryCache && kindSettings.Cache != nil:
		if err := kindSettings.Decrypt(); err != nil {
			return nil, hperrors.Wrap(err)
		}
		// Only the password is shared. A cache has no user to name - Redis and
		// Valkey authenticate as the built-in `default` user, and Memcached has no
		// accounts at all - no database to select, and no root account whose
		// password would have to be kept out of other apps' reach.
		//
		// The variable is published even when the password is empty, which is
		// how a cache with authentication turned off looks: an app referring to
		// ${cache.HIVEPAAS_PASSWORD} then reads an empty value rather than
		// failing on a variable that does not exist.
		//
		// The three that follow are the cache's own tuning. They are published for
		// the app itself and for nobody else, so that what the App Kind screen says
		// about memory, eviction and persistence is what the server is started with
		// - a template's command reads them, and editing them here restarts the app
		// with the new ones. An empty value is passed through as empty: it means
		// "whatever the engine's own default is", which the command decides.
		//
		// The ceiling goes out as a plain byte count rather than "256mb", because
		// what reads it is a shell: Redis takes bytes as they are, and an engine
		// that wants another unit - memcached counts in megabytes - can divide.
		cache := kindSettings.Cache
		return []*envvarservice.EnvVar{
			sharedEnv(base.AppSystemEnvVarPassword, gofn.Must(cache.Password.GetPlain())),
			ownEnv(base.AppSystemEnvVarMaxMemory, strconv.FormatInt(int64(cache.MaxMemory), 10)),
			ownEnv(base.AppSystemEnvVarEvictionRule, cache.EvictionRule),
			ownEnv(base.AppSystemEnvVarPersistenceMode, cache.PersistenceMode),
		}, nil

	default:
		return nil, nil
	}
}

// ownEnv is a variable the app reads itself and no other app may name.
func ownEnv(key, value string) *envvarservice.EnvVar {
	return &envvarservice.EnvVar{
		EnvVar: &entity.EnvVar{
			Key:      key,
			Value:    value,
			IsShared: false,
		},
	}
}

func sharedEnv(key, value string) *envvarservice.EnvVar {
	return &envvarservice.EnvVar{
		EnvVar: &entity.EnvVar{
			Key:      key,
			Value:    value,
			IsShared: true,
		},
	}
}
