package auditdetail

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

var (
	encryptedFieldType = reflect.TypeFor[entity.EncryptedField]()
	hashFieldType      = reflect.TypeFor[entity.HashField]()
)

// FieldChanges lists which fields of an object moved, as JSON paths, and never
// what they moved to or from.
//
// Naming the fields is most of what a payload has to say - "somebody changed the
// smtp password" is the question asked of an audit trail - and it can be said
// without a value ever leaving this function. Paths are the object's JSON schema,
// which is already public in the API spec, so they disclose nothing the reader
// was not entitled to.
//
// The walk is typed rather than textual, for two reasons.
//
// It has to be typed to be correct: EncryptedField and HashField do not compare
// as strings. A field sealed again from the same plaintext gets a fresh nonce and
// a different ciphertext, so comparing serialized JSON would report a password
// change on every save that merely echoed the password back.
//
// It has to be typed to be safe. Redacting whatever carries the `hpenc:` and
// `hpsalt:` markers covers every declared credential field - and the declarations
// are thorough, down to Slack webhook URLs - but it publishes everything it does
// not recognize. A config file's contents, a script body, a command template's
// env var and argument values are all plain strings holding whatever a user
// typed, which routinely includes an API key or a private key, and none of it
// carries a marker. That is a blocklist, and a blocklist here fails open into an
// append-only table.
func FieldChanges(old, current any) []string {
	if old == nil || current == nil {
		return nil
	}
	walker := &fieldWalker{}
	walker.compare("", reflect.ValueOf(old), reflect.ValueOf(current))
	sort.Strings(walker.changed)
	return gofn.ToSet(walker.changed)
}

// fieldWalker collects the paths that differ. It holds only paths: no value it
// reads is ever stored on it.
type fieldWalker struct {
	changed []string
}

func (w *fieldWalker) mark(path string) {
	if path == "" {
		return
	}
	w.changed = append(w.changed, path)
}

//nolint:exhaustive
func (w *fieldWalker) compare(path string, old, current reflect.Value) {
	old, current = deref(old), deref(current)

	switch {
	case !old.IsValid() && !current.IsValid():
		return
	case !old.IsValid() || !current.IsValid(), old.Type() != current.Type():
		w.mark(path)
		return
	}

	switch old.Type() {
	case encryptedFieldType:
		w.compareEncrypted(path, old, current)
		return
	case hashFieldType:
		w.compareHashed(path, old, current)
		return
	}

	switch old.Kind() {
	case reflect.Struct:
		w.compareStruct(path, old, current)
	case reflect.Slice, reflect.Array:
		w.compareList(path, old, current)
	case reflect.Map:
		w.compareMap(path, old, current)
	default:
		if !old.CanInterface() || !current.CanInterface() {
			return
		}
		if !reflect.DeepEqual(old.Interface(), current.Interface()) {
			w.mark(path)
		}
	}
}

func (w *fieldWalker) compareEncrypted(path string, old, current reflect.Value) {
	oldField, ok := addrAs[entity.EncryptedField](old)
	newField, ok2 := addrAs[entity.EncryptedField](current)
	if !ok || !ok2 {
		w.mark(path)
		return
	}

	equal, err := oldField.Equal(newField)
	if err != nil {
		// Undecryptable is not a reason to stay quiet about a field that may have
		// moved. Equal compares plaintext in memory; neither side is kept.
		w.mark(path)
		return
	}
	if !equal {
		w.mark(path)
	}
}

// compareHashed over-reports on purpose.
//
// A hash cannot be reversed, and hashing the same secret again picks a new salt,
// so a value that was re-submitted unchanged reads as changed. That is the right
// way to be wrong: a spurious "the secret was set" costs a reader a moment, a
// missed one costs the trail.
func (w *fieldWalker) compareHashed(path string, old, current reflect.Value) {
	oldField, ok := addrAs[entity.HashField](old)
	newField, ok2 := addrAs[entity.HashField](current)
	if !ok || !ok2 {
		w.mark(path)
		return
	}
	if oldField.String() != newField.String() {
		w.mark(path)
	}
}

func (w *fieldWalker) compareStruct(path string, old, current reflect.Value) {
	typ := old.Type()
	for i := range typ.NumField() {
		field := typ.Field(i)
		if !field.IsExported() {
			continue
		}
		name := jsonFieldName(field)
		if name == "" {
			continue
		}
		w.compare(joinPath(path, name), old.Field(i), current.Field(i))
	}
}

func (w *fieldWalker) compareList(path string, old, current reflect.Value) {
	if old.Len() != current.Len() {
		w.mark(path)
		return
	}
	for i := range old.Len() {
		w.compare(joinPath(path, strconv.Itoa(i)), old.Index(i), current.Index(i))
	}
}

// compareMap reports the keys that moved. Map keys in these types are schema -
// header names and the like - not user secrets, so naming one is the same
// disclosure as naming a struct field.
func (w *fieldWalker) compareMap(path string, old, current reflect.Value) {
	seen := map[string]bool{}
	for _, key := range old.MapKeys() {
		name := fmt.Sprintf("%v", key.Interface())
		seen[name] = true
		w.compare(joinPath(path, name), old.MapIndex(key), current.MapIndex(key))
	}
	for _, key := range current.MapKeys() {
		name := fmt.Sprintf("%v", key.Interface())
		if !seen[name] {
			w.mark(joinPath(path, name))
		}
	}
}

// addrAs takes the addressable form of a struct value. A value reached through an
// unexported field or a map is not addressable, and the pointer methods these
// field types carry cannot be called on it.
func addrAs[T any](value reflect.Value) (*T, bool) {
	if !value.CanAddr() || !value.Addr().CanInterface() {
		return nil, false
	}
	typed, ok := value.Addr().Interface().(*T)
	return typed, ok
}

func deref(value reflect.Value) reflect.Value {
	for value.IsValid() && (value.Kind() == reflect.Pointer || value.Kind() == reflect.Interface) {
		if value.IsNil() {
			return reflect.Value{}
		}
		value = value.Elem()
	}
	return value
}

// jsonFieldName is the name the field is known by outside Go. A field marked
// `json:"-"` is skipped: it is not part of the schema, so naming it would leak an
// implementation detail rather than describe a change.
func jsonFieldName(field reflect.StructField) string {
	tag := field.Tag.Get("json")
	if tag == "-" {
		return ""
	}
	name, _, _ := strings.Cut(tag, ",")
	if name == "" {
		return field.Name
	}
	return name
}

func joinPath(prefix, name string) string {
	if prefix == "" {
		return name
	}
	return prefix + "." + name
}
