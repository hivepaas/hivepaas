package registryserviceimpl

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/datakey"
)

// useDataKey installs a key for the test. A registry auth holds an encrypted
// field, and writing one without a key configured is refused rather than silently
// stored empty - which is the right behavior, and means a test that writes one
// has to install a key the way start-up does.
func useDataKey(t *testing.T) {
	t.Helper()

	key, err := datakey.Generate()
	if err != nil {
		t.Fatalf("datakey.Generate: %v", err)
	}
	datakey.SetActive(key)
	t.Cleanup(func() { datakey.SetActive(nil) })
}

func TestGeneratedPasswordsAreLongAndDifferent(t *testing.T) {
	first, err := generatePassword()
	if err != nil {
		t.Fatalf("generatePassword: %v", err)
	}
	second, err := generatePassword()
	if err != nil {
		t.Fatalf("generatePassword: %v", err)
	}

	assert.GreaterOrEqual(t, len(first), 32)
	assert.NotEqual(t, first, second)
}

// zot checks passwords against bcrypt hashes. A line it cannot parse is a
// registry nobody can log in to, and nothing else would say so.
func TestHtpasswdLineIsBcryptAndVerifies(t *testing.T) {
	line, err := htpasswdLine(registryUsername, "s3cret-password")
	if err != nil {
		t.Fatalf("htpasswdLine: %v", err)
	}

	user, hash, found := strings.Cut(line, ":")
	assert.True(t, found)
	assert.Equal(t, "hivepaas", user)
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(hash), []byte("s3cret-password")))
}

// Rotation keeps the old line beside the new one, so a service that has not been
// redeployed can still pull. Both have to survive into the file.
func TestHtpasswdContentKeepsEveryLine(t *testing.T) {
	content := htpasswdContent("hivepaas:$2y$10$new", "hivepaas:$2y$10$old")

	assert.Equal(t, "hivepaas:$2y$10$new\nhivepaas:$2y$10$old\n", content)
	assert.Equal(t, "hivepaas:$2y$10$only\n", htpasswdContent("hivepaas:$2y$10$only", ""))
}

func TestNewRegistryAuthSettingIsGlobalAndReadable(t *testing.T) {
	useDataKey(t)

	now := time.Date(2026, 9, 21, 10, 0, 0, 0, time.UTC)

	setting, err := newRegistryAuthSetting("auth-1", "registry.example.com", "pw", now)
	if err != nil {
		t.Fatalf("newRegistryAuthSetting: %v", err)
	}

	assert.Equal(t, base.SettingTypeRegistryAuth, setting.Type)
	assert.Equal(t, base.ObjectScopeGlobal, setting.Scope)
	assert.True(t, setting.Inheritable, "every project has to be able to push to it")

	auth := setting.MustAsRegistryAuth()
	assert.Equal(t, "registry.example.com", auth.Address)
	assert.Equal(t, "hivepaas", auth.Username)
	assert.False(t, auth.Readonly, "HivePaaS pushes with this credential")

	plain, err := auth.Password.GetPlain()
	assert.NoError(t, err)
	assert.Equal(t, "pw", plain)
}

// A domain change has to reach the credential, or builds keep pushing to the old
// address with no sign of why.
func TestUpdateRegistryAuthSettingRewritesTheAddress(t *testing.T) {
	useDataKey(t)

	now := time.Date(2026, 9, 21, 10, 0, 0, 0, time.UTC)
	setting, err := newRegistryAuthSetting("auth-1", "old.example.com", "pw", now)
	if err != nil {
		t.Fatalf("newRegistryAuthSetting: %v", err)
	}

	later := now.Add(time.Hour)
	if err = updateRegistryAuthSetting(setting, "new.example.com", "pw2", later); err != nil {
		t.Fatalf("updateRegistryAuthSetting: %v", err)
	}

	auth := setting.MustAsRegistryAuth()
	assert.Equal(t, "new.example.com", auth.Address)
	plain, _ := auth.Password.GetPlain()
	assert.Equal(t, "pw2", plain)
	assert.Equal(t, later, setting.UpdatedAt)
}

// An ordinary save is not a rotation: it must not quietly change the password
// every app is already using.
func TestUpdateRegistryAuthSettingKeepsThePasswordWhenNoneIsGiven(t *testing.T) {
	useDataKey(t)

	now := time.Date(2026, 9, 21, 10, 0, 0, 0, time.UTC)
	setting, err := newRegistryAuthSetting("auth-1", "old.example.com", "pw", now)
	if err != nil {
		t.Fatalf("newRegistryAuthSetting: %v", err)
	}

	if err = updateRegistryAuthSetting(setting, "new.example.com", "", now); err != nil {
		t.Fatalf("updateRegistryAuthSetting: %v", err)
	}

	plain, _ := setting.MustAsRegistryAuth().Password.GetPlain()
	assert.Equal(t, "pw", plain)
}
