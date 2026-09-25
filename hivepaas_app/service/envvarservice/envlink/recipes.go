package envlink

import (
	"sort"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

const (
	FamilyPostgres  = "postgres"
	FamilyMySQL     = "mysql"
	FamilyMongoDB   = "mongodb"
	FamilyRedis     = "redis"
	FamilyMemcached = "memcached"
)

var families = map[string]string{
	"postgres": FamilyPostgres, "postgresql": FamilyPostgres, "timescaledb": FamilyPostgres,
	"postgis": FamilyPostgres, "pgvector": FamilyPostgres,
	"mysql": FamilyMySQL, "mariadb": FamilyMySQL, "percona": FamilyMySQL,
	"mongodb": FamilyMongoDB, "mongo": FamilyMongoDB, "ferretdb": FamilyMongoDB,
	"redis": FamilyRedis, "valkey": FamilyRedis, "keydb": FamilyRedis, "dragonfly": FamilyRedis,
	"memcached": FamilyMemcached,
}

// EngineFamily is the family an engine belongs to, or "" for one the table does
// not know.
func EngineFamily(engine string) string {
	return families[strings.ToLower(strings.TrimSpace(engine))]
}

// refs writes references to one app's variables.
type refs struct{ app string }

func (r refs) of(name string) string { return "${" + r.app + "." + name + "}" }

// Suggest is what to add to link to target, most useful first, the target's own
// shared variables last.
func Suggest(target *Target, facts *Facts) []*Group {
	r := refs{app: target.Key}
	var groups []*Group
	switch target.Category {
	case base.AppCategoryDatabase:
		groups = databaseGroups(r, EngineFamily(target.Engine))
	case base.AppCategoryCache:
		groups = cacheGroups(r, EngineFamily(target.Engine), facts.HasPassword)
	case base.AppCategoryStorage:
		groups = []*Group{storageGroup(r)}
	case base.AppCategoryWebapp:
		groups = []*Group{addressGroup(r, target)}
	default:
		address := addressGroup(r, target)
		address.Warnings = append(address.Warnings,
			target.Name+" declares no kind, so only its addresses and shared variables are offered.")
		groups = []*Group{address}
	}
	if !facts.HasPort {
		portRef := r.of(base.AppSystemEnvVarPort)
		for _, g := range groups {
			for _, v := range g.Vars {
				if strings.Contains(v.Value, portRef) {
					g.Warnings = append(g.Warnings, target.Name+" has no container port; set one in its "+
						"Routing Settings, or the port below is empty.")
					break
				}
			}
		}
	}
	if shared := sharedGroup(r, target, facts.SharedVars); shared != nil {
		groups = append(groups, shared)
	}
	return groups
}

func urlGroup(key, value string) *Group {
	return &Group{
		ID: "connection-url", Title: "Connection string", Recommended: true,
		Description: "One URL most client libraries accept.",
		Vars:        []*Var{{Key: key, Value: value, Description: "Connection URL"}},
	}
}

func individualGroup(recommended bool, vars ...*Var) *Group {
	return &Group{
		ID: "individual", Title: "Individual variables", Recommended: recommended,
		Description: "Each part on its own, for a client configured part by part.",
		Vars:        vars,
	}
}

func databaseGroups(r refs, family string) []*Group {
	host, port := r.of(base.AppSystemEnvVarHost), r.of(base.AppSystemEnvVarPort)
	user, password := r.of(base.AppSystemEnvVarUser), r.of(base.AppSystemEnvVarPassword)
	name, sslMode := r.of(base.AppSystemEnvVarDatabaseName), r.of(base.AppSystemEnvVarSSLMode)
	auth := user + ":" + r.of(base.AppSystemEnvVarPasswordURLEncoded) + "@" + host + ":" + port

	switch family {
	case FamilyPostgres:
		return []*Group{
			urlGroup("DATABASE_URL", "postgres://"+auth+"/"+name+"?sslmode="+sslMode),
			individualGroup(false, append(parts("DB_", host, port, user, password, name),
				&Var{Key: "DB_SSL_MODE", Value: sslMode})...),
			{
				ID: "libpq", Title: "libpq variables",
				Description: "What libpq and most Postgres drivers read with no configuration.",
				Vars: []*Var{
					{Key: "PGHOST", Value: host}, {Key: "PGPORT", Value: port}, {Key: "PGUSER", Value: user},
					{Key: "PGPASSWORD", Value: password}, {Key: "PGDATABASE", Value: name},
					{Key: "PGSSLMODE", Value: sslMode},
				},
			},
		}
	case FamilyMySQL:
		return []*Group{
			urlGroup("DATABASE_URL", "mysql://"+auth+"/"+name),
			individualGroup(false, parts("DB_", host, port, user, password, name)...),
		}
	case FamilyMongoDB:
		return []*Group{
			urlGroup("MONGODB_URI", "mongodb://"+auth+"/"+name+"?authSource=admin"),
			individualGroup(false, parts("MONGO_", host, port, user, password, name)...),
		}
	}
	return []*Group{individualGroup(true, parts("DB_", host, port, user, password, name)...)}
}

// parts are a database's connection parts one by one, each key prefixed; Mongo
// clients call the name DATABASE rather than NAME.
func parts(prefix, host, port, user, password, name string) []*Var {
	nameKey := "NAME"
	if prefix == "MONGO_" {
		nameKey = "DATABASE"
	}
	return []*Var{
		{Key: prefix + "HOST", Value: host}, {Key: prefix + "PORT", Value: port},
		{Key: prefix + "USER", Value: user}, {Key: prefix + "PASSWORD", Value: password},
		{Key: prefix + nameKey, Value: name},
	}
}

func cacheGroups(r refs, family string, hasPassword bool) []*Group {
	host, port := r.of(base.AppSystemEnvVarHost), r.of(base.AppSystemEnvVarPort)
	password := r.of(base.AppSystemEnvVarPassword)

	switch family {
	case FamilyRedis:
		// A cache with authentication off gets a URL without credentials: some
		// clients refuse an empty password in one.
		auth := ""
		if hasPassword {
			auth = ":" + r.of(base.AppSystemEnvVarPasswordURLEncoded) + "@"
		}
		return []*Group{
			urlGroup("REDIS_URL", "redis://"+auth+host+":"+port+"/0"),
			individualGroup(false, &Var{Key: "REDIS_HOST", Value: host}, &Var{Key: "REDIS_PORT", Value: port},
				&Var{Key: "REDIS_PASSWORD", Value: password}),
		}
	case FamilyMemcached:
		return []*Group{{
			ID: "servers", Title: "Servers", Recommended: true,
			Description: "The address list memcached clients take.",
			Vars:        []*Var{{Key: "MEMCACHED_SERVERS", Value: host + ":" + port}},
		}}
	}
	return []*Group{individualGroup(true, &Var{Key: "CACHE_HOST", Value: host},
		&Var{Key: "CACHE_PORT", Value: port}, &Var{Key: "CACHE_PASSWORD", Value: password})}
}

func storageGroup(r refs) *Group {
	return &Group{
		ID: "s3", Title: "S3 client", Recommended: true,
		Description: "What S3 SDKs read: the endpoint inside the cluster, the key pair, the bucket and the region.",
		Vars: []*Var{
			{Key: "S3_ENDPOINT", Value: "http://" + r.of(base.AppSystemEnvVarHost) + ":" + r.of(base.AppSystemEnvVarPort)},
			{Key: "AWS_ACCESS_KEY_ID", Value: r.of(base.AppSystemEnvVarKeyID)},
			{Key: "AWS_SECRET_ACCESS_KEY", Value: r.of(base.AppSystemEnvVarSecret)},
			{Key: "S3_BUCKET", Value: r.of(base.AppSystemEnvVarBucket)},
			{Key: "AWS_REGION", Value: r.of(base.AppSystemEnvVarRegion)},
		},
	}
}

func addressGroup(r refs, target *Target) *Group {
	name := envKeyOf(target.Key)
	return &Group{
		ID: "addresses", Title: "Addresses", Recommended: true,
		Description: "Where to reach the app: inside the cluster, and from outside.",
		Vars: []*Var{
			{Key: name + "_URL", Value: "http://" + r.of(base.AppSystemEnvVarHost) + ":" + r.of(base.AppSystemEnvVarPort),
				Description: "Inside the cluster"},
			{Key: name + "_PUBLIC_URL", Value: r.of(base.AppSystemEnvVarAppURL),
				Description: "From outside; empty while the app has no domain"},
		},
	}
}

func sharedGroup(r refs, target *Target, keys []string) *Group {
	if len(keys) == 0 {
		return nil
	}
	sorted := append([]string{}, keys...)
	sort.Strings(sorted)
	vars := make([]*Var, 0, len(sorted))
	for _, key := range sorted {
		vars = append(vars, &Var{Key: key, Value: r.of(key)})
	}
	return &Group{
		ID: "shared", Title: "Shared variables of " + target.Name,
		Description: "Variables " + target.Name + " declares shared.",
		Vars:        vars,
	}
}
