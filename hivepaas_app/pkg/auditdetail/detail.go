// Package auditdetail builds the Detail payload of an audit entry.
//
// It exists because the rules for what may go into that payload are the same
// whatever is being recorded - a setting, a project, an app - while the code that
// works them out is easy to get subtly wrong, and getting it wrong writes a
// credential into an append-only table that outlives the object it came from.
// Solving it once, here, is the only way the answer stays the same across scopes.
//
// The safe shape is: name the fields that moved, and carry values only for fields
// the caller has decided are not secret. FieldChanges does the first, and is the
// one to reach for when the object holds anything a user typed.
package auditdetail

import (
	"encoding/json"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

// Reserved keys, so a caller's own field cannot displace the structured parts.
const (
	changesKey    = "changes"
	dataFieldsKey = "dataFields"
)

// Change is one before/after pair.
type Change struct {
	From any `json:"from"`
	To   any `json:"to"`
}

// Builder accumulates what an entry's Detail will say.
//
// The zero value is not usable; call New.
type Builder struct {
	fields     map[string]any
	changes    map[string]Change
	dataFields []string
}

func New() *Builder {
	return &Builder{
		fields:  map[string]any{},
		changes: map[string]Change{},
	}
}

// Set records a plain fact about the object - its type, its status now.
//
// For anything that came out of a user-supplied payload, prefer FieldChanges:
// this puts the value in the entry.
func (b *Builder) Set(key string, value any) *Builder {
	if key == "" || key == changesKey || key == dataFieldsKey {
		return b
	}
	if value == nil || value == "" {
		return b
	}
	b.fields[key] = scrub(value)
	return b
}

// Compare records that a field moved, with both values, and does nothing when it
// did not. Meant for the envelope around an object - name, status, expiry - not
// for its payload.
func (b *Builder) Compare(field string, from, to any) *Builder {
	if field == "" || from == to {
		return b
	}
	b.changes[field] = Change{From: scrub(from), To: scrub(to)}
	return b
}

// WithChangedFields records which payload fields moved, by name only.
//
// This is the safe way to describe a payload: see FieldChanges.
func (b *Builder) WithChangedFields(old, current any) *Builder {
	b.dataFields = FieldChanges(old, current)
	return b
}

// String renders the JSON to store.
//
// An unrenderable detail comes back empty rather than as an error: the detail is
// context around the entry, and losing it must never cost the entry itself.
func (b *Builder) String() string {
	out := make(map[string]any, len(b.fields)+2) //nolint:mnd
	for key, value := range b.fields {
		out[key] = value
	}
	if len(b.changes) > 0 {
		out[changesKey] = b.changes
	}
	if len(b.dataFields) > 0 {
		out[dataFieldsKey] = b.dataFields
	}
	if len(out) == 0 {
		return ""
	}

	encoded, err := json.Marshal(out)
	if err != nil {
		return ""
	}
	return string(encoded)
}

// scrub replaces a value that is plainly a stored secret with the placeholder.
//
// Every stored credential carries `hpenc:` or `hpsalt:`, so a value starting with
// one has come from somewhere it should not have. This is a backstop and nothing
// more: it cannot be the safety mechanism, because it only recognizes values that
// were declared as credentials, while the values most worth worrying about are
// the plain strings a user typed - a config file's contents, a script, an env var
// - which hold credentials constantly and are marked with nothing at all. Relying
// on this to decide what may be published would be a blocklist, and a blocklist
// here fails open. FieldChanges is the mechanism; this only catches a mistake.
func scrub(value any) any {
	text, ok := value.(string)
	if !ok {
		return value
	}
	for _, prefix := range base.AllEncryptionPrefixes {
		if strings.HasPrefix(text, prefix) {
			return base.MaskedSecret
		}
	}
	return value
}
