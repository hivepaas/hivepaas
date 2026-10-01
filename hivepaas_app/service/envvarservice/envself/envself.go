// Package envself suggests the variables an engine's image reads to set itself
// up - its user, its password, its first database - each a reference to what
// the app's own kind settings publish, such as ${HIVEPAAS_PASSWORD}. The
// credentials then live in one place, the App Kind screen, and the apps that
// link to this one are told the same ones. Like envlink, it holds no values.
//
// The variables are those of each engine's official image, as the app store's
// templates set them.
package envself

import (
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

// Engine is an engine whose image the suggestions are for.
type Engine struct {
	ID       string
	Title    string
	Category base.AppCategory
	// Image is the official image the variables are those of.
	Image string
	// InitOnly says the image reads the variables only when it creates its
	// data, and keeps its own copy of the credentials afterwards.
	InitOnly bool
	vars     []engineVar
	// command is what to run the image with, for one that reads no variable.
	command string
}

// engineVar is a variable an image reads, and what it is set to.
type engineVar struct {
	key   string
	value string
	// needs is the system variable the value refers to, if any.
	needs       string
	description string
}

func ref(name string) string { return "${" + name + "}" }

func fromKind(key, system, description string) engineVar {
	return engineVar{key: key, value: ref(system), needs: system, description: description}
}

const enginePostgres = "postgres"

const (
	descDB       = "The database created on the first start"
	descUser     = "The user created on the first start"
	descPassword = "The user's password" //nolint:gosec // G101: a description
	descRoot     = "The root user's password"
)

// redisCommand starts Redis or Valkey with the App Kind's password, memory
// ceiling, eviction and persistence, as the app store's templates do: the
// password goes through the environment, not the command, which is stored in
// plain text and which ps shows every process.
func redisCommand(server string) string {
	return `sh -c 'case "$HIVEPAAS_PERSISTENCE_MODE" in ` +
		`none) set -- --appendonly no --save "";; ` +
		`rdb) set -- --appendonly no --save 3600 1 --save 300 100 --save 60 10000;; ` +
		`rdb+aof) set -- --appendonly yes --save 3600 1 --save 300 100 --save 60 10000;; ` +
		`*) set -- --appendonly yes --save "";; ` +
		`esac; ` +
		`exec ` + server + ` --requirepass "$HIVEPAAS_PASSWORD" ` +
		`--maxmemory "${HIVEPAAS_MAX_MEMORY:-0}" ` +
		`--maxmemory-policy "${HIVEPAAS_EVICTION_RULE:-noeviction}" --dir /data "$@"'`
}

// Engines are the engines suggestions are made for, in the order offered.
var Engines = []*Engine{
	{ID: enginePostgres, Title: "PostgreSQL", Category: base.AppCategoryDatabase, Image: enginePostgres, InitOnly: true,
		vars: []engineVar{
			fromKind("POSTGRES_DB", base.AppSystemEnvVarDatabaseName, descDB),
			fromKind("POSTGRES_USER", base.AppSystemEnvVarUser, descUser+", a superuser"),
			fromKind("POSTGRES_PASSWORD", base.AppSystemEnvVarPassword, descPassword),
		}},
	{ID: "mysql", Title: "MySQL", Category: base.AppCategoryDatabase, Image: "mysql", InitOnly: true,
		vars: []engineVar{
			fromKind("MYSQL_DATABASE", base.AppSystemEnvVarDatabaseName, descDB),
			fromKind("MYSQL_USER", base.AppSystemEnvVarUser, descUser),
			fromKind("MYSQL_PASSWORD", base.AppSystemEnvVarPassword, descPassword),
			fromKind("MYSQL_ROOT_PASSWORD", base.AppSystemEnvVarRootPassword, descRoot),
		}},
	{ID: "mariadb", Title: "MariaDB", Category: base.AppCategoryDatabase, Image: "mariadb", InitOnly: true,
		vars: []engineVar{
			fromKind("MARIADB_DATABASE", base.AppSystemEnvVarDatabaseName, descDB),
			fromKind("MARIADB_USER", base.AppSystemEnvVarUser, descUser),
			fromKind("MARIADB_PASSWORD", base.AppSystemEnvVarPassword, descPassword),
			fromKind("MARIADB_ROOT_PASSWORD", base.AppSystemEnvVarRootPassword, descRoot),
		}},
	{ID: "mongodb", Title: "MongoDB", Category: base.AppCategoryDatabase, Image: "mongo", InitOnly: true,
		vars: []engineVar{
			fromKind("MONGO_INITDB_ROOT_USERNAME", base.AppSystemEnvVarUser, "The root user created on the first start"),
			fromKind("MONGO_INITDB_ROOT_PASSWORD", base.AppSystemEnvVarPassword, "The root user's password"),
			fromKind("MONGO_INITDB_DATABASE", base.AppSystemEnvVarDatabaseName,
				"The database the first start's scripts run against"),
		}},
	{ID: "clickhouse", Title: "ClickHouse", Category: base.AppCategoryDatabase, Image: "clickhouse", InitOnly: true,
		vars: []engineVar{
			fromKind("CLICKHOUSE_DB", base.AppSystemEnvVarDatabaseName, descDB),
			fromKind("CLICKHOUSE_USER", base.AppSystemEnvVarUser, descUser),
			fromKind("CLICKHOUSE_PASSWORD", base.AppSystemEnvVarPassword, descPassword),
			{key: "CLICKHOUSE_DEFAULT_ACCESS_MANAGEMENT", value: "1",
				description: "Lets the user create users and grant rights"},
		}},
	{ID: "rabbitmq", Title: "RabbitMQ", Category: base.AppCategoryDatabase, Image: "rabbitmq", InitOnly: true,
		vars: []engineVar{
			fromKind("RABBITMQ_DEFAULT_USER", base.AppSystemEnvVarUser, descUser),
			fromKind("RABBITMQ_DEFAULT_PASS", base.AppSystemEnvVarPassword, descPassword),
			fromKind("RABBITMQ_DEFAULT_VHOST", base.AppSystemEnvVarDatabaseName,
				"The virtual host created on the first start"),
		}},
	{ID: "redis", Title: "Redis", Category: base.AppCategoryCache, Image: "redis",
		command: redisCommand("redis-server")},
	{ID: "valkey", Title: "Valkey", Category: base.AppCategoryCache, Image: "valkey/valkey",
		command: redisCommand("valkey-server")},
}

// aliases are the engines App Kind may name that run one of Engines' images,
// or one built on it.
var aliases = map[string]string{
	"postgresql": enginePostgres, "timescaledb": enginePostgres, "postgis": enginePostgres,
	"pgvector": enginePostgres, "vectorchord": enginePostgres, "mongo": "mongodb",
}

// EngineOf is the engine of Engines an App Kind engine is, or nil.
func EngineOf(kindEngine string) *Engine {
	id := strings.ToLower(strings.TrimSpace(kindEngine))
	if alias, ok := aliases[id]; ok {
		id = alias
	}
	for _, e := range Engines {
		if e.ID == id {
			return e
		}
	}
	return nil
}

// Facts are what the app's kind settings give it to refer to.
type Facts struct {
	Category base.AppCategory
	Engine   string
	// Set are the system variables the kind settings give a value.
	Set map[string]bool
}

// Var is one suggested variable.
type Var struct {
	Key         string
	Value       string
	Description string
}

// Suggestion is what to set for an engine's image.
type Suggestion struct {
	Engine   *Engine
	Vars     []*Var
	Command  string
	Warnings []string
}

// Suggest is what the engine's image needs, and what is missing for it.
func Suggest(engine *Engine, facts *Facts) *Suggestion {
	s := &Suggestion{Engine: engine, Command: engine.command}
	for _, v := range engine.vars {
		s.Vars = append(s.Vars, &Var{Key: v.key, Value: v.value, Description: v.description})
	}

	if facts.Category != engine.Category {
		category := string(facts.Category)
		if category == "" {
			category = "not set"
		}
		s.Warnings = append(s.Warnings, "The values refer to the credentials App Kind gives this app, and its "+
			"category is "+category+": set it to "+string(engine.Category)+", with the credentials, or they "+
			"come out empty.")
		return s
	}
	if kindEngine := EngineOf(facts.Engine); kindEngine != engine {
		s.Warnings = append(s.Warnings, "App Kind names the engine "+orNone(facts.Engine)+", not "+engine.Title+
			": set it there too, so that the apps linking to this one are told how to connect.")
	}
	needs := map[string]bool{}
	for _, v := range engine.vars {
		if v.needs != "" {
			needs[v.needs] = true
		}
	}
	if engine.command != "" {
		needs[base.AppSystemEnvVarPassword] = true
	}
	for _, missing := range []struct{ name, noun string }{
		{base.AppSystemEnvVarDatabaseName, "database name"},
		{base.AppSystemEnvVarUser, "user"},
		{base.AppSystemEnvVarPassword, "password"},
		{base.AppSystemEnvVarRootPassword, "root password"},
	} {
		if needs[missing.name] && !facts.Set[missing.name] {
			s.Warnings = append(s.Warnings, "App Kind sets no "+missing.noun+": "+missing.name+
				" is empty until it does.")
		}
	}
	return s
}

func orNone(engine string) string {
	if strings.TrimSpace(engine) == "" {
		return "none"
	}
	return engine
}
