package entity

import (
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
)

const (
	CurrentSSHKeyVersion = 1
)

var _ = registerSettingParser(base.SettingTypeSSHKey, &sshKeyParser{})

type sshKeyParser struct {
}

func (s *sshKeyParser) New() SettingData {
	return &SSHKey{}
}

type SSHKey struct {
	KeyType    base.PrivateKeyType `json:"keyType"`
	PublicKey  string              `json:"publicKey,omitempty"`
	PrivateKey EncryptedField      `json:"privateKey"`
	Passphrase EncryptedField      `json:"passphrase,omitzero"`
}

func (s *SSHKey) GetType() base.SettingType {
	return base.SettingTypeSSHKey
}

func (s *SSHKey) GetRefObjectIDs() *RefObjectIDs {
	return &RefObjectIDs{}
}

func (s *SSHKey) MustDecrypt() *SSHKey {
	s.PrivateKey.MustGetPlain()
	s.Passphrase.MustGetPlain()
	return s
}

func (s *SSHKey) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentSSHKeyVersion {
		return false, nil
	}
	if setting.Version > CurrentSSHKeyVersion {
		return false, apperrors.New(apperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentSSHKeyVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}

func (s *Setting) AsSSHKey() (*SSHKey, error) {
	return parseSettingAs[*SSHKey](s)
}

func (s *Setting) MustAsSSHKey() *SSHKey {
	return gofn.Must(s.AsSSHKey())
}
