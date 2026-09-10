package auditdetail_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/auditdetail"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/datakey"
)

/// Fixtures - shaped like a real setting payload, not like the walker's cases.

type credentials struct {
	Username string                `json:"username"`
	Password entity.EncryptedField `json:"password"`
	Internal string                `json:"-"`
}

// apiKey is separate because HashField hashes on marshal, unset included, and
// picks a fresh salt each time - an empty one would read as changed everywhere.
type apiKey struct {
	Token entity.HashField `json:"token"`
}

type webhook struct {
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
	Retries []*retry          `json:"retries,omitempty"`
	Creds   *credentials      `json:"creds,omitempty"`
}

type retry struct {
	After string `json:"after"`
}

func withDataKey(t *testing.T) {
	t.Helper()
	key, err := datakey.Generate()
	if err != nil {
		t.Fatal(err)
	}
	datakey.SetActive(key)
}

// stored round-trips through JSON, which is how an object reaches the audit: read
// back from the database, not held over from the request that wrote it.
func stored[T any](t *testing.T, value *T) *T {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	out := new(T)
	if err := json.Unmarshal(encoded, out); err != nil {
		t.Fatal(err)
	}
	return out
}

func decode(t *testing.T, detail string) map[string]any {
	t.Helper()
	out := map[string]any{}
	if detail == "" {
		return out
	}
	if err := json.Unmarshal([]byte(detail), &out); err != nil {
		t.Fatalf("detail is not valid JSON: %v", err)
	}
	return out
}

/// FieldChanges

func TestFieldChangesNamesNestedPaths(t *testing.T) {
	withDataKey(t)
	old := stored(t, &webhook{
		URL:     "https://old",
		Headers: map[string]string{"X-Trace": "a"},
		Retries: []*retry{{After: "1s"}, {After: "5s"}},
		Creds:   &credentials{Username: "admin", Password: entity.NewEncryptedField("one")},
	})
	current := stored(t, &webhook{
		URL:     "https://old",
		Headers: map[string]string{"X-Trace": "b"},
		Retries: []*retry{{After: "1s"}, {After: "9s"}},
		Creds:   &credentials{Username: "admin", Password: entity.NewEncryptedField("two")},
	})

	got := strings.Join(auditdetail.FieldChanges(old, current), ",")
	want := "creds.password,headers.X-Trace,retries.1.after"
	if got != want {
		t.Fatalf("want %q, got %q", want, got)
	}
}

// The reason the walk is typed rather than a comparison of serialized JSON.
// Sealing the same plaintext again picks a fresh nonce, so the stored form
// differs while nothing changed.
func TestFieldChangesIgnoresResealedSecret(t *testing.T) {
	withDataKey(t)
	old := stored(t, &credentials{Username: "admin", Password: entity.NewEncryptedField("same")})
	current := stored(t, &credentials{Username: "admin", Password: entity.NewEncryptedField("same")})

	oldJSON, _ := json.Marshal(old)
	currentJSON, _ := json.Marshal(current)
	if string(oldJSON) == string(currentJSON) {
		t.Fatal("the ciphertext did not change, this test proves nothing")
	}
	if got := auditdetail.FieldChanges(old, current); len(got) != 0 {
		t.Fatalf("nothing changed, got %v", got)
	}
}

// A hash cannot be reversed and re-hashing picks a new salt, so an unchanged
// value reads as changed. Over-reporting is the intended way to be wrong.
func TestFieldChangesOverReportsHashedField(t *testing.T) {
	withDataKey(t)
	// HashField holds a hex token, which is what an API key is.
	old := stored(t, &apiKey{Token: entity.NewHashField("deadbeef")})
	current := stored(t, &apiKey{Token: entity.NewHashField("deadbeef")})

	if got := auditdetail.FieldChanges(old, current); len(got) != 1 || got[0] != "token" {
		t.Fatalf("want [token], got %v", got)
	}
}

func TestFieldChangesSkipsUnexportedSchema(t *testing.T) {
	withDataKey(t)
	old := &credentials{Username: "admin", Internal: "a"}
	current := &credentials{Username: "admin", Internal: "b"}

	// `json:"-"` is not part of the schema, so naming it would leak an
	// implementation detail rather than describe a change.
	if got := auditdetail.FieldChanges(old, current); len(got) != 0 {
		t.Fatalf("want nothing, got %v", got)
	}
}

