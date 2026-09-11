package traefiksettingsuc

import (
	"context"
	"sort"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/auditdetail"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
	"github.com/hivepaas/hivepaas/services/traefik/traefikhelper"
)

// The sections traefik's two settings pages are recorded under.
//
// Prefixed, because these entries share their type with the HivePaaS settings
// next door, and those already use "service" for their own replica count. Two
// different pages under one section name would make the section filter - what a
// reader uses to tell one page from another - answer for both at once.
const (
	auditSectionConfigOptions   = "traefik-config"
	auditSectionServiceSettings = "traefik-service"
)

// auditResNames is what each section is called where the entry is read.
var auditResNames = map[string]string{
	auditSectionConfigOptions:   "traefik config options",
	auditSectionServiceSettings: "traefik service settings",
}

// recordTraefikSettingsUpdate records a change to traefik's own settings.
//
// Filed under the hivepaas scope rather than under the traefik app the setting
// row hangs off. Traefik is a real app - that is where its settings live and how
// its trials are keyed - but nobody reads its history as an app's: it is the
// install's ingress, and what these changes answer for is what was done to this
// install. The HivePaaS settings next door are filed the same way.
//
// It shares their audit type too. Traefik's command decides whether anything is
// reachable at all, which is the question the HivePaaS routing and service
// settings answer for the dashboard, and section already tells the pages apart -
// a type per page would put values in the type filter that section already holds.
func (uc *UC) recordTraefikSettingsUpdate(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	section string,
	detail *auditdetail.Builder,
) error {
	if detail == nil {
		detail = auditdetail.New()
	}

	err := auditservice.RecordAllowed(ctx, uc.auditService, db, &auditservice.Entry{
		Type:    base.AuditLogTypeHivePaaSSettingsUpdate,
		Scope:   base.ObjectScopeHivepaas,
		Source:  base.AuditLogSourceAPIUpdate,
		Section: section,
		Auth:    auth,
		ResType: base.ResourceTypeSetting,
		ResName: auditResNames[section],
		Detail:  detail.String(),
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

// changedCommandArgs names the traefik flags a command change touched, by key and
// never by value.
//
// Keys only, deliberately. The command line is free text an operator typed, and
// while most of it is switches, nothing stops a credential being written into one
// - certificate resolvers and plugins take them - and an audit row outlives the
// service it describes. The keys answer what this trail is actually asked: who
// turned the API on, who changed an entrypoint, who moved the log level.
//
// A change that only reorders the command lists no keys. Order does matter to
// traefik, which takes the last value for a repeated key, but what the entry has
// to say about such a change is that it happened - and the entry says that with
// or without a field list.
func changedCommandArgs(before, after []string) []string {
	beforeByKey := commandArgsByKey(before)
	afterByKey := commandArgsByKey(after)

	changed := make([]string, 0, len(afterByKey))
	for key, arg := range afterByKey {
		if beforeByKey[key] != arg {
			changed = append(changed, key)
		}
	}
	for key := range beforeByKey {
		if _, ok := afterByKey[key]; !ok {
			changed = append(changed, key)
		}
	}
	sort.Strings(changed)
	return changed
}

// commandArgsByKey indexes a command line by flag, keeping the occurrence that
// wins.
//
// The last one, because that is the one traefik runs under. The binary name and
// anything unparseable are dropped: they have no key to report a change under,
// and neither is settable from these endpoints anyway.
func commandArgsByKey(args []string) map[string]string {
	byKey := make(map[string]string, len(args))
	for _, arg := range args {
		arg = strings.TrimSpace(arg)
		key, _, valid := traefikhelper.ParseCommandArg(arg)
		if !valid {
			continue
		}
		byKey[key] = arg
	}
	return byKey
}
