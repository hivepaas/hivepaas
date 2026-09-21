package registryserviceimpl

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

const (
	// registryUsername is fixed. It is the namespace every image HivePaaS pushes
	// lands in - <domain>/hivepaas/<app key> - so it is not something to change
	// later without renaming every image in the registry.
	registryUsername = "hivepaas"

	// registryAuthName is what the credential is called in the settings list, so
	// that somebody who finds it there knows what created it.
	registryAuthName = "System registry"

	// passwordBytes is what is drawn from the random source; base64 makes it 32
	// characters, which is more than anything checking it will ever need.
	passwordBytes = 24
)

// generatePassword returns the registry account's password. It is generated
// rather than asked for: nobody types it, the dashboard shows it as a link to the
// registry auth setting, and every consumer reads it from there.
func generatePassword() (string, error) {
	buf := make([]byte, passwordBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", hperrors.Wrap(err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// htpasswdLine writes the line zot checks a password against. HivePaaS hashes it
// itself, so nothing asks an operator to run htpasswd by hand.
func htpasswdLine(user, password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	return fmt.Sprintf("%s:%s", user, hash), nil
}

// htpasswdContent joins the lines into the file. Empty lines are dropped, which
// is what makes "the previous password, if there is one" expressible without the
// caller branching.
func htpasswdContent(lines ...string) string {
	var out strings.Builder
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		out.WriteString(line)
		out.WriteString("\n")
	}
	return out.String()
}

// newRegistryAuthSetting builds the credential every push and pull uses.
//
// It is an ordinary registry-auth setting, global and inheritable, because that
// is what makes it appear in an app's deployment settings beside Docker Hub and
// anything else the operator added: nothing downstream needs to know that this
// one was created by HivePaaS.
func newRegistryAuthSetting(id, domain, password string, timeNow time.Time) (*entity.Setting, error) {
	setting := &entity.Setting{
		ID:          id,
		Scope:       base.ObjectScopeGlobal,
		Type:        base.SettingTypeRegistryAuth,
		Name:        registryAuthName,
		Status:      base.SettingStatusActive,
		Inheritable: true,
		Version:     entity.CurrentRegistryAuthVersion,
		UpdateVer:   1,
		CreatedAt:   timeNow,
		UpdatedAt:   timeNow,
	}
	err := setting.SetData(&entity.RegistryAuth{
		Username: registryUsername,
		Password: entity.NewEncryptedField(password),
		Address:  domain,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return setting, nil
}

// updateRegistryAuthSetting rewrites the address, and the password when one is
// given. An empty password keeps the stored one, which is what every save that is
// not a rotation does.
func updateRegistryAuthSetting(setting *entity.Setting, domain, password string, timeNow time.Time) error {
	auth, err := setting.AsRegistryAuth()
	if err != nil {
		return hperrors.Wrap(err)
	}

	auth.Address = domain
	auth.Username = registryUsername
	if password != "" {
		auth.Password = entity.NewEncryptedField(password)
	}
	if err = setting.SetData(auth); err != nil {
		return hperrors.Wrap(err)
	}
	setting.UpdateVer++
	setting.UpdatedAt = timeNow
	return nil
}