func TestFieldChangesHandlesNilAndLengthChanges(t *testing.T) {
	withDataKey(t)
	if got := auditdetail.FieldChanges(nil, &credentials{}); got != nil {
		t.Errorf("want nil, got %v", got)
	}

	old := stored(t, &webhook{Retries: []*retry{{After: "1s"}}})
	current := stored(t, &webhook{Retries: []*retry{{After: "1s"}, {After: "2s"}}})
	if got := auditdetail.FieldChanges(old, current); len(got) != 1 || got[0] != "retries" {
		t.Errorf("want [retries], got %v", got)
	}

	before := stored(t, &webhook{Headers: map[string]string{"A": "1"}})
	after := stored(t, &webhook{Headers: map[string]string{"A": "1", "B": "2"}})
	if got := auditdetail.FieldChanges(before, after); len(got) != 1 || got[0] != "headers.B" {
		t.Errorf("want [headers.B], got %v", got)
	}
}

/// Builder

func TestBuilderRendersFieldsChangesAndPaths(t *testing.T) {
	withDataKey(t)
	old := stored(t, &credentials{Username: "admin", Password: entity.NewEncryptedField("one")})
	current := stored(t, &credentials{Username: "root", Password: entity.NewEncryptedField("two")})

	detail := decode(t, auditdetail.New().
		Set("objectType", "basic-auth").
		Set("empty", "").
		Compare("name", "web-auth", "web-auth-2").
		Compare("unchanged", "same", "same").
		WithChangedFields(old, current).
		String())

	if detail["objectType"] != "basic-auth" {
		t.Errorf("objectType = %v", detail["objectType"])
	}
	if _, ok := detail["empty"]; ok {
		t.Error("an empty value must not be rendered")
	}

	changes, _ := detail["changes"].(map[string]any)
	if _, ok := changes["unchanged"]; ok {
		t.Error("an unchanged field must not be listed")
	}
	if len(changes) != 1 {
		t.Errorf("changes = %v", changes)
	}

	fields, _ := detail["dataFields"].([]any)
	if len(fields) != 2 {
		t.Errorf("dataFields = %v", fields)
	}
}

func TestBuilderIsEmptyWhenNothingWasRecorded(t *testing.T) {
	if got := auditdetail.New().String(); got != "" {
		t.Fatalf("want empty, got %q", got)
	}
}

// Reserved keys must not be displaceable, or a caller could overwrite the
// structured parts with something of their own.
func TestBuilderRefusesReservedKeys(t *testing.T) {
	detail := decode(t, auditdetail.New().
		Set("changes", "mine").
		Set("dataFields", "mine").
		Compare("name", "a", "b").
		String())

	changes, ok := detail["changes"].(map[string]any)
	if !ok || len(changes) != 1 {
		t.Fatalf("changes was displaced: %v", detail["changes"])
	}
	if _, ok := detail["dataFields"]; ok {
		t.Error("dataFields was displaced")
	}
}

// The backstop. It is not the safety mechanism - FieldChanges is - but a stored
// secret reaching a value slot is a mistake worth catching rather than storing.
func TestBuilderMasksAStoredSecretThatReachesAValue(t *testing.T) {
	withDataKey(t)
	field := entity.NewEncryptedField("the-real-password")
	sealed, err := field.GetEncrypted()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(sealed, base.EncryptionKeyPrefix) {
		t.Fatalf("fixture is not a sealed value: %q", sealed)
	}

	rendered := auditdetail.New().
		Set("value", sealed).
		Compare("password", sealed, "hpsalt:whatever").
		String()

	if strings.Contains(rendered, sealed) {
		t.Fatal("a sealed value was stored verbatim")
	}
	if strings.Contains(rendered, "hpsalt:whatever") {
		t.Fatal("a salted value was stored verbatim")
	}
	if !strings.Contains(rendered, base.MaskedSecret) {
		t.Fatalf("want the placeholder, got %s", rendered)
	}
}
