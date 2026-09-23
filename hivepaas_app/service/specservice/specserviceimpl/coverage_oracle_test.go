package specserviceimpl

import (
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
)

// The exporter owns its own types rather than reusing the dashboard DTOs, which
// gives up the guarantee those provided for free: the Get and Update DTOs carry
// identical field sets, so anything readable was writable.
//
// These tests recover it. A test may import anything - a test is not a layer -
// so the Update request DTOs are used as an oracle, in both directions. A spec
// field with no writable counterpart means export produces something import can
// never write back. A writable field the spec omits is a coverage gap, and is
// what replaces "a field added to a DTO propagates automatically": instead of
// appearing in specs silently, it fails here with its name.
//
// The comparison is one level deep: it covers the top-level field of each block,
// so a whole block going missing is caught, but a field dropped from inside
// Capabilities or VolumeOptions is not. Recursing would mean matching structures
// that legitimately differ in shape - Storage.Mounts is a map here and a slice
// there - and the honest limit is worth more than a check that looks deeper than
// it is. Nested fields are covered by the round-trip test instead.

// fieldNames returns the tag names of a struct's exported fields, flattening
// embedded structs the way an encoder does.
func fieldNames(t *testing.T, v any, tag string) []string {
	t.Helper()
	typ := reflect.TypeOf(v)
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}

	var names []string
	for i := range typ.NumField() {
		field := typ.Field(i)
		if !field.IsExported() {
			continue
		}
		name, _, _ := strings.Cut(field.Tag.Get(tag), ",")

		if field.Anonymous && name == "" {
			embedded := field.Type
			for embedded.Kind() == reflect.Pointer {
				embedded = embedded.Elem()
			}
			if embedded.Kind() == reflect.Struct {
				names = append(names, fieldNames(t, reflect.New(embedded).Interface(), tag)...)
				continue
			}
		}
		if name == "" || name == "-" {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func contains(list []string, want string) bool {
	for _, item := range list {
		if item == want {
			return true
		}
	}
	return false
}

// specAliases records the places a spec field is deliberately named differently
// from its DTO counterpart, mapping spec name -> DTO name. Each one is a
// readability choice inside a block that already names the subject.
var specAliases = map[string]string{
	"attachments": "networkAttachments", // inside a networks: block
	// Inside a storage: block. The screen edits every mount in one list; a spec
	// keeps the ones HivePaaS manages apart from the rest, so that only those
	// travel in the form import can build again anywhere.
	"dockerMounts": "mounts",
}

type oracleCase struct {
	name string
	spec any
	// update is the Update request DTO that decides what is writable.
	update any
	// omitted are writable fields the spec deliberately does not carry.
	omitted []string
}

func oracleCases() []oracleCase {
	return []oracleCase{
		{
			name:    "resources",
			spec:    specmodel.Resources{},
			update:  appsettingsdto.UpdateAppResourceSettingsReq{},
			omitted: []string{"updateVer"},
		},
		{
			name:   "storage",
			spec:   specmodel.Storage{},
			update: appsettingsdto.UpdateAppStorageSettingsReq{},
			// resetStorage is a command flag, not configuration: it clears what
			// the mounts being added land on. There is nothing for a spec to
			// carry, and a spec that carried it would delete data on import.
			omitted: []string{"updateVer", "resetStorage"},
		},
		{
			name:    "networks",
			spec:    specmodel.Networks{},
			update:  appsettingsdto.UpdateAppNetworkSettingsReq{},
			omitted: []string{"updateVer"},
		},
		{
			name:    "service",
			spec:    specmodel.Service{},
			update:  appsettingsdto.UpdateAppServiceSettingsReq{},
			omitted: []string{"updateVer"},
		},
		{
			name:   "container",
			spec:   specmodel.Container{},
			update: appsettingsdto.UpdateAppContainerSettingsReq{},
			// command and workingDir live in Source, which carries the value the
			// setting and the service agree on rather than duplicating it.
			omitted: []string{"updateVer", "command", "workingDir"},
		},
	}
}

func TestEverySpecFieldIsWritable(t *testing.T) {
	for _, tc := range oracleCases() {
		t.Run(tc.name, func(t *testing.T) {
			writable := fieldNames(t, tc.update, "json")
			for _, name := range fieldNames(t, tc.spec, "yaml") {
				if alias, ok := specAliases[name]; ok {
					name = alias
				}
				assert.True(t, contains(writable, name),
					"spec exports %q but no Update request accepts it", name)
			}
		})
	}
}

func TestNoWritableFieldIsMissingFromTheSpec(t *testing.T) {
	for _, tc := range oracleCases() {
		t.Run(tc.name, func(t *testing.T) {
			specNames := fieldNames(t, tc.spec, "yaml")
			for i, name := range specNames {
				if alias, ok := specAliases[name]; ok {
					specNames[i] = alias
				}
			}

			var gaps []string
			for _, name := range fieldNames(t, tc.update, "json") {
				if !contains(specNames, name) && !contains(tc.omitted, name) {
					gaps = append(gaps, name)
				}
			}
			assert.Empty(t, gaps,
				"these configurable fields are missing from the spec: %v", gaps)
		})
	}
}
